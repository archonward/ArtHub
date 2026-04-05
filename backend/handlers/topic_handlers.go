package handlers

import (
	"database/sql"
	"log"
	"net/http"
)

func TopicsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetTopics(w, r)
	case http.MethodPost:
		CreateTopic(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func TopicResource(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetTopicByID(w, r)
	case http.MethodPut:
		UpdateTopic(w, r)
	case http.MethodDelete:
		DeleteTopic(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func GetTopics(w http.ResponseWriter, r *http.Request) {
	rows, err := db().Query(`
		SELECT id, title, description, created_by, created_at
		FROM topics
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("GetTopics query failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load topics")
		return
	}
	defer rows.Close()

	topics := make([]Topic, 0)
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.ID, &topic.Title, &topic.Description, &topic.CreatedBy, &topic.CreatedAt); err != nil {
			log.Printf("GetTopics scan failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to parse topics")
			return
		}
		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetTopics rows failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to read topics")
		return
	}

	writeJSON(w, http.StatusOK, topics)
}

func GetTopicByID(w http.ResponseWriter, r *http.Request) {
	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid topic id")
		return
	}

	var topic Topic
	err = db().QueryRow(`
		SELECT id, title, description, created_by, created_at
		FROM topics
		WHERE id = ?
	`, topicID).Scan(&topic.ID, &topic.Title, &topic.Description, &topic.CreatedBy, &topic.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}
	if err != nil {
		log.Printf("GetTopicByID failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load topic")
		return
	}

	writeJSON(w, http.StatusOK, topic)
}

func CreateTopic(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		CreatedBy   int    `json:"created_by"`
	}

	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input.Title = trimRequired(input.Title)
	input.Description = trimRequired(input.Description)

	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if input.CreatedBy <= 0 {
		writeError(w, http.StatusBadRequest, "valid created_by user ID is required")
		return
	}

	result, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES (?, ?, ?)
	`, input.Title, input.Description, input.CreatedBy)
	if err != nil {
		log.Printf("CreateTopic insert failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create topic")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("CreateTopic lastInsertId failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to retrieve topic")
		return
	}

	var topic Topic
	err = db().QueryRow(`
		SELECT id, title, description, created_by, created_at
		FROM topics
		WHERE id = ?
	`, id).Scan(&topic.ID, &topic.Title, &topic.Description, &topic.CreatedBy, &topic.CreatedAt)
	if err != nil {
		log.Printf("CreateTopic reload failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to retrieve topic")
		return
	}

	writeJSON(w, http.StatusCreated, topic)
}

func UpdateTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid topic id")
		return
	}

	var existing Topic
	err = db().QueryRow(`
		SELECT id, title, description, created_by, created_at
		FROM topics
		WHERE id = ?
	`, topicID).Scan(&existing.ID, &existing.Title, &existing.Description, &existing.CreatedBy, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}
	if err != nil {
		log.Printf("UpdateTopic lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load topic")
		return
	}

	if err := requireOwnership(r, existing.CreatedBy); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input.Title = trimRequired(input.Title)
	input.Description = trimRequired(input.Description)
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	if _, err := db().Exec(`
		UPDATE topics
		SET title = ?, description = ?
		WHERE id = ?
	`, input.Title, input.Description, topicID); err != nil {
		log.Printf("UpdateTopic update failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to update topic")
		return
	}

	existing.Title = input.Title
	existing.Description = input.Description
	writeJSON(w, http.StatusOK, existing)
}

func DeleteTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid topic id")
		return
	}

	var ownerID int
	if err := db().QueryRow(`SELECT created_by FROM topics WHERE id = ?`, topicID).Scan(&ownerID); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	} else if err != nil {
		log.Printf("DeleteTopic lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load topic")
		return
	}

	if err := requireOwnership(r, ownerID); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	tx, err := db().Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete topic")
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM comments WHERE post_id IN (SELECT id FROM posts WHERE topic_id = ?)`, topicID); err != nil {
		log.Printf("DeleteTopic comments failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete topic comments")
		return
	}
	if _, err := tx.Exec(`DELETE FROM posts WHERE topic_id = ?`, topicID); err != nil {
		log.Printf("DeleteTopic posts failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete topic posts")
		return
	}

	result, err := tx.Exec(`DELETE FROM topics WHERE id = ?`, topicID)
	if err != nil {
		log.Printf("DeleteTopic topic failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete topic")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to finalize topic deletion")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
