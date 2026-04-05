package handlers

import (
	"database/sql"

	"github.com/archonward/CampusCommons/backend/database"
)

func db() *sql.DB {
	return database.DB
}
