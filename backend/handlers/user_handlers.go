package handlers

import (
	"database/sql"
	"log"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var input struct {
		Username string `json:"username"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	username := trimRequired(input.Username)
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	var user User
	err := db().QueryRow(`SELECT id, username FROM users WHERE username = ?`, username).Scan(&user.ID, &user.Username)
	switch {
	case err == sql.ErrNoRows:
		result, insertErr := db().Exec(`INSERT INTO users (username) VALUES (?)`, username)
		if insertErr != nil {
			log.Printf("Login insert failed: %v", insertErr)
			writeError(w, http.StatusInternalServerError, "failed to register user")
			return
		}

		userID, insertErr := result.LastInsertId()
		if insertErr != nil {
			writeError(w, http.StatusInternalServerError, "failed to retrieve user")
			return
		}
		user = User{ID: int(userID), Username: username}
	case err != nil:
		log.Printf("Login lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	writeJSON(w, http.StatusOK, user)
}
