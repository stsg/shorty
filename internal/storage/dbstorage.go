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

var ErrURLDeleted = errors.New("URL deleted")

type DBStorage struct {
	db *sql.DB
}

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

func (s *DBStorage) IsShortURLExist(shortURL string) bool {
	var longURL string
	query := "SELECT original_url FROM urls WHERE short_url = $1"
	err := s.db.QueryRow(query, shortURL).Scan(&longURL)
	return !errors.Is(err, sql.ErrNoRows)
}

func (s *DBStorage) IsRealURLExist(longURL string) bool {
	var shortURL string
	query := "SELECT short_url FROM urls WHERE original_url = $1"
	err := s.db.QueryRow(query, longURL).Scan(&shortURL)
	return !errors.Is(err, sql.ErrNoRows)
}

func (s *DBStorage) IsReady() bool {
	err := s.db.Ping()
	return err == nil
}

func IsTableExist(db *sql.DB, table string) bool {
	var n int64
	query := "SELECT 1 FROM information_schema.tables WHERE table_name = $1"
	err := db.QueryRow(query, table).Scan(&n)
	return err == nil
}

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

func (s *DBStorage) DeleteURLs(userID uint64, delURLs []string) error {
	for _, i := range delURLs {
		query := "UPDATE SET deleted = true WHERE short_url = $1 and user_id = $2"
		_, err := s.db.Exec(query, i, userID)
		if err != nil {
			return err
		}
	}

	return nil
}
