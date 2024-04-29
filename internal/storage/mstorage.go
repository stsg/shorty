package storage

import (
	"errors"
)

// ShortURL length

// MapStorage is a struct that holds memory storage data.
type MapStorage struct {
	m map[string]UserURL
}

// UserURL is a struct that holds user URL data.
type UserURL struct {
	LongURL string
	UserID  uint64
}

// NewMapStorage initializes and returns a new instance of MapStorage.
//
// It creates a new map[string]UserURL and assigns it to the m field of the MapStorage struct.
// The function returns a pointer to the newly created MapStorage instance and a nil error.
func NewMapStorage() (*MapStorage, error) {
	return &MapStorage{m: make(map[string]UserURL)}, nil
}

// Save saves the short URL and long URL for a given user ID in the MapStorage.
//
// Parameters:
// - userID uint64: the user ID
// - shortURL string: the short URL to be saved
// - longURL string: the long URL to be saved
// Return type: error
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

// GetRealURL retrieves the long URL associated with the given short URL.
//
// Parameters:
// - shortURL: The short URL for which the long URL needs to be retrieved.
//
// Returns:
// - string: The long URL corresponding to the short URL.
// - error: An error if the short URL is longer than ShortURLLength or if the short URL does not exist.
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

// GetShortURLBatch retrieves short URLs for a batch of long URLs.
//
// userID: The ID of the user.
// bAddr: The base address for the short URLs.
// longURLs: A slice of ReqJSONBatch containing the long URLs.
// []ResJSONBatch: A slice of ResJSONBatch containing the short URLs and any error messages.
// error: An error if the retrieval fails.
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

// GetShortURL retrieves the short URL for a given user and long URL.
//
// Parameters:
// - userID: The ID of the user.
// - longURL: The long URL for which to retrieve the short URL.
//
// Returns:
// - string: The short URL corresponding to the long URL.
// - error: An error if the short URL cannot be retrieved.
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

// IsShortURLExist checks if a short URL already exists in the MapStorage.
//
// Parameters:
//
//	shortURL - the short URL to check existence for.
//
// Returns:
//
//	bool - indicating if the short URL exists.
func (s *MapStorage) IsShortURLExist(shortURL string) bool {
	_, exist := s.m[shortURL]
	return exist
}

// IsRealURLExist checks if the given long URL exists in the MapStorage.
//
// It takes a longURL string as a parameter and returns a boolean.
func (s *MapStorage) IsRealURLExist(longURL string) bool {
	for _, lURL := range s.m {
		if lURL.LongURL == longURL {
			return true
		}
	}
	return false
}

// IsReady checks if the MapStorage is ready.
//
// Returns a boolean value indicating if the MapStorage is ready.
func (s *MapStorage) IsReady() bool {
	return true
}

// GetAllURLs retrieves all URLs for a given user and base address.
//
// Parameters:
// - userID: The ID of the user.
// - bAddr: The base address.
//
// Returns:
// - []ResJSONURL: A slice of ResJSONURL structs containing the retrieved URLs.
// - error: An error if any occurred during the retrieval process.
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

// GetLastID returns the last ID from the MapStorage.
//
// It does not take any parameters.
// It returns an integer and an error.
func (s *MapStorage) GetLastID() (int, error) {
	return len(s.m), nil
}

// DeleteURLs deletes multiple URLs for a given user.
//
// userID: The ID of the user.
// delURLs: A slice of strings containing the URLs to be deleted.
// error: An error if any occurred during the deletion process.
func (s *MapStorage) DeleteURLs(userID uint64, delURLs []string) error {
	for _, url := range delURLs {
		err := s.DeleteURL(map[string]uint64{url: userID})
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteURL deletes the specified URLs from the MapStorage.
//
// delURL: a map containing the URLs to be deleted along with their corresponding values.
// error: an error, if any.
func (s *MapStorage) DeleteURL(delURL map[string]uint64) error {
	for key, value := range delURL {
		if value != 0 {
			delete(s.m, key)
		}
	}

	return nil
}

func (s *MapStorage) GetStats() (ResJSONStats, error) {
	urls := len(s.m)
	users := make(map[uint64]uint64)
	for _, lURL := range s.m {
		users[lURL.UserID]++
	}
	return ResJSONStats{
		URLCount:  urls,
		UserCount: len(users),
	}, nil
}
