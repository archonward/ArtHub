package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type nullableVote struct {
	sql.NullInt64
}

type postSort string

const (
	postSortTop         postSort = "top"
	postSortNew         postSort = "new"
	defaultPostPageSize          = 10
	maxPostPageSize              = 50
)

type paginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

type PaginationMetadata struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalItems int  `json:"total_items"`
	TotalPages int  `json:"total_pages"`
	HasPrev    bool `json:"has_prev"`
	HasNext    bool `json:"has_next"`
}

type CompanyPostsPage struct {
	Posts      []Post             `json:"posts"`
	Pagination PaginationMetadata `json:"pagination"`
}

func CompanyPostsResource(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetPostsByCompany(w, r)
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

func PostVoteResource(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		VoteOnPost(w, r)
	case http.MethodDelete:
		DeletePostVote(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodPost, http.MethodDelete)
	}
}

func GetPostsByCompany(w http.ResponseWriter, r *http.Request) {
	companyID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_company_id", "company id must be a positive integer")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM companies WHERE id = ?)", companyID)
	if err != nil {
		log.Printf("GetPostsByCompany company lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_query_failed", "failed to verify company")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "company_not_found", "company not found")
		return
	}

	sortMode, err := parsePostSort(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_sort", err.Error())
		return
	}

	pagination, err := parsePaginationParams(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}

	posts, totalItems, err := loadPostsByCompany(companyID, currentUserFromContext(r), sortMode, pagination)
	if err != nil {
		log.Printf("GetPostsByCompany query failed: %v", err)
		writeError(w, http.StatusInternalServerError, "posts_query_failed", "failed to fetch posts")
		return
	}

	writeJSON(w, http.StatusOK, CompanyPostsPage{
		Posts: posts,
		Pagination: buildPaginationMetadata(
			pagination.Page,
			pagination.PageSize,
			totalItems,
		),
	})
}

func GetPostByID(w http.ResponseWriter, r *http.Request) {
	postID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post_id", "post id must be a positive integer")
		return
	}

	post, err := loadPostByID(postID, currentUserFromContext(r))
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

	companyID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_company_id", "company id must be a positive integer")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM companies WHERE id = ?)", companyID)
	if err != nil {
		log.Printf("CreatePost company lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_query_failed", "failed to verify company")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "company_not_found", "company not found")
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
		INSERT INTO posts (company_id, title, content, created_by)
		VALUES (?, ?, ?, ?)
	`, companyID, input.Title, input.Body, user.ID)
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

	post, err := loadPostByID(int(id), user)
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

	existing, err := loadPostByID(postID, user)
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
		SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, input.Title, input.Body, postID); err != nil {
		log.Printf("UpdatePost update failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_update_failed", "failed to update post")
		return
	}

	updated, err := loadPostByID(postID, user)
	if err != nil {
		log.Printf("UpdatePost reload failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to retrieve post")
		return
	}

	writeJSON(w, http.StatusOK, updated)
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

func VoteOnPost(w http.ResponseWriter, r *http.Request) {
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

	var input struct {
		Value int `json:"value"`
	}
	if err := decodeJSON(r, &input); err != nil {
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	if input.Value != -1 && input.Value != 1 {
		writeError(w, http.StatusBadRequest, "validation_error", "value must be 1 or -1")
		return
	}

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postID)
	if err != nil {
		log.Printf("VoteOnPost post lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to verify post")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}

	if _, err := db().Exec(`
		INSERT INTO votes (user_id, post_id, vote_value)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, post_id)
		DO UPDATE SET vote_value = excluded.vote_value, updated_at = CURRENT_TIMESTAMP
	`, user.ID, postID, input.Value); err != nil {
		log.Printf("VoteOnPost upsert failed: %v", err)
		writeError(w, http.StatusInternalServerError, "vote_save_failed", "failed to save vote")
		return
	}

	post, err := loadPostByID(postID, user)
	if err != nil {
		log.Printf("VoteOnPost reload failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to retrieve post")
		return
	}

	writeJSON(w, http.StatusOK, post)
}

func DeletePostVote(w http.ResponseWriter, r *http.Request) {
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

	exists, err := resourceExists("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postID)
	if err != nil {
		log.Printf("DeletePostVote post lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to verify post")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}

	if _, err := db().Exec(`DELETE FROM votes WHERE user_id = ? AND post_id = ?`, user.ID, postID); err != nil {
		log.Printf("DeletePostVote failed: %v", err)
		writeError(w, http.StatusInternalServerError, "vote_delete_failed", "failed to remove vote")
		return
	}

	post, err := loadPostByID(postID, user)
	if err != nil {
		log.Printf("DeletePostVote reload failed: %v", err)
		writeError(w, http.StatusInternalServerError, "post_query_failed", "failed to retrieve post")
		return
	}

	writeJSON(w, http.StatusOK, post)
}

func loadPostByID(postID int, user *User) (Post, error) {
	currentUserID := 0
	if user != nil {
		currentUserID = user.ID
	}

	var post Post
	var currentUserVote nullableVote
	err := db().QueryRow(`
		SELECT
			p.id,
			p.company_id,
			p.title,
			p.content AS body,
			p.created_by,
			p.created_at,
			p.updated_at,
			COALESCE(SUM(v.vote_value), 0) AS vote_score,
			MAX(CASE WHEN v.user_id = ? THEN v.vote_value END) AS current_user_vote
		FROM posts p
		LEFT JOIN votes v ON v.post_id = p.id
		WHERE p.id = ?
		GROUP BY p.id, p.company_id, p.title, p.content, p.created_by, p.created_at, p.updated_at
	`, currentUserID, postID).Scan(
		&post.ID,
		&post.CompanyID,
		&post.Title,
		&post.Body,
		&post.CreatedBy,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.VoteScore,
		&currentUserVote,
	)
	if err != nil {
		return Post{}, err
	}

	post.CurrentUserVote = currentUserVote.pointer()
	return post, nil
}

func loadPostsByCompany(companyID int, user *User, sortMode postSort, pagination paginationParams) ([]Post, int, error) {
	currentUserID := 0
	if user != nil {
		currentUserID = user.ID
	}

	var totalItems int
	if err := db().QueryRow(`SELECT COUNT(*) FROM posts WHERE company_id = ?`, companyID).Scan(&totalItems); err != nil {
		return nil, 0, err
	}

	rows, err := db().Query(fmt.Sprintf(`
		SELECT
			p.id,
			p.company_id,
			p.title,
			p.content AS body,
			p.created_by,
			p.created_at,
			p.updated_at,
			COALESCE(SUM(v.vote_value), 0) AS vote_score,
			MAX(CASE WHEN v.user_id = ? THEN v.vote_value END) AS current_user_vote
		FROM posts p
		LEFT JOIN votes v ON v.post_id = p.id
		WHERE p.company_id = ?
		GROUP BY p.id, p.company_id, p.title, p.content, p.created_by, p.created_at, p.updated_at
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, sortOrderClause(sortMode)), currentUserID, companyID, pagination.PageSize, pagination.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	posts := make([]Post, 0)
	for rows.Next() {
		var post Post
		var currentUserVote nullableVote
		if err := rows.Scan(
			&post.ID,
			&post.CompanyID,
			&post.Title,
			&post.Body,
			&post.CreatedBy,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.VoteScore,
			&currentUserVote,
		); err != nil {
			return nil, 0, err
		}
		post.CurrentUserVote = currentUserVote.pointer()
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return posts, totalItems, nil
}

