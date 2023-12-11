package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/stsg/shorty/cmd/shortener/config"
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
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	lurl, exist := Shorty[surl]
	rw.Header().Set("Location", lurl)
	rw.Header().Set("Content-Type", "text/plain")
	if !exist {
		rw.Header().Set("Content-Type", "text/plain")
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
	// rw.Write([]byte("http://localhost:8080/" + surl))
	rw.Write([]byte(config.ShortyCnf.Host + surl))
}

func main() {
	Shorty["123456"] = "https://www.google.com"

	config.ParseFlags()
	config.ParseEnv()

	r := chi.NewRouter()

	r.Post("/", getShortURL)
	r.Get("/{id}", getRealURL)

	fmt.Println("Short URLs server address", config.ShortyCnf.Host)
	fmt.Println("Running server on", config.ShortyCnf.RunAddr)
	err := http.ListenAndServe(config.ShortyCnf.RunAddr, r)
	if err != nil {
		panic(err)
	}
}
