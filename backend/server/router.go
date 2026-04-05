package server

import (
	"net/http"
	"os"

	"github.com/archonward/CampusCommons/backend/handlers"
	"github.com/rs/cors"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/login", handlers.Login)
	mux.HandleFunc("/topics", handlers.TopicsCollection)
	mux.HandleFunc("/topics/{id}", handlers.TopicResource)
	mux.HandleFunc("/topics/{id}/posts", handlers.TopicPostsResource)
	mux.HandleFunc("/posts/{id}", handlers.PostResource)
	mux.HandleFunc("/posts/{id}/comments", handlers.PostCommentsResource)

	allowedOrigin := os.Getenv("CAMPUSCOMMONS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{allowedOrigin},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-User-ID"},
	})

	return c.Handler(mux)
}
