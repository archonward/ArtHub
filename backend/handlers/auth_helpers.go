package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName    = "arthub_session"
	sessionDuration      = 7 * 24 * time.Hour
	minPasswordLength    = 8
	maxPasswordLength    = 72 // bcrypt limit
	minUsernameLength    = 3
	maxUsernameLength    = 32
	authUserContextKey   = contextKey("auth_user")
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var (
	errAuthenticationRequired = errors.New("authentication required")
	errForbidden              = errors.New("forbidden")
)

type contextKey string

type authSession struct {
	ID        int
	UserID    int
	TokenHash string
	ExpiresAt time.Time
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func validateUsername(username string) error {
	switch {
	case username == "":
		return errors.New("username is required")
	case len(username) < minUsernameLength:
		return errors.New("username must be at least 3 characters")
	case len(username) > maxUsernameLength:
		return errors.New("username must be 32 characters or fewer")
	case !usernamePattern.MatchString(username):
		return errors.New("username may only contain letters, numbers, underscores, and hyphens")
	default:
		return nil
	}
}

func validatePassword(password string) error {
	switch {
	case password == "":
		return errors.New("password is required")
	case len(password) < minPasswordLength:
		return errors.New("password must be at least 8 characters")
	case len(password) > maxPasswordLength:
		return errors.New("password must be 72 characters or fewer")
	default:
		return nil
	}
}

func generateSessionToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}

func createSession(userID int) (string, authSession, error) {
	token, tokenHash, err := generateSessionToken()
	if err != nil {
		return "", authSession{}, err
	}

	expiresAt := time.Now().UTC().Add(sessionDuration)
	result, err := db().Exec(`
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES (?, ?, ?)
	`, userID, tokenHash, expiresAt)
	if err != nil {
		return "", authSession{}, err
	}

	sessionID, err := result.LastInsertId()
	if err != nil {
		return "", authSession{}, err
	}

	return token, authSession{
		ID:        int(sessionID),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}, nil
}

func sessionCookieSecure() bool {
	secure, _ := strconv.ParseBool(os.Getenv("ARTHUB_SECURE_COOKIES"))
	return secure
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   sessionCookieSecure(),
		Expires:  expiresAt,
		MaxAge:   int(sessionDuration.Seconds()),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   sessionCookieSecure(),
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func sessionTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	if trimRequired(cookie.Value) == "" {
		return "", http.ErrNoCookie
	}
	return cookie.Value, nil
}

func lookupSessionByToken(token string) (*authSession, error) {
	var session authSession
	err := db().QueryRow(`
		SELECT id, user_id, token_hash, expires_at
		FROM sessions
		WHERE token_hash = ?
	`, hashSessionToken(token)).Scan(&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func deleteSessionByToken(token string) error {
	_, err := db().Exec(`DELETE FROM sessions WHERE token_hash = ?`, hashSessionToken(token))
	return err
}

func getUserByID(userID int) (*User, error) {
	var user User
	err := db().QueryRow(`
		SELECT id, username, created_at
		FROM users
		WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func resolveAuthenticatedUser(r *http.Request) (*User, error) {
	token, err := sessionTokenFromRequest(r)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}
		return nil, err
	}

	session, err := lookupSessionByToken(token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	now := time.Now().UTC()
	if now.After(session.ExpiresAt) {
		_ = deleteSessionByToken(token)
		return nil, nil
	}

	if _, err := db().Exec(`UPDATE sessions SET last_seen_at = ? WHERE id = ?`, now, session.ID); err != nil {
		return nil, err
	}

	return getUserByID(session.UserID)
}

func requireAuthenticatedUser(r *http.Request) (*User, error) {
	if user := currentUserFromContext(r); user != nil {
		return user, nil
	}

	user, err := resolveAuthenticatedUser(r)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errAuthenticationRequired
	}
	return user, nil
}

func authorizeOwnership(user *User, ownerID int) error {
	if user == nil {
		return errAuthenticationRequired
	}
	if user.ID != ownerID {
		return errForbidden
	}
	return nil
}

func attachCurrentUser(r *http.Request, user *User) *http.Request {
	ctx := context.WithValue(r.Context(), authUserContextKey, user)
	return r.WithContext(ctx)
}

func currentUserFromContext(r *http.Request) *User {
	user, _ := r.Context().Value(authUserContextKey).(*User)
	return user
}

func OptionalSessionAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := resolveAuthenticatedUser(r)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "session_resolution_failed", "failed to resolve session")
			return
		}
		next(w, attachCurrentUser(r, user))
	}
}
