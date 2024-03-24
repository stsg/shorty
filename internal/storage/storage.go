package storage

import (
	"errors"
	"math/rand"

	"github.com/stsg/shorty/internal/config"
)

// ReqJSON request JSON for serializing/deserializng URL
type ReqJSON struct {
	URL string `json:"url,omitempty"`
}

// ResJSON result JSON for serializing/deserializng URL
type ResJSON struct {
	Result string `json:"result"`
}

// ReqJSONBatch request JSON for serializing/deserializng URLs batch
type ReqJSONBatch struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url,omitempty"`
}

// ResJSONBatch result JSON for serializing/deserializng URLs batch
type ResJSONBatch struct {
	ID     string `json:"correlation_id"`
	Result string `json:"short_url,omitempty"`
}

// ResJSONURL result JSON for serializing/deserializng URLs list
type ResJSONURL struct {
	Result string `json:"short_url,omitempty"`
	URL    string `json:"original_url,omitempty"`
}

var ShortURLLength = 6

// ErrUniqueViolation is an error that is returned when a short URL already exist.
var ErrUniqueViolation = errors.New("short URL already exist")

// This class definition represents a storage interface in Go. Here's a list explaining what each method does:
//
// Save(userID uint64, shortURL string, longURL string) error: Saves a short URL and its corresponding long URL for a specific user.
// GetRealURL(shortURL string) (string, error): Retrieves the real (long) URL associated with a given short URL.
// GetShortURL(userID uint64, longURL string) (string, error): Retrieves the short URL associated with a given long URL for a specific user.
// GetShortURLBatch(userID uint64, bAddr string, longURLs []ReqJSONBatch) ([]ResJSONBatch, error): Retrieves short URLs in batch for a specific user.
// GetAllURLs(userID uint64, bAddr string) ([]ResJSONURL, error): Retrieves all URLs for a specific user.
// IsRealURLExist(longURL string) bool: Checks if a real (long) URL exists in the storage.
// IsShortURLExist(longURL string) bool: Checks if a short URL exists in the storage.
// IsReady() bool: Checks if the storage is ready.
// GetLastID() (int, error): Retrieves the last ID used.
type Storage interface {
	Save(userID uint64, shortURL string, longURL string) error
	GetRealURL(shortURL string) (string, error)
	GetShortURL(userID uint64, longURL string) (string, error)
	GetShortURLBatch(userID uint64, bAddr string, longURLs []ReqJSONBatch) ([]ResJSONBatch, error)
	GetAllURLs(userID uint64, bAddr string) ([]ResJSONURL, error)
	IsRealURLExist(longURL string) bool
	IsShortURLExist(longURL string) bool
	IsReady() bool
	GetLastID() (int, error)
	DeleteURLs(userID uint64, delURLs []string) error
	DeleteURL(delURL map[string]uint64) error
}

// GenShortURL generates a random short URL of length ShortURLLength using the characters from the charset.
//
// It returns the generated short URL as a string.
func GenShortURL() string {
	charset := "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	if ShortURLLength < 1 {
		return ""
	}

	shortURL := make([]byte, ShortURLLength)
	for i := range shortURL {
		shortURL[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortURL)
}

// New initializes and returns a Storage based on the provided configuration.
//
// Parameter:
//
//	conf - config.Config
//
// Return:
//
//	Storage - the initialized storage object
//	error - an error if the storage creation fails
func New(conf config.Config) (Storage, error) {
	if conf.GetStorageType() == "file" {
		storage, err := NewFileStorage(conf)
		if err != nil {
			return nil, errors.New("cannot create file storage")
		}
		return storage, nil
	}

	if conf.GetStorageType() == "db" {
		storage, err := NewDBStorage(conf)
		if err != nil {
			return nil, errors.New("cannot create DB storage")
		}
		return storage, nil
	}

	storage, _ := NewMapStorage()
	return storage, nil
}
