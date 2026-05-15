package server

import (
	"net/http"
	"os"
	"strings"

	"github.com/archonward/ArtHub/backend/handlers"
	"github.com/rs/cors"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ArtHub backend is live!"))
	})
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

	allowedOrigins := allowedOriginsFromEnv()

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return c.Handler(mux)
}

func allowedOriginsFromEnv() []string {
	value := os.Getenv("ARTHUB_ALLOWED_ORIGINS")
	if value == "" {
		value = os.Getenv("ARTHUB_ALLOWED_ORIGIN")
	}
	if value == "" {
		value = os.Getenv("CAMPUSCOMMONS_ALLOWED_ORIGIN")
	}
	if value == "" {
		return []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}

	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}
	return origins
}
