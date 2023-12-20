package storage

import (
	"errors"
	"math/rand"
)

const ShortURLLength = 6

type MapStorage struct {
	m map[string]string
}

func NewMapStorage() *MapStorage {
	return &MapStorage{m: make(map[string]string)}
}

func (s *MapStorage) GetRealURL(shortURL string) (string, error) {
	if len(shortURL) > ShortURLLength {
		return "", errors.New("short URL longer than ShortURLLength")
	}
	longURL, exist := s.m[shortURL]
	if !exist {
		return "", errors.New("short URL not exist")
	}
	return longURL, nil
}

func (s *MapStorage) GetShortURL(longURL string) (string, error) {
	for surl, lurl := range s.m {
		if lurl == longURL {
			return surl, errors.New("short URL already exist")
		}
	}
	for {
		surl := genShortURL()
		_, exist := s.m[surl]
		if !exist {
			return surl, nil
		}
	}
}

func genShortURL() string {
	charset := "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	shortURL := make([]byte, ShortURLLength)
	for i := range shortURL {
		shortURL[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortURL)
}