func parsePostSort(r *http.Request) (postSort, error) {
	rawSort := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("sort")))
	if rawSort == "" {
		return postSortTop, nil
	}

	switch postSort(rawSort) {
	case postSortTop, postSortNew:
		return postSort(rawSort), nil
	default:
		return "", fmt.Errorf("sort must be one of: top, new")
	}
}

func sortOrderClause(sortMode postSort) string {
	switch sortMode {
	case postSortNew:
		return "p.created_at DESC, p.id DESC"
	case postSortTop:
		fallthrough
	default:
		return "vote_score DESC, p.created_at DESC, p.id DESC"
	}
}

func parsePaginationParams(r *http.Request) (paginationParams, error) {
	page := 1
	pageSize := defaultPostPageSize

	if rawPage := strings.TrimSpace(r.URL.Query().Get("page")); rawPage != "" {
		parsedPage, err := strconv.Atoi(rawPage)
		if err != nil || parsedPage <= 0 {
			return paginationParams{}, fmt.Errorf("page must be a positive integer")
		}
		page = parsedPage
	}

	if rawPageSize := strings.TrimSpace(r.URL.Query().Get("pageSize")); rawPageSize != "" {
		parsedPageSize, err := strconv.Atoi(rawPageSize)
		if err != nil || parsedPageSize <= 0 {
			return paginationParams{}, fmt.Errorf("pageSize must be a positive integer")
		}
		if parsedPageSize > maxPostPageSize {
			return paginationParams{}, fmt.Errorf("pageSize must be %d or fewer", maxPostPageSize)
		}
		pageSize = parsedPageSize
	}

	return paginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}, nil
}

func buildPaginationMetadata(page, pageSize, totalItems int) PaginationMetadata {
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}

	return PaginationMetadata{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasPrev:    page > 1 && totalPages > 0,
		HasNext:    page < totalPages,
	}
}

func (vote nullableVote) pointer() *int {
	if !vote.Valid {
		return nil
	}

	value := int(vote.Int64)
	return &value
}
