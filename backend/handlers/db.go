package handlers

import (
	"database/sql"

	"github.com/archonward/ArtHub/backend/database"
)

func db() *sql.DB {
	return database.DB
}
