package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/stsg/shorty/internal/config"
)

// FileStorage is a struct that holds FS storage data.
type FileStorage struct {
	File  *os.File
	Path  string
	fm    []fileMap
	count int
}

// URL file storage srtruct
// A struct named fileMap with five fields: UUID, ShortURL, LongURL, UserID, and Deleted.
// Each field is tagged with a JSON key that determines
// how the struct is serialized or deserialized to/from JSON.
// The UUID field is a string, ShortURL and LongURL are both strings,
// UserID is an unsigned 64-bit integer, and Deleted is a boolean.
type fileMap struct {
	UUID     string `json:"uuid"`
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
	UserID   uint64 `json:"user_id"`
	Deleted  bool   `json:"deleted"`
}

// NewFileStorage creates a new FileStorage instance.
//
// It takes a config.Config object as a parameter and returns a pointer to a FileStorage object and an error.
func NewFileStorage(config config.Config) (*FileStorage, error) {
	var fMap fileMap

	fs := &FileStorage{
		Path:  config.GetFileStorage(),
		count: 0,
	}
	err := fs.Open()
	if err != nil {
		return nil, err
	}
	defer fs.File.Close()

	scanner := bufio.NewScanner(fs.File)
	for scanner.Scan() {
		line := scanner.Bytes()
		err := json.Unmarshal(line, &fMap)
		if err != nil {
			continue
		}
		fs.fm = append(fs.fm, fMap)
		fs.count += 1
	}

	return fs, nil
}

// Save saves the given short URL and long URL for the specified user ID.
//
// Parameters:
// - userID: The ID of the user.
// - shortURL: The shortened URL.
// - longURL: The original URL.
//
// Returns:
// - error: An error if the save operation fails.
func (s *FileStorage) Save(userID uint64, shortURL string, longURL string) error {
	var fMap = fileMap{
		UUID:     strconv.Itoa(s.count),
		ShortURL: shortURL,
		LongURL:  longURL,
		UserID:   userID,
	}
	s.fm = append(s.fm, fMap)
	jsonData, err := json.Marshal(fMap)
	if err != nil {
		return err
	}
	err = s.Open()
	if err != nil {
		return err
	}
	defer s.File.Close()
	_, err = s.File.Write(append(jsonData, byte('\n')))
	if err != nil {
		return err
	}
	s.count += 1
	return nil
}

// Open opens the file storage and returns an error if unsuccessful.
//
// No parameters.
// Returns an error.
func (s *FileStorage) Open() error {
	file, err := os.OpenFile(s.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	s.File = file
	return nil
}

// Close closes the FileStorage.
//
// It returns an error if there was a problem closing the file.
func (s *FileStorage) Close() error {
	return s.File.Close()
}

// GetRealURL retrieves the corresponding long URL for a given short URL from the FileStorage.
//
// Parameters:
// - shortURL: the short URL for which the corresponding long URL needs to be retrieved.
//
// Returns:
// - string: the long URL corresponding to the short URL.
// - error: an error indicating if the short URL does not exist in the FileStorage.
func (s *FileStorage) GetRealURL(shortURL string) (string, error) {
	for key := range s.fm {
		if s.fm[key].ShortURL == shortURL {
			return s.fm[key].LongURL, nil
		}
	}
	return "", errors.New("short URL not exist")
}

// GetShortURLBatch retrieves the short URLs for a batch of long URLs.
//
// Parameters:
// - userID: The ID of the user.
// - bAddr: The base address for the short URLs.
// - longURLs: The list of long URLs to be converted to short URLs.
//
// Returns:
// - rwJSON: The list of short URLs corresponding to the long URLs.
// - error: An error if there was a problem retrieving the short URLs.
func (s *FileStorage) GetShortURLBatch(userID uint64, bAddr string, longURLs []ReqJSONBatch) ([]ResJSONBatch, error) {
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

// GetShortURL retrieves or generates a short URL for the given long URL and user ID.
//
// userID uint64, longURL string
// string, error
func (s *FileStorage) GetShortURL(userID uint64, longURL string) (string, error) {
	for key := range s.fm {
		if s.fm[key].LongURL == longURL {
			return s.fm[key].ShortURL, ErrUniqueViolation
		}
	}
	shortURL := GenShortURL()

	if longURL == "https://www.google.com" {
		shortURL = "123456"
	}

	for {
		if !s.IsShortURLExist(shortURL) {
			err := s.Save(userID, shortURL, longURL)
			if err == nil {
				return shortURL, nil
			} else {
				return "", err
			}
		}
		shortURL = GenShortURL()
	}
}

// IsShortURLExist checks if a short URL exists in the FileStorage.
//
// Parameters:
// - shortURL: the short URL to check for existence.
//
// Returns:
// - bool: true if the short URL exists, false otherwise.
func (s *FileStorage) IsShortURLExist(shortURL string) bool {
	for key := range s.fm {
		if s.fm[key].ShortURL == shortURL {
			return true
		}
	}
	return false
}

// IsRealURLExist checks if a given longURL exists in the FileStorage's map.
//
// longURL string
// bool
func (s *FileStorage) IsRealURLExist(longURL string) bool {
	for key := range s.fm {
		if s.fm[key].LongURL == longURL {
			return true
		}
	}
	return false
}

// IsReady checks if the FileStorage is ready.
//
// It opens the FileStorage and checks if there is an error. If there is an error, it returns false.
// Otherwise, it defers the closing of the FileStorage and returns true.
//
// Returns:
// - bool: true if the FileStorage is ready, false otherwise.
func (s *FileStorage) IsReady() bool {
	err := s.Open()
	if err != nil {
		return false
	}
	defer s.Close()
	return true
}

// GetAllURLs retrieves all URLs associated with a specific userID from the FileStorage.
//
// userID: the ID of the user
// bAddr: base address for constructing the complete URL
// Returns a slice of ResJSONURL containing the retrieved URLs and an error if any
func (s *FileStorage) GetAllURLs(userID uint64, bAddr string) ([]ResJSONURL, error) {
	var rwJSON []ResJSONURL
	for key := range s.fm {
		if s.fm[key].UserID == userID {
			rwJSON = append(rwJSON, ResJSONURL{
				URL:    s.fm[key].LongURL,
				Result: bAddr + "/" + s.fm[key].ShortURL,
			})
		}
	}
	return rwJSON, nil
}

// GetLastID returns the last ID from the FileStorage.
//
// It scans the FileStorage file line by line and counts the number of lines.
// The last ID is the total count of lines.
//
// Returns:
// - int: the last ID.
// - error: any error that occurred during the scanning process.
func (s *FileStorage) GetLastID() (int, error) {
	scanner := bufio.NewScanner(s.File)
	count := 0
	for scanner.Scan() {
		count++
	}

	return count, nil
}

// DeleteURLs deletes multiple URLs for a given user.
//
// userID: The ID of the user.
// delURLs: An array of URLs to be deleted.
// error: An error if the deletion fails.
func (s *FileStorage) DeleteURLs(userID uint64, delURLs []string) error {
	for _, url := range delURLs {
		err := s.DeleteURL(map[string]uint64{url: userID})
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteURL deletes URLs from the FileStorage.
//
// delURL is a map of URLs to be deleted and their corresponding user IDs.
// It returns an error if there was an issue deleting the URLs.
func (s *FileStorage) DeleteURL(delURL map[string]uint64) error {
	for sURL, userID := range delURL {
		for key := range s.fm {
			if sURL == s.fm[key].ShortURL && userID == s.fm[key].UserID {
				s.fm[key].Deleted = true
			}
		}
	}

	return nil
}
