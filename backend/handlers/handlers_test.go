package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/archonward/CampusCommons/backend/database"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if _, err := testDB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := database.CreateSchema(testDB); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	database.DB = testDB
	t.Cleanup(func() {
		_ = testDB.Close()
	})
}

func executeJSONRequest(handler http.HandlerFunc, method, path string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler(recorder, req)
	return recorder
}

func decodeSuccessEnvelope[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()

	var envelope ResponseEnvelope[T]
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode success envelope: %v", err)
	}
	return envelope.Data
}

func decodeErrorEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) ErrorResponseEnvelope {
	t.Helper()

	var envelope ErrorResponseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	return envelope
}

func signupAndSessionCookie(t *testing.T, username string) (*User, *http.Cookie) {
	t.Helper()

	recorder := executeJSONRequest(Signup, http.MethodPost, "/auth/signup", map[string]string{
		"username": username,
		"password": "verysecurepassword",
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected signup 201, got %d", recorder.Code)
	}

	user := decodeSuccessEnvelope[User](t, recorder)
	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}

	return &user, cookies[0]
}

func TestGetTopicByIDReturnsTopic(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	result, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID)
	if err != nil {
		t.Fatalf("insert topic: %v", err)
	}
	topicID, _ := result.LastInsertId()

	req := httptest.NewRequest(http.MethodGet, "/topics/1", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()
	GetTopicByID(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	topic := decodeSuccessEnvelope[Topic](t, recorder)
	if int64(topic.ID) != topicID || topic.Title != "Markets" {
		t.Fatalf("unexpected topic payload: %+v", topic)
	}
}

func TestCreateTopicRequiresAuthentication(t *testing.T) {
	setupTestDB(t)

	recorder := executeJSONRequest(CreateTopic, http.MethodPost, "/topics", map[string]any{
		"title":       "Markets",
		"description": "desc",
	})

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "not_authenticated" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestCreateTopicUsesAuthenticatedUserInsteadOfCreatedByPayload(t *testing.T) {
	setupTestDB(t)
	user, cookie := signupAndSessionCookie(t, "owner")

	payload, _ := json.Marshal(map[string]any{
		"title":       "Markets",
		"description": "desc",
		"created_by":  999,
	})

	req := httptest.NewRequest(http.MethodPost, "/topics", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	recorder := httptest.NewRecorder()

	CreateTopic(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	topic := decodeSuccessEnvelope[Topic](t, recorder)
	if topic.CreatedBy != user.ID {
		t.Fatalf("expected created_by=%d, got %+v", user.ID, topic)
	}
}

func TestCreateTopicRejectsXUserIDBypassWithoutSession(t *testing.T) {
	setupTestDB(t)

	payload, _ := json.Marshal(map[string]any{
		"title":       "Markets",
		"description": "desc",
		"created_by":  1,
	})

	req := httptest.NewRequest(http.MethodPost, "/topics", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	recorder := httptest.NewRecorder()

	CreateTopic(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "not_authenticated" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestUpdateTopicRejectsNonOwnerEvenWithSpoofedXUserID(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	otherUser, otherCookie := signupAndSessionCookie(t, "other")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"title":       "Updated",
		"description": "Updated description",
	})

	req := httptest.NewRequest(http.MethodPut, "/topics/1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	req.AddCookie(otherCookie)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	_ = otherUser
	UpdateTopic(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "forbidden" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestUpdateTopicSucceedsForOwnerSession(t *testing.T) {
	setupTestDB(t)
	owner, ownerCookie := signupAndSessionCookie(t, "owner")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"title":       "Updated",
		"description": "Updated description",
	})

	req := httptest.NewRequest(http.MethodPut, "/topics/1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(ownerCookie)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	UpdateTopic(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	topic := decodeSuccessEnvelope[Topic](t, recorder)
	if topic.Title != "Updated" {
		t.Fatalf("expected updated topic, got %+v", topic)
	}
}

func TestDeletePostRequiresAuthentication(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/posts/1", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	DeletePost(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestCreateCommentUsesAuthenticatedUserInsteadOfCreatedByPayload(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	commenter, commentCookie := signupAndSessionCookie(t, "commenter")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	payload, _ := json.Marshal(map[string]any{
		"body":       "Comment body",
		"created_by": 999,
	})

	req := httptest.NewRequest(http.MethodPost, "/posts/1/comments", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(commentCookie)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	_ = commenter
	CreateComment(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	comment := decodeSuccessEnvelope[Comment](t, recorder)
	if comment.CreatedBy != commenter.ID {
		t.Fatalf("expected created_by=%d, got %+v", commenter.ID, comment)
	}
}
