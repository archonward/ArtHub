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
		writeError(w, http.StatusBadRequest, "invalid_topic_id", "topic id must be a positive integer")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM topics WHERE id = ?)", topicID)
	if err != nil {
		log.Printf("GetPostsByTopic topic lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_query_failed", "failed to verify topic")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "topic_not_found", "topic not found")
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
		writeError(w, http.StatusInternalServerError, "posts_query_failed", "failed to fetch posts")
		return
	}
	defer rows.Close()

	posts := make([]Post, 0)
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.TopicID, &post.Title, &post.Body, &post.CreatedBy, &post.CreatedAt); err != nil {
			log.Printf("GetPostsByTopic scan failed: %v", err)
			writeError(w, http.StatusInternalServerError, "posts_parse_failed", "failed to parse posts")
			return
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetPostsByTopic rows failed: %v", err)
		writeError(w, http.StatusInternalServerError, "posts_read_failed", "failed to read posts")
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

func GetPostByID(w http.ResponseWriter, r *http.Request) {
	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post_id", "post id must be a positive integer")
		return
	}

	var post Post
	err = db().QueryRow(`
		SELECT id, topic_id, title, body, created_by, created_at
		FROM posts
		WHERE id = ?
	`, postID).Scan(&post.ID, &post.TopicID, &post.Title, &post.Body, &post.CreatedBy, &post.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}
	if err != nil {
		log.Printf("GetPostByID failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to load post")
		return
	}

	writeJSON(w, http.StatusOK, post)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	user, err := requireAuthenticatedUser(r)
	if err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	topicID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_topic_id", "topic id must be a positive integer")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM topics WHERE id = ?)", topicID)
	if err != nil {
		log.Printf("CreatePost topic lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "topic_query_failed", "failed to verify topic")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "topic_not_found", "topic not found")
		return
	}

	var input struct {
		Title     string `json:"title"`
		Body      string `json:"body"`
		CreatedBy int    `json:"created_by"`
	}
	if err := decodeJSON(r, &input); err != nil {
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	input.Title = trimRequired(input.Title)
	input.Body = trimRequired(input.Body)
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "title is required")
		return
	}
	if input.Body == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "body is required")
		return
	}

	result, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (?, ?, ?, ?)
	`, topicID, input.Title, input.Body, user.ID)
	if err != nil {
		log.Printf("CreatePost insert failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_create_failed", "failed to create post")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "post_create_failed", "failed to retrieve post")
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
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to retrieve post")
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	user, err := requireAuthenticatedUser(r)
	if err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post_id", "post id must be a positive integer")
		return
	}

	var existing Post
	err = db().QueryRow(`
		SELECT id, topic_id, title, body, created_by, created_at
		FROM posts
		WHERE id = ?
	`, postID).Scan(&existing.ID, &existing.TopicID, &existing.Title, &existing.Body, &existing.CreatedBy, &existing.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}
	if err != nil {
		log.Printf("UpdatePost lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to load post")
		return
	}

	if err := authorizeOwnership(user, existing.CreatedBy); err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	var input struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := decodeJSON(r, &input); err != nil {
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	input.Title = trimRequired(input.Title)
	input.Body = trimRequired(input.Body)
	if input.Title == "" || input.Body == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "title and body are required")
		return
	}

	if _, err := db().Exec(`
		UPDATE posts
		SET title = ?, body = ?
		WHERE id = ?
	`, input.Title, input.Body, postID); err != nil {
		log.Printf("UpdatePost update failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_update_failed", "failed to update post")
		return
	}

	existing.Title = input.Title
	existing.Body = input.Body
	writeJSON(w, http.StatusOK, existing)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	user, err := requireAuthenticatedUser(r)
	if err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post_id", "post id must be a positive integer")
		return
	}

	var ownerID int
	if err := db().QueryRow(`SELECT created_by FROM posts WHERE id = ?`, postID).Scan(&ownerID); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	} else if err != nil {
		log.Printf("DeletePost lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to load post")
		return
	}

	if err := authorizeOwnership(user, ownerID); err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	tx, err := db().Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "post_delete_failed", "failed to delete post")
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`DELETE FROM posts WHERE id = ?`, postID)
	if err != nil {
		log.Printf("DeletePost post failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_delete_failed", "failed to delete post")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "post_delete_failed", "failed to finalize post deletion")
		return
	}

	writeNoContent(w)
}
