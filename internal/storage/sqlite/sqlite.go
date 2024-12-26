package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"url-shortener/internal/storage"

	_ "github.com/mattn/go-sqlite3" // init sqlite3 driver
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("#{op}: #{err}")
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(alias string, urlToSave string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(alias, url) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(alias, urlToSave)
	if err != nil {
		//TODO:	refactor this shit
		if sqLiteErr, ok := err.(sqlite3.Error); ok && sqLiteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last inset id: %w", op, err)
	}

	return id, nil
}

// Кратко говоря о том, что я делаю
func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("#{op}: prepare statement: #{err}")
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}

	if err != nil {
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) (string, error) {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("delete from url where alias = ?")
	if err != nil {
		return "", fmt.Errorf("#{op}: prepare statement: #{err}")
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return fmt.Sprintf("DELETE FROM url WHERE alias = ?", alias), nil
}