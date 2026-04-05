package handlers

import (
	"log"
	"net/http"
)

func PostCommentsResource(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetCommentsByPost(w, r)
	case http.MethodPost:
		CreateComment(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func GetCommentsByPost(w http.ResponseWriter, r *http.Request) {
	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postID)
	if err != nil {
		log.Printf("GetCommentsByPost post lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to verify post")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}

	rows, err := db().Query(`
		SELECT id, post_id, body, created_by, created_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at ASC
	`, postID)
	if err != nil {
		log.Printf("GetCommentsByPost query failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch comments")
		return
	}
	defer rows.Close()

	comments := make([]Comment, 0)
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.Body, &comment.CreatedBy, &comment.CreatedAt); err != nil {
			log.Printf("GetCommentsByPost scan failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to parse comments")
			return
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetCommentsByPost rows failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to read comments")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postID)
	if err != nil {
		log.Printf("CreateComment post lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to verify post")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}

	var input struct {
		Body      string `json:"body"`
		CreatedBy int    `json:"created_by"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input.Body = trimRequired(input.Body)
	if input.Body == "" {
		writeError(w, http.StatusBadRequest, "comment body is required")
		return
	}
	if input.CreatedBy <= 0 {
		writeError(w, http.StatusBadRequest, "valid created_by user ID is required")
		return
	}

	result, err := db().Exec(`
		INSERT INTO comments (post_id, body, created_by)
		VALUES (?, ?, ?)
	`, postID, input.Body, input.CreatedBy)
	if err != nil {
		log.Printf("CreateComment insert failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve comment")
		return
	}

	var comment Comment
	err = db().QueryRow(`
		SELECT id, post_id, body, created_by, created_at
		FROM comments
		WHERE id = ?
	`, id).Scan(&comment.ID, &comment.PostID, &comment.Body, &comment.CreatedBy, &comment.CreatedAt)
	if err != nil {
		log.Printf("CreateComment reload failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to retrieve comment")
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}
