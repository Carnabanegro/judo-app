package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open opens (or creates) the SQLite database at the given path.
// Pass ":memory:" for an in-memory database.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open %q: %w", path, err)
	}
	// SQLite performs best with a single writer connection.
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: migrate: %w", err)
	}
	return db, nil
}

// migrate applies the schema in a single idempotent transaction.
func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA foreign_keys=ON;

		CREATE TABLE IF NOT EXISTS tournaments (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			location    TEXT NOT NULL,
			date        TEXT NOT NULL,
			created_at  TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS divisions (
			id            TEXT PRIMARY KEY,
			tournament_id TEXT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
			age_group     TEXT NOT NULL,
			gender        TEXT NOT NULL,
			weight_class  TEXT NOT NULL,
			format        TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS categories (
			id          TEXT PRIMARY KEY,
			division_id TEXT NOT NULL REFERENCES divisions(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS athletes (
			id          TEXT PRIMARY KEY,
			category_id TEXT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			name        TEXT NOT NULL,
			club        TEXT NOT NULL,
			weight      REAL NOT NULL,
			birth_date  TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS brackets (
			category_id TEXT PRIMARY KEY REFERENCES categories(id) ON DELETE CASCADE,
			data        TEXT NOT NULL  -- JSON blob of the full Bracket struct
		);

		-- matches: mutable match state extracted for efficient querying.
		-- The bracket JSON still stores structure; this table owns status/tatami.
		CREATE TABLE IF NOT EXISTS matches (
			id           TEXT PRIMARY KEY,
			category_id  TEXT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			round        INTEGER NOT NULL,
			position     INTEGER NOT NULL,
			is_repechage INTEGER NOT NULL DEFAULT 0,
			athlete_a_id TEXT,
			athlete_b_id TEXT,
			status       TEXT NOT NULL DEFAULT 'PENDING',
			tatami_id    TEXT NOT NULL DEFAULT '',
			result_json  TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_matches_category ON matches(category_id);
		CREATE INDEX IF NOT EXISTS idx_matches_status   ON matches(status);
		CREATE INDEX IF NOT EXISTS idx_divisions_tournament ON divisions(tournament_id);
		CREATE INDEX IF NOT EXISTS idx_categories_division  ON categories(division_id);
		CREATE INDEX IF NOT EXISTS idx_athletes_category    ON athletes(category_id);
	`)
	return err
}
