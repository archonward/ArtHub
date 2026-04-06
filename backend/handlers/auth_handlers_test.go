package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignupRejectsShortPassword(t *testing.T) {
	setupTestDB(t)

	recorder := executeJSONRequest(Signup, http.MethodPost, "/auth/signup", map[string]string{
		"username": "arthur",
		"password": "short",
	})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "validation_error" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestSignupCreatesSessionCookie(t *testing.T) {
	setupTestDB(t)

	recorder := executeJSONRequest(Signup, http.MethodPost, "/auth/signup", map[string]string{
		"username": "arthur",
		"password": "verysecurepassword",
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	user := decodeSuccessEnvelope[User](t, recorder)
	if user.Username != "arthur" {
		t.Fatalf("unexpected user payload: %+v", user)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != sessionCookieName {
		t.Fatalf("expected session cookie to be set")
	}

	var sessionCount int
	if err := db().QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&sessionCount); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionCount != 1 {
		t.Fatalf("expected one session, got %d", sessionCount)
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	setupTestDB(t)

	passwordHash, err := hashPassword("correcthorsebattery")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO users (username, password_hash)
		VALUES ('arthur', ?)
	`, passwordHash); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	recorder := executeJSONRequest(Login, http.MethodPost, "/auth/login", map[string]string{
		"username": "arthur",
		"password": "wrongpassword",
	})

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "invalid_credentials" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestLoginSetsSessionCookie(t *testing.T) {
	setupTestDB(t)

	passwordHash, err := hashPassword("correcthorsebattery")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO users (username, password_hash)
		VALUES ('arthur', ?)
	`, passwordHash); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	recorder := executeJSONRequest(Login, http.MethodPost, "/auth/login", map[string]string{
		"username": "arthur",
		"password": "correcthorsebattery",
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != sessionCookieName {
		t.Fatalf("expected auth cookie to be set")
	}
}

func TestLogoutClearsSession(t *testing.T) {
	setupTestDB(t)

	signupRecorder := executeJSONRequest(Signup, http.MethodPost, "/auth/signup", map[string]string{
		"username": "arthur",
		"password": "verysecurepassword",
	})
	sessionCookie := signupRecorder.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(sessionCookie)
	recorder := httptest.NewRecorder()

	Logout(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var count int
	if err := db().QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected session to be deleted")
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 || cookies[0].MaxAge != -1 {
		t.Fatalf("expected cleared cookie")
	}
}

func TestCurrentSessionUserReturnsAuthenticatedUser(t *testing.T) {
	setupTestDB(t)

	signupRecorder := executeJSONRequest(Signup, http.MethodPost, "/auth/signup", map[string]string{
		"username": "arthur",
		"password": "verysecurepassword",
	})
	sessionCookie := signupRecorder.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(sessionCookie)
	recorder := httptest.NewRecorder()

	CurrentSessionUser(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	user := decodeSuccessEnvelope[User](t, recorder)
	if user.Username != "arthur" {
		t.Fatalf("unexpected user payload: %+v", user)
	}
}

func TestCurrentSessionUserRejectsMissingSession(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	recorder := httptest.NewRecorder()

	CurrentSessionUser(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "not_authenticated" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestSignupRejectsDuplicateUsername(t *testing.T) {
	setupTestDB(t)

	passwordHash, err := hashPassword("verysecurepassword")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if _, err := db().Exec(`INSERT INTO users (username, password_hash) VALUES ('legacy', ?)`, passwordHash); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	recorder := executeJSONRequest(Signup, http.MethodPost, "/auth/signup", map[string]string{
		"username": "legacy",
		"password": "verysecurepassword",
	})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}
}

func TestResolveAuthenticatedUserReturnsNilForExpiredSession(t *testing.T) {
	setupTestDB(t)

	if _, err := db().Exec(`
		INSERT INTO users (username, password_hash)
		VALUES ('arthur', 'placeholder')
	`); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	token := "expired-token"
	tokenHash := hashSessionToken(token)
	if _, err := db().Exec(`
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES (1, ?, datetime('now', '-1 day'))
	`, tokenHash); err != nil {
		t.Fatalf("insert expired session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", bytes.NewReader(nil))
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})

	user, err := resolveAuthenticatedUser(req)
	if err != nil {
		t.Fatalf("resolve user: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user for expired session")
	}
}
