package storage

import "errors"

// URLStorage defines the interface for URL storage operations
type URLStorage interface {
	SaveURL(originalURL, shortCode string) error
	GetURL(shortCode string) (string, error)
	GetURLStats(shortCode string) (int, error)
	Close() error
}

// Custom errors
var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
	ErrInvalidURL  = errors.New("invalid url")
)
