package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var DB *sql.DB

func InitDB() error {
	// Create data directory if it doesn't exist
	const dataDir = "data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.Mkdir(dataDir, 0755)
		if err != nil {
			return err
		}
	}

	// Open SQLite database file (will be created if it doesn't exist)
	dbPath := "data/campuscommons.db"
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err := DB.Ping(); err != nil {			// Test the connection
		return err
	}

	log.Println("Connected to SQLite database at", dbPath)

	if err := enableForeignKeys(DB); err != nil {
		return err
	}

	if err := CreateSchema(DB); err != nil {
		return err
	}

	return nil
}

func enableForeignKeys(db *sql.DB) error {
	_, err := db.Exec(`PRAGMA foreign_keys = ON;`)
	return err
}

func CreateSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		created_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(created_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		topic_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT NOT NULL,
		created_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(topic_id) REFERENCES topics(id),
		FOREIGN KEY(created_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		body TEXT NOT NULL,
		created_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(post_id) REFERENCES posts(id),
		FOREIGN KEY(created_by) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_topics_created_by ON topics(created_by);
	CREATE INDEX IF NOT EXISTS idx_posts_topic_id ON posts(topic_id);
	CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	log.Println("Tables created, if they didn't exist")
	return nil
}
