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
	const dataDir = "data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.Mkdir(dataDir, 0o755); err != nil {
			return err
		}
	}

	dbPath := "data/campuscommons.db"
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	if err := configureSQLite(DB); err != nil {
		return err
	}
	if err := DB.Ping(); err != nil {
		return err
	}
	if err := CreateSchema(DB); err != nil {
		return err
	}

	log.Println("Connected to SQLite database at", dbPath)
	return nil
}

func configureSQLite(db *sql.DB) error {
	pragmas := []string{
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA busy_timeout = 5000;`,
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return err
		}
	}

	return nil
}

func CreateSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE CHECK(length(trim(username)) > 0),
			password_hash TEXT NOT NULL CHECK(length(trim(password_hash)) > 0),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token_hash TEXT NOT NULL UNIQUE CHECK(length(trim(token_hash)) > 0),
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS companies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticker TEXT NOT NULL UNIQUE CHECK(length(trim(ticker)) > 0 AND ticker = upper(trim(ticker))),
			name TEXT NOT NULL CHECK(length(trim(name)) > 0),
			description TEXT NOT NULL DEFAULT '',
			created_by INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			company_id INTEGER NOT NULL,
			title TEXT NOT NULL CHECK(length(trim(title)) > 0),
			content TEXT NOT NULL CHECK(length(trim(content)) > 0),
			created_by INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(company_id) REFERENCES companies(id) ON DELETE CASCADE,
			FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			content TEXT NOT NULL CHECK(length(trim(content)) > 0),
			created_by INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE IF NOT EXISTS votes (
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			vote_value INTEGER NOT NULL CHECK(vote_value IN (-1, 1)),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(user_id, post_id),
			FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);`,
		`CREATE INDEX IF NOT EXISTS idx_companies_created_by ON companies(created_by);`,
		`CREATE INDEX IF NOT EXISTS idx_companies_ticker ON companies(ticker);`,
		`CREATE INDEX IF NOT EXISTS idx_companies_created_at ON companies(created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_company_id ON posts(company_id);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_created_by ON posts(created_by);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_company_created_at ON posts(company_id, created_at DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post_created_at ON comments(post_id, created_at ASC, id ASC);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_created_by ON comments(created_by);`,
		`CREATE INDEX IF NOT EXISTS idx_votes_post_id ON votes(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_votes_user_id ON votes(user_id);`,
	}

	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Database schema initialized")
	return nil
}

func tableColumns(db schemaQueryer, tableName string) (map[string]bool, error) {
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

type schemaQueryer interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}
