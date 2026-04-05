package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type errorResponse struct {
	Error string `json:"error"`
}

func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Backend is running, database connected",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethods ...string) {
	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	}
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if decoder.More() {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func parsePathID(r *http.Request, key string) (int, error) {
	value := r.PathValue(key)
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}

	return id, nil
}

func trimRequired(value string) string {
	return strings.TrimSpace(value)
}

func optionalActorID(r *http.Request) (int, error) {
	raw := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if raw == "" {
		return 0, nil
	}

	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid X-User-ID header")
	}

	return id, nil
}

func requireOwnership(r *http.Request, ownerID int) error {
	actorID, err := optionalActorID(r)
	if err != nil {
		return err
	}

	if actorID == 0 {
		return nil
	}

	if actorID != ownerID {
		return errors.New("you are not allowed to modify this resource")
	}

	return nil
}

func resourceExists(query string, id int) (bool, error) {
	var exists bool
	err := db().QueryRow(query, id).Scan(&exists)
	return exists, err
}

func notFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
