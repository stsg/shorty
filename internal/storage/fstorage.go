package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
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
	UUID      string `json:"uuid"`
	Short_URL string `json:"short_url"`
	Long_URL  string `json:"original_url"`
}

func NewFileStorage(config config.Config) (*FileStorage, error) {
	var fMap fileMap

	fs := &FileStorage{
		Path:  config.GetFileStor(),
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
		UUID:      strconv.Itoa(s.count),
		Short_URL: shortURL,
		Long_URL:  longURL,
	}

	s.fm = append(s.fm, fMap)
	jsonData, err := json.Marshal(fMap)
	if err != nil {
		return err
	}
	_, err = s.File.Write(append(jsonData, byte('\n')))
	if err != nil {
		return err
	}
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
	var fMap fileMap

	err := s.Open()
	if err != nil {
		fmt.Println("cannot open file")
		return "", err
	}
	defer s.File.Close()

	scanner := bufio.NewScanner(s.File)
	for scanner.Scan() {
		line := scanner.Bytes()
		err := json.Unmarshal(line, &fMap)
		if err != nil {
			fmt.Println("cannot read JSON from file", err)
			continue
		}
		if fMap.Short_URL == shortURL {
			return fMap.Long_URL, nil
		}
	}
	err = scanner.Err()
	if err != nil {
		fmt.Println("cannot scan JSON from file")
		return "", err
	}
	return "", nil
}

func (s *FileStorage) GetShortURL(longURL string) (string, error) {
	var fMap fileMap

	err := s.Open()
	if err != nil {
		fmt.Println("cannot open file")
		return "", err
	}
	defer s.File.Close()

	shortURL := GenShortURL()

	scanner := bufio.NewScanner(s.File)
	for scanner.Scan() {
		line := scanner.Bytes()
		err := json.Unmarshal(line, &fMap)
		if err != nil {
			fmt.Println("cannot read JSON from file", err)
			continue
		}
		if fMap.Long_URL == longURL {
			return fMap.Short_URL, errors.New("short URL already exist")
		}
		if fMap.Short_URL == shortURL {
			shortURL = GenShortURL()
		}
	}
	s.Save(shortURL, longURL)
	return shortURL, nil
}

func (s *FileStorage) IsShortURLExist(shortURL string) bool {
	var fMap fileMap

	err := s.Open()
	if err != nil {
		return false
	}
	defer s.File.Close()

	scanner := bufio.NewScanner(s.File)
	for scanner.Scan() {
		line := scanner.Bytes()
		err := json.Unmarshal(line, &fMap)
		if err != nil {
			continue
		}
		if fMap.Short_URL == shortURL {
			return true
		}
	}
	return false
}

func (s *FileStorage) IsRealURLExist(longURL string) bool {
	var fMap fileMap

	err := s.Open()
	if err != nil {
		return false
	}
	defer s.File.Close()

	scanner := bufio.NewScanner(s.File)
	for scanner.Scan() {
		line := scanner.Bytes()
		err := json.Unmarshal(line, &fMap)
		if err != nil {
			continue
		}
		if fMap.Long_URL == longURL {
			return true
		}
	}
	return false
}
