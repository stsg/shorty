package storage

import (
	"errors"
)

const ShortURLLength = 6

type MapStorage struct {
	m map[string]string
}

func NewMapStorage() *MapStorage {
	return &MapStorage{m: make(map[string]string)}
}

func (s MapStorage) Save(shortURL string, longURL string) error {
	_, exist := s.m[shortURL]
	if exist {
		return errors.New("short URL already exist")
	}
	s.m[shortURL] = longURL
	return nil
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
			return surl, errors.New("short URL already exist: ")
		}
	}
	for {
		surl := GenShortURL()
		_, exist := s.m[surl]
		if !exist {
			s.Save(surl, longURL)
			return surl, nil
		}
	}
}

func (s *MapStorage) IsShortURLExist(shortURL string) bool {
	for surl := range s.m {
		if surl == shortURL {
			return true
		}
	}
	return false
}

func (s *MapStorage) IsRealURLExist(longURL string) bool {
	for _, lurl := range s.m {
		if lurl == longURL {
			return true
		}
	}
	return false
}
