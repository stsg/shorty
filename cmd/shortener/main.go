package main

import (
	// "fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	// "net/url"
)

// type URLshortener struct {
// 	url map[string]string
// }

//type Shorty map[string]string

// var shorty Shorty
var shorty = make(map[string]string)

//var shorty = make(map[string]string)

func genShortURL() string {
	charset := "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	length := 6

	shortURL := make([]byte, length)
	for i := range shortURL {
		shortURL[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortURL)
}

func idHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			surl := strings.TrimPrefix(req.URL.Path, "/")
			lurl, exist := shorty[surl]
			if !exist {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.Header().Set("Location", lurl)
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(rw, req)
	})
}

func mainHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	url, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	lurl := string(url)
	for _, url := range shorty {
		if url == lurl {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	surl := genShortURL()
	shorty[surl] = lurl
	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte("http://localhost:8080/" + surl))
}

func main() {
	// shorty["surl"] = "lurl"
	// fmt.Println(shorty["surl"])
	mux := http.NewServeMux()
	mux.Handle(`/`, idHandler(http.HandlerFunc(mainHandler)))
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
