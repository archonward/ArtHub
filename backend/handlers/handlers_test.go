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

func signupAndSessionCookie(t *testing.T, username string) (*User, *http.Cookie) {
	t.Helper()

	recorder := executeJSONRequest(Signup, http.MethodPost, "/auth/signup", map[string]string{
		"username": username,
		"password": "verysecurepassword",
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected signup 201, got %d", recorder.Code)
	}

	user := decodeSuccessEnvelope[User](t, recorder)
	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}

	return &user, cookies[0]
}

func TestGetTopicByIDReturnsTopic(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	result, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID)
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

	topic := decodeSuccessEnvelope[Topic](t, recorder)
	if int64(topic.ID) != topicID || topic.Title != "Markets" {
		t.Fatalf("unexpected topic payload: %+v", topic)
	}
}

func TestCreateTopicRequiresAuthentication(t *testing.T) {
	setupTestDB(t)

	recorder := executeJSONRequest(CreateTopic, http.MethodPost, "/topics", map[string]any{
		"title":       "Markets",
		"description": "desc",
	})

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "not_authenticated" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestCreateTopicUsesAuthenticatedUserInsteadOfCreatedByPayload(t *testing.T) {
	setupTestDB(t)
	user, cookie := signupAndSessionCookie(t, "owner")

	payload, _ := json.Marshal(map[string]any{
		"title":       "Markets",
		"description": "desc",
		"created_by":  999,
	})

	req := httptest.NewRequest(http.MethodPost, "/topics", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	recorder := httptest.NewRecorder()

	CreateTopic(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	topic := decodeSuccessEnvelope[Topic](t, recorder)
	if topic.CreatedBy != user.ID {
		t.Fatalf("expected created_by=%d, got %+v", user.ID, topic)
	}
}

func TestCreateTopicRejectsXUserIDBypassWithoutSession(t *testing.T) {
	setupTestDB(t)

	payload, _ := json.Marshal(map[string]any{
		"title":       "Markets",
		"description": "desc",
		"created_by":  1,
	})

	req := httptest.NewRequest(http.MethodPost, "/topics", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	recorder := httptest.NewRecorder()

	CreateTopic(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "not_authenticated" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestUpdateTopicRejectsNonOwnerEvenWithSpoofedXUserID(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	otherUser, otherCookie := signupAndSessionCookie(t, "other")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"title":       "Updated",
		"description": "Updated description",
	})

	req := httptest.NewRequest(http.MethodPut, "/topics/1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	req.AddCookie(otherCookie)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	_ = otherUser
	UpdateTopic(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "forbidden" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestUpdateTopicSucceedsForOwnerSession(t *testing.T) {
	setupTestDB(t)
	owner, ownerCookie := signupAndSessionCookie(t, "owner")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	payload, _ := json.Marshal(map[string]string{
		"title":       "Updated",
		"description": "Updated description",
	})

	req := httptest.NewRequest(http.MethodPut, "/topics/1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(ownerCookie)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	UpdateTopic(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	topic := decodeSuccessEnvelope[Topic](t, recorder)
	if topic.Title != "Updated" {
		t.Fatalf("expected updated topic, got %+v", topic)
	}
}

func TestDeletePostRequiresAuthentication(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/posts/1", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	DeletePost(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestCreateCommentUsesAuthenticatedUserInsteadOfCreatedByPayload(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	commenter, commentCookie := signupAndSessionCookie(t, "commenter")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	payload, _ := json.Marshal(map[string]any{
		"body":       "Comment body",
		"created_by": 999,
	})

	req := httptest.NewRequest(http.MethodPost, "/posts/1/comments", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(commentCookie)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	_ = commenter
	CreateComment(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	comment := decodeSuccessEnvelope[Comment](t, recorder)
	if comment.CreatedBy != commenter.ID {
		t.Fatalf("expected created_by=%d, got %+v", commenter.ID, comment)
	}
}

func TestVoteOnPostRequiresAuthentication(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/posts/1/vote", bytes.NewReader([]byte(`{"value":1}`)))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	VoteOnPost(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestVoteOnPostReturnsUpdatedVoteSummary(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	voter, voterCookie := signupAndSessionCookie(t, "voter")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/posts/1/vote", bytes.NewReader([]byte(`{"value":1}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(voterCookie)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	VoteOnPost(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	post := decodeSuccessEnvelope[Post](t, recorder)
	if post.VoteScore != 1 {
		t.Fatalf("expected vote_score=1, got %+v", post)
	}
	if post.CurrentUserVote == nil || *post.CurrentUserVote != 1 {
		t.Fatalf("expected current_user_vote=1, got %+v", post)
	}

	var rowCount int
	if err := db().QueryRow(`SELECT COUNT(*) FROM votes WHERE user_id = ? AND post_id = 1`, voter.ID).Scan(&rowCount); err != nil {
		t.Fatalf("count votes: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("expected one vote row, got %d", rowCount)
	}
}

func TestVoteOnPostSwitchesExistingVote(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	voter, voterCookie := signupAndSessionCookie(t, "voter")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	firstReq := httptest.NewRequest(http.MethodPost, "/posts/1/vote", bytes.NewReader([]byte(`{"value":1}`)))
	firstReq.Header.Set("Content-Type", "application/json")
	firstReq.AddCookie(voterCookie)
	firstReq.SetPathValue("id", "1")
	firstRecorder := httptest.NewRecorder()
	VoteOnPost(firstRecorder, firstReq)

	secondReq := httptest.NewRequest(http.MethodPost, "/posts/1/vote", bytes.NewReader([]byte(`{"value":-1}`)))
	secondReq.Header.Set("Content-Type", "application/json")
	secondReq.AddCookie(voterCookie)
	secondReq.SetPathValue("id", "1")
	secondRecorder := httptest.NewRecorder()
	VoteOnPost(secondRecorder, secondReq)

	if secondRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", secondRecorder.Code)
	}

	post := decodeSuccessEnvelope[Post](t, secondRecorder)
	if post.VoteScore != -1 {
		t.Fatalf("expected vote_score=-1, got %+v", post)
	}
	if post.CurrentUserVote == nil || *post.CurrentUserVote != -1 {
		t.Fatalf("expected current_user_vote=-1, got %+v", post)
	}

	var storedVote int
	if err := db().QueryRow(`SELECT vote_value FROM votes WHERE user_id = ? AND post_id = 1`, voter.ID).Scan(&storedVote); err != nil {
		t.Fatalf("select vote: %v", err)
	}
	if storedVote != -1 {
		t.Fatalf("expected stored vote -1, got %d", storedVote)
	}
}

func TestVoteOnPostIsIdempotentForSameValue(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	voter, voterCookie := signupAndSessionCookie(t, "voter")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	for range 2 {
		req := httptest.NewRequest(http.MethodPost, "/posts/1/vote", bytes.NewReader([]byte(`{"value":1}`)))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(voterCookie)
		req.SetPathValue("id", "1")
		recorder := httptest.NewRecorder()
		VoteOnPost(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	}

	var rowCount int
	if err := db().QueryRow(`SELECT COUNT(*) FROM votes WHERE user_id = ? AND post_id = 1`, voter.ID).Scan(&rowCount); err != nil {
		t.Fatalf("count votes: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("expected one vote row, got %d", rowCount)
	}

	var voteScore int
	if err := db().QueryRow(`SELECT COALESCE(SUM(vote_value), 0) FROM votes WHERE post_id = 1`).Scan(&voteScore); err != nil {
		t.Fatalf("sum votes: %v", err)
	}
	if voteScore != 1 {
		t.Fatalf("expected vote score 1, got %d", voteScore)
	}
}

func TestDeletePostVoteRemovesCurrentUsersVote(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	voter, voterCookie := signupAndSessionCookie(t, "voter")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by)
		VALUES (1, 'First post', 'Body copy', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert post: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO votes (user_id, post_id, vote_value)
		VALUES (?, 1, 1)
	`, voter.ID); err != nil {
		t.Fatalf("insert vote: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/posts/1/vote", nil)
	req.AddCookie(voterCookie)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	DeletePostVote(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	post := decodeSuccessEnvelope[Post](t, recorder)
	if post.VoteScore != 0 {
		t.Fatalf("expected vote_score=0, got %+v", post)
	}
	if post.CurrentUserVote != nil {
		t.Fatalf("expected current_user_vote to be nil, got %+v", post)
	}
}

func TestGetPostsByTopicSortsByTop(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	voterA, _ := signupAndSessionCookie(t, "voterA")
	voterB, _ := signupAndSessionCookie(t, "voterB")

	insertTopicAndPostsForSorting(t, owner.ID)

	if _, err := db().Exec(`INSERT INTO votes (user_id, post_id, vote_value) VALUES (?, 1, 1), (?, 1, 1), (?, 2, 1)`, voterA.ID, voterB.ID, voterA.ID); err != nil {
		t.Fatalf("insert votes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/topics/1/posts?sort=top", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	GetPostsByTopic(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	page := decodeSuccessEnvelope[TopicPostsPage](t, recorder)
	posts := page.Posts
	if len(posts) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(posts))
	}
	if posts[0].ID != 1 || posts[1].ID != 2 || posts[2].ID != 3 {
		t.Fatalf("unexpected order: %+v", posts)
	}
	if page.Pagination.Page != 1 || page.Pagination.PageSize != defaultPostPageSize || page.Pagination.TotalItems != 3 {
		t.Fatalf("unexpected pagination: %+v", page.Pagination)
	}
}

func TestGetPostsByTopicSortsByNew(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	insertTopicAndPostsForSorting(t, owner.ID)

	req := httptest.NewRequest(http.MethodGet, "/topics/1/posts?sort=new", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	GetPostsByTopic(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	posts := decodeSuccessEnvelope[TopicPostsPage](t, recorder).Posts
	if len(posts) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(posts))
	}
	if posts[0].ID != 3 || posts[1].ID != 2 || posts[2].ID != 1 {
		t.Fatalf("unexpected order: %+v", posts)
	}
}

func TestGetPostsByTopicTopSortUsesDeterministicTieBreakers(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	voter, _ := signupAndSessionCookie(t, "voter")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by, created_at)
		VALUES
			(1, 'Post A', 'Body A', ?, '2026-04-06 10:00:00'),
			(1, 'Post B', 'Body B', ?, '2026-04-06 10:00:00')
	`, owner.ID, owner.ID); err != nil {
		t.Fatalf("insert posts: %v", err)
	}

	if _, err := db().Exec(`INSERT INTO votes (user_id, post_id, vote_value) VALUES (?, 1, 1), (?, 2, 1)`, voter.ID, owner.ID); err != nil {
		t.Fatalf("insert votes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/topics/1/posts?sort=top", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	GetPostsByTopic(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	posts := decodeSuccessEnvelope[TopicPostsPage](t, recorder).Posts
	if posts[0].ID != 2 || posts[1].ID != 1 {
		t.Fatalf("expected id DESC tie-break, got %+v", posts)
	}
}

func TestGetPostsByTopicRejectsInvalidSort(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/topics/1/posts?sort=hot", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	GetPostsByTopic(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "invalid_sort" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestGetPostsByTopicPaginatesResultsAcrossPages(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	insertTopicAndPostsForSorting(t, owner.ID)

	req := httptest.NewRequest(http.MethodGet, "/topics/1/posts?sort=new&page=2&pageSize=1", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	GetPostsByTopic(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	page := decodeSuccessEnvelope[TopicPostsPage](t, recorder)
	if len(page.Posts) != 1 || page.Posts[0].ID != 2 {
		t.Fatalf("unexpected page posts: %+v", page.Posts)
	}
	if page.Pagination.Page != 2 || page.Pagination.PageSize != 1 || page.Pagination.TotalItems != 3 || page.Pagination.TotalPages != 3 {
		t.Fatalf("unexpected pagination: %+v", page.Pagination)
	}
	if !page.Pagination.HasPrev || !page.Pagination.HasNext {
		t.Fatalf("expected prev and next to be true: %+v", page.Pagination)
	}
}

func TestGetPostsByTopicPaginationPreservesSortMode(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")
	voterA, _ := signupAndSessionCookie(t, "voterA")
	voterB, _ := signupAndSessionCookie(t, "voterB")

	insertTopicAndPostsForSorting(t, owner.ID)

	if _, err := db().Exec(`INSERT INTO votes (user_id, post_id, vote_value) VALUES (?, 1, 1), (?, 1, 1), (?, 2, 1)`, voterA.ID, voterB.ID, voterA.ID); err != nil {
		t.Fatalf("insert votes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/topics/1/posts?sort=top&page=2&pageSize=1", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	GetPostsByTopic(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	page := decodeSuccessEnvelope[TopicPostsPage](t, recorder)
	if len(page.Posts) != 1 || page.Posts[0].ID != 2 {
		t.Fatalf("unexpected paginated top sort result: %+v", page.Posts)
	}
}

func TestGetPostsByTopicRejectsInvalidPage(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/topics/1/posts?page=0", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	GetPostsByTopic(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "invalid_pagination" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func TestGetPostsByTopicRejectsInvalidPageSize(t *testing.T) {
	setupTestDB(t)
	owner, _ := signupAndSessionCookie(t, "owner")

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, owner.ID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/topics/1/posts?pageSize=100", nil)
	req.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	GetPostsByTopic(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	errEnvelope := decodeErrorEnvelope(t, recorder)
	if errEnvelope.Error.Code != "invalid_pagination" {
		t.Fatalf("unexpected error: %+v", errEnvelope)
	}
}

func insertTopicAndPostsForSorting(t *testing.T, ownerID int) {
	t.Helper()

	if _, err := db().Exec(`
		INSERT INTO topics (title, description, created_by)
		VALUES ('Markets', 'Research discussion', ?)
	`, ownerID); err != nil {
		t.Fatalf("insert topic: %v", err)
	}

	if _, err := db().Exec(`
		INSERT INTO posts (topic_id, title, body, created_by, created_at)
		VALUES
			(1, 'Oldest', 'Body 1', ?, '2026-04-06 08:00:00'),
			(1, 'Middle', 'Body 2', ?, '2026-04-06 09:00:00'),
			(1, 'Newest', 'Body 3', ?, '2026-04-06 10:00:00')
	`, ownerID, ownerID, ownerID); err != nil {
		t.Fatalf("insert posts: %v", err)
	}
}
