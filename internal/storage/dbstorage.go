package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/stsg/shorty/internal/config"
)

const uniqueViolation = pq.ErrorCode("23505")

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

func (s *DBStorage) Save(shortURL string, longURL string) error {
	var dbErr *pq.Error

	query := "INSERT INTO urls(short_url, original_url) VALUES ($1, $2)"
	_, err := s.db.Exec(query, shortURL, longURL)
	if err != nil {
		if errors.As(err, &dbErr) && dbErr.Code == uniqueViolation {
			return ErrUniqueViolation
		}
		return err
	}

	return nil
}

func (s *DBStorage) GetRealURL(shortURL string) (string, error) {
	var longURL string
	query := "SELECT original_url FROM urls WHERE short_url = $1"
	err := s.db.QueryRow(query, shortURL).Scan(&longURL)
	if err != nil {
		return "", err
	}
	return longURL, nil
}

func (s *DBStorage) GetShortURL(longURL string) (string, error) {
	var shortURL string
	query := "SELECT short_url FROM urls WHERE original_url = $1"
	err := s.db.QueryRow(query, longURL).Scan(&shortURL)
	if !errors.Is(err, sql.ErrNoRows) {
		return shortURL, ErrUniqueViolation
	}

	shortURL = GenShortURL()

	for {
		if !s.IsShortURLExist(shortURL) {
			err = s.Save(shortURL, longURL)
			if err == nil {
				return shortURL, nil
			} else {
				return "", err
			}
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
