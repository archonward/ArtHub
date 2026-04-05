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

func TestLoginCreatesUser(t *testing.T) {
	setupTestDB(t)

	recorder := executeJSONRequest(Login, http.MethodPost, "/login", map[string]string{
		"username": "arthur",
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var user User
	if err := json.Unmarshal(recorder.Body.Bytes(), &user); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if user.ID == 0 || user.Username != "arthur" {
		t.Fatalf("unexpected user payload: %+v", user)
	}
}

func TestGetTopicByIDReturnsTopic(t *testing.T) {
	setupTestDB(t)

	result, err := db().Exec(`INSERT INTO users (username) VALUES ('owner')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, _ := result.LastInsertId()

	result, err = db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, userID)
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

	var topic Topic
	if err := json.Unmarshal(recorder.Body.Bytes(), &topic); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if int64(topic.ID) != topicID || topic.Title != "Markets" {
		t.Fatalf("unexpected topic payload: %+v", topic)
	}
}

func TestUpdateTopicRejectsWrongOwnerHeader(t *testing.T) {
	setupTestDB(t)

	result, err := db().Exec(`INSERT INTO users (username) VALUES ('owner')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	ownerID, _ := result.LastInsertId()

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, ownerID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"title":       "Updated",
		"description": "Updated description",
	})

	req := httptest.NewRequest(http.MethodPut, "/topics/1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "999")
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	UpdateTopic(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}
