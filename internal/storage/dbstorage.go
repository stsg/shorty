package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"

	"github.com/stsg/shorty/internal/config"
)

const uniqueViolation = pq.ErrorCode("23505")

// URLDeleted error is returned when a URL is deleted.
var ErrURLDeleted = errors.New("URL deleted")

// DBStorage is a struct that holds DB storage data.
type DBStorage struct {
	db *sql.DB
}

// NewDBStorage initializes a new DBStorage object to DB instance based on the provided config.
//
// Parameter:
// - config: config.Config - the configuration settings for the database storage.
// Returns:
// - *DBStorage: the initialized DBStorage object.
// - error: an error if any occurs during the initialization process.
func NewDBStorage(config config.Config) (*DBStorage, error) {
	db, err := sql.Open("postgres", config.GetDBStorage())
	if err != nil {
		return nil, fmt.Errorf("DB open error: %s", err)
	}

	if !IsTableExist(db, "urls") {
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create migration driver: %s", err)
		}
		m, err := migrate.NewWithDatabaseInstance(
			"file://data/db/migration",
			"postgres", driver,
		)
		if err != nil {
			return nil, fmt.Errorf("DB migrate registration error: %s", err)
		}
		err = m.Up()
		if err != nil {
			return nil, fmt.Errorf("DB migration error: %s", err)
		}
	}

	return &DBStorage{db: db}, nil
}

// Save saves a short URL and its corresponding long URL to the database for a given user.
//
// Parameters:
// - userID: the ID of the user for whom the URL is being saved.
// - shortURL: the shortened URL.
// - longURL: the original URL.
//
// Returns:
// - error: an error if there was a problem saving the URL to the database.
func (s *DBStorage) Save(userID uint64, shortURL string, longURL string) error {
	var dbErr *pq.Error

	query := "INSERT INTO urls(short_url, original_url, user_id, deleted) VALUES ($1, $2, $3, $4)"
	_, err := s.db.Exec(query, shortURL, longURL, userID, false)
	if err != nil {
		if errors.As(err, &dbErr) && dbErr.Code == uniqueViolation {
			return ErrUniqueViolation
		}
		return err
	}

	return nil
}

// SaveNew saves a new short URL in the database.
//
// It takes the following parameters:
// - userID: an unsigned 64-bit integer representing the ID of the user.
// - shortURL: a string representing the short URL.
// - longURL: a string representing the long URL.
//
// It returns an error if there was an issue saving the short URL.
func (s *DBStorage) SaveNew(userID uint64, shortURL string, longURL string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.New("cannot start transaction when saving new short URL")
	}
	defer tx.Rollback()

	if !s.IsShortURLExist(shortURL) {
		err = s.Save(userID, shortURL, longURL)
		if err == nil {
			if err = tx.Commit(); err != nil {
				return errors.New("cannot commit transaction when saving new short URL")
			}
			return nil
		}
		return errors.New("cannot save new short URL")
	}

	return ErrUniqueViolation
}

// GetRealURL retrieves the original URL associated with the provided short URL.
//
// Parameter: shortURL string
// Returns: string, error
func (s *DBStorage) GetRealURL(shortURL string) (string, error) {
	var longURL string
	var deleted bool

	query := "SELECT original_url, deleted FROM urls WHERE short_url = $1"
	err := s.db.QueryRow(query, shortURL).Scan(&longURL, &deleted)
	if deleted {
		return "", ErrURLDeleted
	}
	if err != nil {
		return "", err
	}
	return longURL, nil
}

// GetShortURLBatch retrieves short URLs for a batch of long URLs.
//
// Parameters:
// - userID: The ID of the user.
// - bAddr: The base address for the short URLs.
// - longURLs: A slice of ReqJSONBatch containing the long URLs.
//
// Returns:
// - rwJSON: A slice of ResJSONBatch containing the short URLs.
// - error: An error if any occurred during the retrieval process.
func (s *DBStorage) GetShortURLBatch(userID uint64, bAddr string, longURLs []ReqJSONBatch) ([]ResJSONBatch, error) {
	var rwJSON []ResJSONBatch

	tx, err := s.db.Begin()
	if err != nil {
		return rwJSON, errors.New("cannot start transaction when saving new short URL")
	}
	defer tx.Rollback()

	for _, rqElemJSON := range longURLs {
		shortURL, err := s.GetShortURL(userID, rqElemJSON.URL)
		shortURL = bAddr + "/" + shortURL
		rwElemJSON := ResJSONBatch{
			ID:     rqElemJSON.ID,
			Result: shortURL,
		}
		if err != nil {
			rwElemJSON.Result = err.Error()
		}
		rwJSON = append(rwJSON, rwElemJSON)
	}
	if err = tx.Commit(); err != nil {
		return rwJSON, errors.New("cannot commit transaction when saving new short URL")
	}
	return rwJSON, nil
}

