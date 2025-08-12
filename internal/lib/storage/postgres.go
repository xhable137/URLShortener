package storage

import (
	"URLShortener/internal/config"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(cfg *config.Config) (*PostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.Port,
		cfg.PostgreSQL.User,
		cfg.PostgreSQL.Password,
		cfg.PostgreSQL.DBName,
		cfg.PostgreSQL.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	slog.Info("Connected to PostgreSQL",
		slog.String("host", cfg.PostgreSQL.Host),
		slog.String("port", cfg.PostgreSQL.Port),
		slog.String("database", cfg.PostgreSQL.DBName),
	)

	return &PostgresStorage{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		original_url TEXT NOT NULL,
		short_code VARCHAR(10) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		clicks INTEGER DEFAULT 0
	);
	`

	_, err := db.Exec(query)
	return err
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// URL methods
func (s *PostgresStorage) SaveURL(originalURL, shortCode string) error {
	query := `INSERT INTO urls (original_url, short_code) VALUES ($1, $2)`
	_, err := s.db.Exec(query, originalURL, shortCode)
	return err
}

func (s *PostgresStorage) GetURL(shortCode string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_code = $1`
	var originalURL string
	err := s.db.QueryRow(query, shortCode).Scan(&originalURL)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return "", ErrURLNotFound
		}
		return "", err
	}

	// Increment clicks
	updateQuery := `UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1`
	if _, err := s.db.Exec(updateQuery, shortCode); err != nil {
		// Log error but don't fail the request
		slog.Error("failed to increment clicks", "error", err, "short_code", shortCode)
	}

	return originalURL, nil
}

func (s *PostgresStorage) GetURLStats(shortCode string) (int, error) {
	query := `SELECT clicks FROM urls WHERE short_code = $1`
	var clicks int
	err := s.db.QueryRow(query, shortCode).Scan(&clicks)
	return clicks, err
}
