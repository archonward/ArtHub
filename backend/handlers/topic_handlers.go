package handlers

import (
	"database/sql"
	"errors"
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
		writeError(w, http.StatusInternalServerError, "topics_query_failed", "failed to load topics")
		return
	}
	defer rows.Close()

	topics := make([]Topic, 0)
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.ID, &topic.Title, &topic.Description, &topic.CreatedBy, &topic.CreatedAt); err != nil {
			log.Printf("GetTopics scan failed: %v", err)
			writeError(w, http.StatusInternalServerError, "topics_parse_failed", "failed to parse topics")
			return
		}
		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetTopics rows failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topics_read_failed", "failed to read topics")
		return
	}

	writeJSON(w, http.StatusOK, topics)
}

func GetTopicByID(w http.ResponseWriter, r *http.Request) {
	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_topic_id", "topic id must be a positive integer")
		return
	}

	var topic Topic
	err = db().QueryRow(`
		SELECT id, title, description, created_by, created_at
		FROM topics
		WHERE id = ?
	`, topicID).Scan(&topic.ID, &topic.Title, &topic.Description, &topic.CreatedBy, &topic.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "topic_not_found", "topic not found")
		return
	}
	if err != nil {
		log.Printf("GetTopicByID failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_query_failed", "failed to load topic")
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
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	input.Title = trimRequired(input.Title)
	input.Description = trimRequired(input.Description)

	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "title is required")
		return
	}
	if input.CreatedBy <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "created_by must be a positive integer")
		return
	}
	if err := ensureUserExists(input.CreatedBy); err != nil {
		if errors.Is(err, errUserNotFound) {
			writeError(w, http.StatusBadRequest, "invalid_created_by", "created_by user does not exist")
			return
		}
		log.Printf("CreateTopic user lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "user_lookup_failed", "failed to verify user")
		return
	}

	result, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES (?, ?, ?)
	`, input.Title, input.Description, input.CreatedBy)
	if err != nil {
		log.Printf("CreateTopic insert failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_create_failed", "failed to create topic")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("CreateTopic lastInsertId failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_create_failed", "failed to retrieve topic")
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
		writeError(w, http.StatusInternalServerError, "topic_query_failed", "failed to retrieve topic")
		return
	}

	writeJSON(w, http.StatusCreated, topic)
}

func UpdateTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_topic_id", "topic id must be a positive integer")
		return
	}

	var existing Topic
	err = db().QueryRow(`
		SELECT id, title, description, created_by, created_at
		FROM topics
		WHERE id = ?
	`, topicID).Scan(&existing.ID, &existing.Title, &existing.Description, &existing.CreatedBy, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "topic_not_found", "topic not found")
		return
	}
	if err != nil {
		log.Printf("UpdateTopic lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_query_failed", "failed to load topic")
		return
	}

	if err := requireOwnership(r, existing.CreatedBy); err != nil {
		status, code, message := actorError(err)
		writeError(w, status, code, message)
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &input); err != nil {
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	input.Title = trimRequired(input.Title)
	input.Description = trimRequired(input.Description)
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "title is required")
		return
	}

	if _, err := db().Exec(`
		UPDATE topics
		SET title = ?, description = ?
		WHERE id = ?
	`, input.Title, input.Description, topicID); err != nil {
		log.Printf("UpdateTopic update failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_update_failed", "failed to update topic")
		return
	}

	existing.Title = input.Title
	existing.Description = input.Description
	writeJSON(w, http.StatusOK, existing)
}

func DeleteTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_topic_id", "topic id must be a positive integer")
		return
	}

	var ownerID int
	if err := db().QueryRow(`SELECT created_by FROM topics WHERE id = ?`, topicID).Scan(&ownerID); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "topic_not_found", "topic not found")
		return
	} else if err != nil {
		log.Printf("DeleteTopic lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_query_failed", "failed to load topic")
		return
	}

	if err := requireOwnership(r, ownerID); err != nil {
		status, code, message := actorError(err)
		writeError(w, status, code, message)
		return
	}

	tx, err := db().Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "topic_delete_failed", "failed to delete topic")
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`DELETE FROM topics WHERE id = ?`, topicID)
	if err != nil {
		log.Printf("DeleteTopic topic failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_delete_failed", "failed to delete topic")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		writeError(w, http.StatusNotFound, "topic_not_found", "topic not found")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "topic_delete_failed", "failed to finalize topic deletion")
		return
	}

	writeNoContent(w)
}
