package storage

import (
	"errors"
)

const ShortURLLength = 6

type MapStorage struct {
	m map[string]UserURL
}

type UserURL struct {
	LongURL string
	UserID  uint64
}

func NewMapStorage() (*MapStorage, error) {
	return &MapStorage{m: make(map[string]UserURL)}, nil
}

func (s *MapStorage) Save(userID uint64, shortURL string, longURL string) error {
	_, exist := s.m[shortURL]
	if exist {
		return ErrUniqueViolation
	}
	uURL := UserURL{
		LongURL: longURL,
		UserID:  userID,
	}
	s.m[shortURL] = uURL
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
	return longURL.LongURL, nil
}

func (s *MapStorage) GetShortURLBatch(userID uint64, bAddr string, longURLs []ReqJSONBatch) ([]ResJSONBatch, error) {
	var rwJSON []ResJSONBatch
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
	return rwJSON, nil
}
func (s *MapStorage) GetShortURL(userID uint64, longURL string) (string, error) {
	for sURL, lURL := range s.m {
		if lURL.LongURL == longURL {
			return sURL, ErrUniqueViolation
		}
	}
	for {
		sURL := GenShortURL()
		_, exist := s.m[sURL]
		if !exist {
			err := s.Save(userID, sURL, longURL)
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
		if lURL.LongURL == longURL {
			return true
		}
	}
	return false
}

func (s *MapStorage) IsReady() bool {
	return true
}

func (s *MapStorage) GetAllURLs(userID uint64, bAddr string) ([]ResJSONURL, error) {
	var rwJSON []ResJSONURL
	for sURL, lURL := range s.m {
		if lURL.UserID == userID {
			rwElemJSON := ResJSONURL{
				URL:    lURL.LongURL,
				Result: bAddr + "/" + sURL,
			}
			rwJSON = append(rwJSON, rwElemJSON)
		}
	}
	return rwJSON, nil
}

func (s *MapStorage) GetLastID() (int, error) {
	return len(s.m), nil
}