// GetShortURL retrieves or generates a short URL for the given long URL and user ID.
//
// Parameters:
// - userID uint64: the user ID associated with the URL.
// - longURL string: the long URL to generate a short URL for.
// Return type(s): string, error
func (s *DBStorage) GetShortURL(userID uint64, longURL string) (string, error) {
	var shortURL string
	query := "SELECT short_url FROM urls WHERE original_url = $1"
	err := s.db.QueryRow(query, longURL).Scan(&shortURL)
	if !errors.Is(err, sql.ErrNoRows) {
		return shortURL, ErrUniqueViolation
	}

	shortURL = GenShortURL()

	for {
		err = s.SaveNew(userID, shortURL, longURL)
		if err == nil {
			return shortURL, nil
		}
		if !errors.Is(err, ErrUniqueViolation) {
			return "", err
		}
		shortURL = GenShortURL()
	}
}

// IsShortURLExist checks if the short URL exists in the DBStorage.
//
// shortURL string
// bool
func (s *DBStorage) IsShortURLExist(shortURL string) bool {
	var longURL string
	query := "SELECT original_url FROM urls WHERE short_url = $1"
	err := s.db.QueryRow(query, shortURL).Scan(&longURL)
	return !errors.Is(err, sql.ErrNoRows)
}

// IsRealURLExist checks if the given long URL exists in the database.
//
// longURL string
// bool
func (s *DBStorage) IsRealURLExist(longURL string) bool {
	var shortURL string
	query := "SELECT short_url FROM urls WHERE original_url = $1"
	err := s.db.QueryRow(query, longURL).Scan(&shortURL)
	return !errors.Is(err, sql.ErrNoRows)
}

// IsReady checks if the DBStorage is ready.
//
// No parameters.
// Returns a boolean value.
func (s *DBStorage) IsReady() bool {
	err := s.db.Ping()
	return err == nil
}

// IsTableExist checks if a table exists in the database.
//
// It takes a *sql.DB object and a string representing the table name as parameters.
// It returns a boolean value indicating whether the table exists or not.
func IsTableExist(db *sql.DB, table string) bool {
	var n int64
	query := "SELECT 1 FROM information_schema.tables WHERE table_name = $1"
	err := db.QueryRow(query, table).Scan(&n)
	return err == nil
}

// GetAllURLs retrieves all URLs for a given user and base address.
//
// userID uint64, bAddr string
// []ResJSONURL, error
func (s *DBStorage) GetAllURLs(userID uint64, bAddr string) ([]ResJSONURL, error) {
	var rwJSON []ResJSONURL
	query := "SELECT short_url, original_url FROM urls WHERE user_id = $1"
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL, longURL string
		if err := rows.Scan(&shortURL, &longURL); err != nil {
			return nil, err
		}
		shortURL = bAddr + "/" + shortURL
		rwJSON = append(rwJSON, ResJSONURL{
			URL:    longURL,
			Result: shortURL,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rwJSON, nil
}

// GetLastID retrieves the last ID from the "urls" table in the database.
//
// It returns an integer representing the last ID and an error if any occurred.
func (s *DBStorage) GetLastID() (int, error) {
	var lastID sql.NullInt64
	query := "SELECT MAX(uuid) FROM urls"
	err := s.db.QueryRow(query).Scan(&lastID)
	if err != nil {
		return 0, err
	}
	if !lastID.Valid {
		return 0, nil
	}
	return int(lastID.Int64), nil
}

// DeleteURLs deletes URLs associated with a specific user.
//
// userID: the ID of the user whose URLs are being deleted.
// delURLs: a slice of strings containing the URLs to be deleted.
// error: an error indicating any issues that occurred during the deletion process.
func (s *DBStorage) DeleteURLs(userID uint64, delURLs []string) error {
	for _, i := range delURLs {
		query := "UPDATE urls SET deleted = true WHERE short_url = $1 and user_id = $2"
		_, err := s.db.Exec(query, i, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteURL updates the "deleted" field in the "urls" table for the given short URLs and user IDs.
//
// It takes a map of short URLs to user IDs as input. The function iterates over the map and
// executes an SQL query to update the "deleted" field to true for each short URL and user ID
// combination. If any error occurs during the execution of the query, it is returned.
//
// The function returns an error if there was an error executing the SQL query, otherwise it
// returns nil.
func (s *DBStorage) DeleteURL(delURL map[string]uint64) error {
	query := "UPDATE urls SET deleted = true WHERE short_url = $1 and user_id = $2"
	for sURL, userID := range delURL {
		_, err := s.db.Exec(query, sURL, userID)
		if err != nil {
			return err
		}
	}

	return nil
}
