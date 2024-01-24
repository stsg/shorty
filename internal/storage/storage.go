package storage

import (
	"errors"
	"math/rand"

	"github.com/stsg/shorty/internal/config"
)

type ReqJSON struct {
	URL string `json:"url,omitempty"`
}

type ResJSON struct {
	Result string `json:"result"`
}

type ReqJSONBatch struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url,omitempty"`
}

type ResJSONBatch struct {
	ID     string `json:"correlation_id"`
	Result string `json:"short_url,omitempty"`
}

//	type handle struct {
//		Config  config.Config
//		storage Storage
//	}
//type reqJSONBatch struct {
//	ID  string `json:"correlation_id"`
//	URL string `json:"original_url,omitempty"`
//}
//
//type resJSONBatch struct {
//	ID     string `json:"correlation_id"`
//	Result string `json:"short_url,omitempty"`
//}

var ErrUniqueViolation = errors.New("short URL already exist")

type Storage interface {
	Save(shortURL string, longURL string) error
	GetRealURL(shortURL string) (string, error)
	GetShortURL(longURL string) (string, error)
	GetShortURLBatch(string, []ReqJSONBatch) ([]ResJSONBatch, error)
	IsRealURLExist(longURL string) bool
	IsShortURLExist(longURL string) bool
	IsReady() bool
}

func GenShortURL() string {
	charset := "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	shortURL := make([]byte, ShortURLLength)
	for i := range shortURL {
		shortURL[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortURL)
}

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
