package storage

import (
	"errors"
)

const ShortURLLength = 6

type MapStorage struct {
	m map[string]string
}

func NewMapStorage() (*MapStorage, error) {
	return &MapStorage{m: make(map[string]string)}, nil
}

func (s *MapStorage) Save(shortURL string, longURL string) error {
	_, exist := s.m[shortURL]
	if exist {
		return ErrUniqueViolation
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
	for sURL, lURL := range s.m {
		if lURL == longURL {
			return sURL, ErrUniqueViolation
		}
	}
	for {
		sURL := GenShortURL()
		_, exist := s.m[sURL]
		if !exist {
			err := s.Save(sURL, longURL)
			if err != nil {
				return "", err
			}
			return sURL, nil
		}
	}
}

func (s *MapStorage) IsShortURLExist(shortURL string) bool {
	_, exist := s.m[shortURL]
	return exist
}

func (s *MapStorage) IsRealURLExist(longURL string) bool {
	for _, lURL := range s.m {
		if lURL == longURL {
			return true
		}
	}
	return false
}

func (s *MapStorage) IsReady() bool {
	return true
}
