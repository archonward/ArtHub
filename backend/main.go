package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/archonward/ArtHub/backend/database"
	"github.com/archonward/ArtHub/backend/server"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	fmt.Printf("Server starting on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, server.NewHandler()))
}
