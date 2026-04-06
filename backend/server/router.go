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
	mux.HandleFunc("/auth/signup", handlers.Signup)
	mux.HandleFunc("/auth/login", handlers.Login)
	mux.HandleFunc("/auth/logout", handlers.Logout)
	mux.HandleFunc("/auth/me", handlers.CurrentSessionUser)
	mux.HandleFunc("/companies", handlers.OptionalSessionAuth(handlers.CompaniesCollection))
	mux.HandleFunc("/companies/{id}", handlers.OptionalSessionAuth(handlers.CompanyResource))
	mux.HandleFunc("/companies/{id}/posts", handlers.OptionalSessionAuth(handlers.CompanyPostsResource))
	mux.HandleFunc("/posts/{id}", handlers.OptionalSessionAuth(handlers.PostResource))
	mux.HandleFunc("/posts/{id}/comments", handlers.OptionalSessionAuth(handlers.PostCommentsResource))
	mux.HandleFunc("/posts/{id}/vote", handlers.OptionalSessionAuth(handlers.PostVoteResource))

	allowedOrigin := os.Getenv("CAMPUSCOMMONS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return c.Handler(mux)
}
