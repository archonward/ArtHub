package handlers

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Company struct {
	ID          int       `json:"id"`
	Ticker      string    `json:"ticker"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedBy   int       `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Post struct {
	ID              int       `json:"id"`
	CompanyID       int       `json:"company_id"`
	Title           string    `json:"title"`
	Body            string    `json:"body"`
	CreatedBy       int       `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	VoteScore       int       `json:"vote_score"`
	CurrentUserVote *int      `json:"current_user_vote"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	Body      string    `json:"body"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
