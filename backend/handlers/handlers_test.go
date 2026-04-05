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

func seedTopic(t *testing.T) (int64, int64) {
	t.Helper()

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

	return userID, topicID
}

func seedPost(t *testing.T) (int64, int64, int64) {
	t.Helper()

	userID, topicID := seedTopic(t)
	result, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (?, 'First post', 'Body copy', ?)
	`, topicID, userID)
	if err != nil {
		t.Fatalf("insert post: %v", err)
	}
	postID, _ := result.LastInsertId()
	return userID, topicID, postID
}

func TestLoginCreatesUser(t *testing.T) {
	setupTestDB(t)

	recorder := executeJSONRequest(Login, http.MethodPost, "/login", map[string]string{
		"username": "arthur",
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	user := decodeSuccessEnvelope[User](t, recorder)
	if user.ID == 0 || user.Username != "arthur" {
		t.Fatalf("unexpected user payload: %+v", user)
	}
}

func TestLoginRejectsWhitespaceUsername(t *testing.T) {
	setupTestDB(t)

	recorder := executeJSONRequest(Login, http.MethodPost, "/login", map[string]string{
		"username": "   ",
	})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "validation_error" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestGetTopicByIDReturnsTopic(t *testing.T) {
	setupTestDB(t)
	_, topicID := seedTopic(t)

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

func TestCreateTopicRejectsUnknownCreatedBy(t *testing.T) {
	setupTestDB(t)

	recorder := executeJSONRequest(CreateTopic, http.MethodPost, "/topics", map[string]any{
		"title":       "Markets",
		"description": "desc",
		"created_by":  999,
	})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "invalid_created_by" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestCreatePostRejectsUnknownFieldPayload(t *testing.T) {
	setupTestDB(t)
	userID, topicID := seedTopic(t)

	payload := []byte(`{"title":"Markets","body":"Body","created_by":1,"extra":"nope"}`)
	req := httptest.NewRequest(http.MethodPost, "/topics/1/posts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	_ = userID
	_ = topicID
	CreatePost(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "invalid_json" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestUpdateTopicRejectsMissingActorHeader(t *testing.T) {
	setupTestDB(t)
	_, _ = seedTopic(t)

	payload, _ := json.Marshal(map[string]string{
		"title":       "Updated",
		"description": "Updated description",
	})

	req := httptest.NewRequest(http.MethodPut, "/topics/1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	UpdateTopic(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "actor_required" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestUpdateTopicRejectsWrongOwnerHeader(t *testing.T) {
	setupTestDB(t)
	_, _ = seedTopic(t)

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

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "forbidden" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestDeletePostRejectsMissingResource(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodDelete, "/posts/42", nil)
	req.Header.Set("X-User-ID", "1")
	req.SetPathValue("id", "42")
	recorder := httptest.NewRecorder()

	DeletePost(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "post_not_found" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestDeleteTopicCascadesToPostsAndComments(t *testing.T) {
	setupTestDB(t)
	userID, topicID, postID := seedPost(t)

	if _, err := db().Exec(`
		INSERT INTO comments (post_id, body, created_by)
		VALUES (?, 'Comment body', ?)
	`, postID, userID); err != nil {
		t.Fatalf("insert comment: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/topics/1", nil)
	req.Header.Set("X-User-ID", "1")
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	DeleteTopic(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	deleted := decodeSuccessEnvelope[map[string]bool](t, recorder)
	if !deleted["deleted"] {
		t.Fatalf("expected deleted response, got %+v", deleted)
	}

	var count int
	if err := db().QueryRow(`SELECT COUNT(*) FROM topics WHERE id = ?`, topicID).Scan(&count); err != nil {
		t.Fatalf("count topics: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected topic to be deleted")
	}

	if err := db().QueryRow(`SELECT COUNT(*) FROM posts WHERE id = ?`, postID).Scan(&count); err != nil {
		t.Fatalf("count posts: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected post to be deleted")
	}

	if err := db().QueryRow(`SELECT COUNT(*) FROM comments WHERE post_id = ?`, postID).Scan(&count); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected comments to be deleted")
	}
}

func TestCreateCommentRejectsWhitespaceBody(t *testing.T) {
	setupTestDB(t)
	userID, _, _ := seedPost(t)

	payload, _ := json.Marshal(map[string]any{
		"body":       "   ",
		"created_by": userID,
	})

	req := httptest.NewRequest(http.MethodPost, "/posts/1/comments", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	CreateComment(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "validation_error" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}
