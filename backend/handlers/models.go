package handlers

import "time"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type Topic struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	CreatedBy   int       `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Post struct {
	ID        int       `json:"id"`
	TopicID   int       `json:"topic_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	Body      string    `json:"body"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}
