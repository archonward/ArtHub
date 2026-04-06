package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type ResponseEnvelope[T any] struct {
	Data T `json:"data"`
}

type ErrorResponseEnvelope struct {
	Error ErrorDetail `json:"error"`
}

func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Backend is running, database connected",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ResponseEnvelope[any]{Data: payload})
}

func writeNoContent(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponseEnvelope{
		Error: ErrorDetail{
			Message: message,
			Code:    code,
		},
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethods ...string) {
	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	}
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func decodeJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body is required")
		}
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
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

func resourceExists(query string, id int) (bool, error) {
	var exists bool
	err := db().QueryRow(query, id).Scan(&exists)
	return exists, err
}

func notFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func validationError(message string) (string, string) {
	return "validation_error", message
}

func malformedJSONError(err error) (string, string) {
	switch {
	case strings.Contains(err.Error(), "unknown field"):
		return "invalid_json", err.Error()
	case strings.Contains(err.Error(), "request body is required"):
		return "invalid_json", "request body is required"
	case strings.Contains(err.Error(), "single JSON object"):
		return "invalid_json", "request body must contain a single JSON object"
	default:
		return "invalid_json", "invalid JSON payload"
	}
}

func authError(err error) (int, string, string) {
	switch {
	case errors.Is(err, errAuthenticationRequired):
		return http.StatusUnauthorized, "not_authenticated", "authentication required"
	case errors.Is(err, errForbidden):
		return http.StatusForbidden, "forbidden", "you are not allowed to modify this resource"
	default:
		return http.StatusInternalServerError, "authorization_failed", "failed to authorize request"
	}
}
