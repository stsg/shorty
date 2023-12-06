package main

import (
	// "fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	// "net/url"
)

var Shorty = make(map[string]string)

const ShortURLLength = 6

func genShortURL() string {
	charset := "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	shortURL := make([]byte, ShortURLLength)
	for i := range shortURL {
		shortURL[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortURL)
}

func getRealURL(rw http.ResponseWriter, req *http.Request) {
	surl := strings.TrimPrefix(req.URL.Path, "/")
	surl = strings.TrimSuffix(surl, "/")
	if len(surl) > ShortURLLength {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	lurl, exist := Shorty[surl]
	rw.Header().Set("Location", lurl)
	rw.Header().Set("Content-Type", "text/plain")
	if !exist {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusTemporaryRedirect)
	rw.Write([]byte(lurl))
}

func getShortURL(rw http.ResponseWriter, req *http.Request) {
	url, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	lurl := string(url)
	for _, url := range Shorty {
		if url == lurl {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	surl := genShortURL()
	Shorty[surl] = lurl
	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte("http://localhost:8080/" + surl))
}

func mainHandler(rw http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		getShortURL(rw, req)
	case http.MethodGet:
		getRealURL(rw, req)
	default:
		rw.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	Shorty["123456"] = "https://www.google.com"

	mux := http.NewServeMux()
	mux.Handle(`/`, http.HandlerFunc(mainHandler))
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
