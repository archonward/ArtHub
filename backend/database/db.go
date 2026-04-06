package database

import (
	"database/sql"
	"log"
	"os"
	"strings"

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

	if err := DB.Ping(); err != nil { // Test the connection
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
		username TEXT UNIQUE NOT NULL CHECK(length(trim(username)) > 0),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL CHECK(length(trim(title)) > 0),
		description TEXT,
		created_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		topic_id INTEGER NOT NULL,
		title TEXT NOT NULL CHECK(length(trim(title)) > 0),
		body TEXT NOT NULL CHECK(length(trim(body)) > 0),
		created_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(topic_id) REFERENCES topics(id) ON DELETE CASCADE,
		FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		body TEXT NOT NULL CHECK(length(trim(body)) > 0),
		created_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS votes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		vote_value INTEGER NOT NULL CHECK(vote_value IN (-1, 1)),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, post_id),
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_topics_created_by ON topics(created_by);
	CREATE INDEX IF NOT EXISTS idx_topics_created_at ON topics(created_at);
	CREATE INDEX IF NOT EXISTS idx_posts_topic_id ON posts(topic_id);
	CREATE INDEX IF NOT EXISTS idx_posts_created_by ON posts(created_by);
	CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
	CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
	CREATE INDEX IF NOT EXISTS idx_comments_created_by ON comments(created_by);
	CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);
	CREATE INDEX IF NOT EXISTS idx_votes_post_id ON votes(post_id);
	CREATE INDEX IF NOT EXISTS idx_votes_user_id ON votes(user_id);

	CREATE TRIGGER IF NOT EXISTS trg_topics_delete_posts
	AFTER DELETE ON topics
	BEGIN
		DELETE FROM posts WHERE topic_id = OLD.id;
	END;

	CREATE TRIGGER IF NOT EXISTS trg_posts_delete_comments
	AFTER DELETE ON posts
	BEGIN
		DELETE FROM comments WHERE post_id = OLD.id;
	END;

	CREATE TRIGGER IF NOT EXISTS trg_votes_set_updated_at
	AFTER UPDATE ON votes
	BEGIN
		UPDATE votes
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = NEW.id;
	END;
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	if err := ensureUserAuthColumns(db); err != nil {
		return err
	}

	if err := ensureSessionsTable(db); err != nil {
		return err
	}

	log.Println("Tables created, if they didn't exist")
	return nil
}

func ensureUserAuthColumns(db *sql.DB) error {
	columns, err := tableColumns(db, "users")
	if err != nil {
		return err
	}

	if !columns["password_hash"] {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN password_hash TEXT`); err != nil {
			return err
		}
	}

	return nil
}

func ensureSessionsTable(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token_hash TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`

	_, err := db.Exec(schema)
	return err
}

func tableColumns(db *sql.DB, tableName string) (map[string]bool, error) {
	rows, err := db.Query(`PRAGMA table_info(` + tableName + `)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns[strings.ToLower(name)] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}
