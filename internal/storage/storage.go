package storage

import (
	"errors"
	"math/rand"

	"github.com/stsg/shorty/internal/config"
)

type Storage interface {
	Save(shortURL string, longURL string) error
	GetRealURL(shortURL string) (string, error)
	GetShortURL(longURL string) (string, error)
	IsRealURLExist(longURL string) bool
	IsShortURLExist(longURL string) bool
	IsReady() error
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
	f, err := NewFileStorage(conf)
	if err != nil {
		return nil, errors.New("cannot create storage")
	}
	return f, nil
}
