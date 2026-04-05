package handlers

import (
	"database/sql"
	"log"
	"net/http"
)

func TopicPostsResource(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetPostsByTopic(w, r)
	case http.MethodPost:
		CreatePost(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func PostResource(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetPostByID(w, r)
	case http.MethodPut:
		UpdatePost(w, r)
	case http.MethodDelete:
		DeletePost(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func GetPostsByTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid topic id")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM topics WHERE id = ?)", topicID)
	if err != nil {
		log.Printf("GetPostsByTopic topic lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to verify topic")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}

	rows, err := db().Query(`
		SELECT id, topic_id, title, body, created_by, created_at
		FROM posts
		WHERE topic_id = ?
		ORDER BY created_at ASC
	`, topicID)
	if err != nil {
		log.Printf("GetPostsByTopic query failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch posts")
		return
	}
	defer rows.Close()

	posts := make([]Post, 0)
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.TopicID, &post.Title, &post.Body, &post.CreatedBy, &post.CreatedAt); err != nil {
			log.Printf("GetPostsByTopic scan failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to parse posts")
			return
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetPostsByTopic rows failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to read posts")
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

func GetPostByID(w http.ResponseWriter, r *http.Request) {
	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var post Post
	err = db().QueryRow(`
		SELECT id, topic_id, title, body, created_by, created_at
		FROM posts
		WHERE id = ?
	`, postID).Scan(&post.ID, &post.TopicID, &post.Title, &post.Body, &post.CreatedBy, &post.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		log.Printf("GetPostByID failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load post")
		return
	}

	writeJSON(w, http.StatusOK, post)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid topic id")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM topics WHERE id = ?)", topicID)
	if err != nil {
		log.Printf("CreatePost topic lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to verify topic")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "topic not found")
		return
	}

	var input struct {
		Title     string `json:"title"`
		Body      string `json:"body"`
		CreatedBy int    `json:"created_by"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input.Title = trimRequired(input.Title)
	input.Body = trimRequired(input.Body)
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if input.Body == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}
	if input.CreatedBy <= 0 {
		writeError(w, http.StatusBadRequest, "valid created_by user ID is required")
		return
	}

	result, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (?, ?, ?, ?)
	`, topicID, input.Title, input.Body, input.CreatedBy)
	if err != nil {
		log.Printf("CreatePost insert failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve post")
		return
	}

	var post Post
	err = db().QueryRow(`
		SELECT id, topic_id, title, body, created_by, created_at
		FROM posts
		WHERE id = ?
	`, id).Scan(&post.ID, &post.TopicID, &post.Title, &post.Body, &post.CreatedBy, &post.CreatedAt)
	if err != nil {
		log.Printf("CreatePost reload failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to retrieve post")
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var existing Post
	err = db().QueryRow(`
		SELECT id, topic_id, title, body, created_by, created_at
		FROM posts
		WHERE id = ?
	`, postID).Scan(&existing.ID, &existing.TopicID, &existing.Title, &existing.Body, &existing.CreatedBy, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		log.Printf("UpdatePost lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load post")
		return
	}

	if err := requireOwnership(r, existing.CreatedBy); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	var input struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input.Title = trimRequired(input.Title)
	input.Body = trimRequired(input.Body)
	if input.Title == "" || input.Body == "" {
		writeError(w, http.StatusBadRequest, "title and body are required")
		return
	}

	if _, err := db().Exec(`
		UPDATE posts
		SET title = ?, body = ?
		WHERE id = ?
	`, input.Title, input.Body, postID); err != nil {
		log.Printf("UpdatePost update failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to update post")
		return
	}

	existing.Title = input.Title
	existing.Body = input.Body
	writeJSON(w, http.StatusOK, existing)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var ownerID int
	if err := db().QueryRow(`SELECT created_by FROM posts WHERE id = ?`, postID).Scan(&ownerID); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "post not found")
		return
	} else if err != nil {
		log.Printf("DeletePost lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load post")
		return
	}

	if err := requireOwnership(r, ownerID); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	tx, err := db().Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete post")
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM comments WHERE post_id = ?`, postID); err != nil {
		log.Printf("DeletePost comments failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete post comments")
		return
	}

	result, err := tx.Exec(`DELETE FROM posts WHERE id = ?`, postID)
	if err != nil {
		log.Printf("DeletePost post failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete post")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to finalize post deletion")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
