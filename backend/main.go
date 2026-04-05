package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/archonward/CampusCommons/backend/database"
	"github.com/archonward/CampusCommons/backend/server"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, server.NewHandler()))
}
