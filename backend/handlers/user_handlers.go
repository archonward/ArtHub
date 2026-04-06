package handlers

import (
	"database/sql"
	"log"
	"net/http"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	username := trimRequired(input.Username)
	password := input.Password

	if err := validateUsername(username); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := validatePassword(password); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "password_hash_failed", "failed to secure password")
		return
	}

	var existingUserID int
	var existingPasswordHash sql.NullString
	err = db().QueryRow(`
		SELECT id, password_hash
		FROM users
		WHERE username = ?
	`, username).Scan(&existingUserID, &existingPasswordHash)
	switch {
	case err == sql.ErrNoRows:
		result, insertErr := db().Exec(`
			INSERT INTO users (username, password_hash)
			VALUES (?, ?)
		`, username, passwordHash)
		if insertErr != nil {
			log.Printf("Signup insert failed: %v", insertErr)
			writeError(w, http.StatusInternalServerError, "user_create_failed", "failed to create user")
			return
		}
		lastID, insertErr := result.LastInsertId()
		if insertErr != nil {
			writeError(w, http.StatusInternalServerError, "user_create_failed", "failed to retrieve user")
			return
		}
		existingUserID = int(lastID)
	case err != nil:
		log.Printf("Signup lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "signup_failed", "failed to create account")
		return
	default:
		if existingPasswordHash.Valid && trimRequired(existingPasswordHash.String) != "" {
			writeError(w, http.StatusConflict, "username_taken", "username is already registered")
			return
		}

		if _, updateErr := db().Exec(`
			UPDATE users
			SET password_hash = ?
			WHERE id = ?
		`, passwordHash, existingUserID); updateErr != nil {
			log.Printf("Signup legacy-user upgrade failed: %v", updateErr)
			writeError(w, http.StatusInternalServerError, "signup_failed", "failed to activate account")
			return
		}
	}

	token, session, err := createSession(existingUserID)
	if err != nil {
		log.Printf("Signup session creation failed: %v", err)
		writeError(w, http.StatusInternalServerError, "session_create_failed", "failed to create session")
		return
	}
	setSessionCookie(w, token, session.ExpiresAt)

	user, err := getUserByID(existingUserID)
	if err != nil || user == nil {
		writeError(w, http.StatusInternalServerError, "user_query_failed", "failed to load current user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	username := trimRequired(input.Username)
	password := input.Password

	if username == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "username is required")
		return
	}
	if password == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "password is required")
		return
	}

	var userID int
	var passwordHash sql.NullString
	err := db().QueryRow(`
		SELECT id, password_hash
		FROM users
		WHERE username = ?
	`, username).Scan(&userID, &passwordHash)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
		return
	}
	if err != nil {
		log.Printf("Login lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "login_failed", "login failed")
		return
	}

	if !passwordHash.Valid || trimRequired(passwordHash.String) == "" {
		writeError(w, http.StatusUnauthorized, "password_not_set", "account exists but does not have a password yet")
		return
	}

	if err := comparePassword(passwordHash.String, password); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
		return
	}

	token, session, err := createSession(userID)
	if err != nil {
		log.Printf("Login session creation failed: %v", err)
		writeError(w, http.StatusInternalServerError, "session_create_failed", "failed to create session")
		return
	}
	setSessionCookie(w, token, session.ExpiresAt)

	user, err := getUserByID(userID)
	if err != nil || user == nil {
		writeError(w, http.StatusInternalServerError, "user_query_failed", "failed to load current user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	token, err := sessionTokenFromRequest(r)
	if err == nil {
		if deleteErr := deleteSessionByToken(token); deleteErr != nil {
			log.Printf("Logout session delete failed: %v", deleteErr)
			writeError(w, http.StatusInternalServerError, "logout_failed", "failed to logout")
			return
		}
	}

	clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"logged_out": true})
}

func CurrentSessionUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	user, err := requireAuthenticatedUser(r)
	if err != nil {
		if err.Error() == "authentication required" {
			writeError(w, http.StatusUnauthorized, "not_authenticated", "authentication required")
			return
		}
		log.Printf("CurrentSessionUser failed: %v", err)
		writeError(w, http.StatusInternalServerError, "session_resolution_failed", "failed to resolve session")
		return
	}

	writeJSON(w, http.StatusOK, user)
}
