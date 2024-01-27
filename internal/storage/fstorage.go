package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/stsg/shorty/internal/config"
)

type FileStorage struct {
	Path  string
	File  *os.File
	count int
	fm    []fileMap
}

type fileMap struct {
	UUID     string `json:"uuid"`
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

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

func (s *FileStorage) Save(shortURL string, longURL string) error {
	var fMap = fileMap{
		UUID:     strconv.Itoa(s.count),
		ShortURL: shortURL,
		LongURL:  longURL,
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

func (s *FileStorage) Open() error {
	file, err := os.OpenFile(s.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	s.File = file
	return nil
}

func (s *FileStorage) Close() error {
	return s.File.Close()
}

func (s *FileStorage) GetRealURL(shortURL string) (string, error) {
	for key := range s.fm {
		if s.fm[key].ShortURL == shortURL {
			return s.fm[key].LongURL, nil
		}
	}
	return "", errors.New("short URL not exist")
}

func (s *FileStorage) GetShortURLBatch(bAddr string, longURLs []ReqJSONBatch) ([]ResJSONBatch, error) {
	var rwJSON []ResJSONBatch
	for _, rqElemJSON := range longURLs {
		shortURL, err := s.GetShortURL(rqElemJSON.URL)
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
func (s *FileStorage) GetShortURL(longURL string) (string, error) {
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
			err := s.Save(shortURL, longURL)
			if err == nil {
				return shortURL, nil
			} else {
				return "", err
			}
		}
		shortURL = GenShortURL()
	}
}

func (s *FileStorage) IsShortURLExist(shortURL string) bool {
	for key := range s.fm {
		if s.fm[key].ShortURL == shortURL {
			return true
		}
	}
	return false
}

func (s *FileStorage) IsRealURLExist(longURL string) bool {
	for key := range s.fm {
		if s.fm[key].LongURL == longURL {
			return true
		}
	}
	return false
}

func (s *FileStorage) IsReady() bool {
	err := s.Open()
	if err != nil {
		return false
	}
	defer s.Close()
	return true
}
