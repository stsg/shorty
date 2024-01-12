package storage

import (
	"database/sql"
	"github.com/stsg/shorty/internal/config"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(config config.Config) (*DBStorage, error) {
	db, err := sql.Open("postgres", config.GetDBStor())
	if err != nil {
		return nil, err
	}

	defer db.Close()
	return &DBStorage{db: db}, nil
}
func (s *DBStorage) IsReady() error {
	err := s.db.Ping()
	if err != nil {
		return err
	}
	return nil
}
