package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupSchemaDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := configureSQLite(db); err != nil {
		t.Fatalf("configure sqlite: %v", err)
	}

	if err := CreateSchema(db); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func assertNoForeignKeyViolations(t *testing.T, db *sql.DB) {
	t.Helper()

	rows, err := db.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		t.Fatalf("foreign_key_check query: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		var table string
		var rowID int64
		var parent string
		var fkID int
		if err := rows.Scan(&table, &rowID, &parent, &fkID); err != nil {
			t.Fatalf("scan foreign key violation: %v", err)
		}
		t.Fatalf("foreign key violation in %s row %d referencing %s (fk %d)", table, rowID, parent, fkID)
	}
}

func TestCreateSchemaCreatesFreshCompanyCentricTables(t *testing.T) {
	db := setupSchemaDB(t)

	expectedColumns := map[string][]string{
		"users":     {"id", "username", "password_hash", "created_at"},
		"sessions":  {"id", "user_id", "token_hash", "expires_at", "created_at", "last_seen_at"},
		"companies": {"id", "ticker", "name", "description", "created_by", "created_at", "updated_at"},
		"posts":     {"id", "company_id", "title", "content", "created_by", "created_at", "updated_at"},
		"comments":  {"id", "post_id", "content", "created_by", "created_at", "updated_at"},
		"votes":     {"post_id", "user_id", "vote_value", "created_at", "updated_at"},
	}

	for tableName, columns := range expectedColumns {
		tableCols, err := tableColumns(db, tableName)
		if err != nil {
			t.Fatalf("tableColumns(%s): %v", tableName, err)
		}
		for _, column := range columns {
			if !tableCols[column] {
				t.Fatalf("expected %s.%s to exist; got columns %+v", tableName, column, tableCols)
			}
		}
	}

	if _, err := db.Exec(`INSERT INTO users (id, username, password_hash) VALUES (1, 'owner', 'hash1')`); err != nil {
		t.Fatalf("insert owner: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO companies (ticker, name, description, created_by) VALUES ('msft', 'Microsoft', '', 1)`); err == nil {
		t.Fatalf("expected lowercase ticker insert to fail check constraint")
	}
}

func TestCompanyDeletionCascadesPostsCommentsAndVotes(t *testing.T) {
	db := setupSchemaDB(t)

	if _, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES
			(1, 'owner', 'hash1'),
			(2, 'voter', 'hash2');
		INSERT INTO companies (id, ticker, name, description, created_by) VALUES
			(1, 'AAPL', 'Apple Inc.', '', 1);
		INSERT INTO posts (id, company_id, title, content, created_by) VALUES
			(1, 1, 'Post', 'Post content', 1);
		INSERT INTO comments (id, post_id, content, created_by) VALUES
			(1, 1, 'Comment content', 1);
		INSERT INTO votes (post_id, user_id, vote_value) VALUES
			(1, 2, 1);
	`); err != nil {
		t.Fatalf("seed relational data: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM companies WHERE id = 1`); err != nil {
		t.Fatalf("delete company: %v", err)
	}

	for _, tableName := range []string{"companies", "posts", "comments", "votes"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM ` + tableName).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", tableName, err)
		}
		if count != 0 {
			t.Fatalf("expected %s to be empty after cascade, got %d rows", tableName, count)
		}
	}

	assertNoForeignKeyViolations(t, db)
}

func TestPostDeletionCascadesCommentsAndVotes(t *testing.T) {
	db := setupSchemaDB(t)

	if _, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES
			(1, 'owner', 'hash1'),
			(2, 'voter', 'hash2');
		INSERT INTO companies (id, ticker, name, description, created_by) VALUES
			(1, 'AAPL', 'Apple Inc.', '', 1);
		INSERT INTO posts (id, company_id, title, content, created_by) VALUES
			(1, 1, 'Post', 'Post content', 1);
		INSERT INTO comments (id, post_id, content, created_by) VALUES
			(1, 1, 'Comment content', 1);
		INSERT INTO votes (post_id, user_id, vote_value) VALUES
			(1, 2, -1);
	`); err != nil {
		t.Fatalf("seed post graph: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM posts WHERE id = 1`); err != nil {
		t.Fatalf("delete post: %v", err)
	}

	for _, tableName := range []string{"comments", "votes"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM ` + tableName).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", tableName, err)
		}
		if count != 0 {
			t.Fatalf("expected %s to be empty after post cascade, got %d rows", tableName, count)
		}
	}

	assertNoForeignKeyViolations(t, db)
}

func TestVoteUniquenessAndUserCascade(t *testing.T) {
	db := setupSchemaDB(t)

	if _, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES
			(1, 'owner', 'hash1'),
			(2, 'voter', 'hash2');
		INSERT INTO companies (id, ticker, name, description, created_by) VALUES
			(1, 'AAPL', 'Apple Inc.', '', 1);
		INSERT INTO posts (id, company_id, title, content, created_by) VALUES
			(1, 1, 'Post', 'Post content', 1);
		INSERT INTO votes (post_id, user_id, vote_value) VALUES
			(1, 2, 1);
	`); err != nil {
		t.Fatalf("seed vote data: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO votes (post_id, user_id, vote_value) VALUES (1, 2, -1)`); err == nil {
		t.Fatalf("expected duplicate vote insert to fail")
	}

	if _, err := db.Exec(`DELETE FROM users WHERE id = 2`); err != nil {
		t.Fatalf("delete voter: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM votes`).Scan(&count); err != nil {
		t.Fatalf("count votes: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected votes to cascade on user delete, got %d", count)
	}
}

func TestSessionCascadesOnUserDelete(t *testing.T) {
	db := setupSchemaDB(t)

	if _, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES (1, 'owner', 'hash1');
		INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (1, 'token-hash', datetime('now', '+1 day'));
	`); err != nil {
		t.Fatalf("seed session data: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM users WHERE id = 1`); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected sessions to cascade on user delete, got %d", count)
	}
}
