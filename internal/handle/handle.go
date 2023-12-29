package handle

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/storage"
)

type Handle struct {
	config config.Config
	strg   storage.Storage
}

type reqJSON struct {
	URL string `json:"url,omitempty"`
}

type resJSON struct {
	Result string `json:"result"`
}

func NewHandle(config config.Config, strg storage.Storage) Handle {
	hndl := Handle{}
	hndl.config = config
	hndl.strg = strg

	return hndl
}

func (h *Handle) HandleShortID(rw http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	id = strings.TrimSuffix(id, "/")
	lurl, err := h.strg.GetRealURL(id)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(err.Error()))
		return

	}
	rw.Header().Set("Location", lurl)
	rw.WriteHeader(http.StatusTemporaryRedirect)
	rw.Write([]byte(lurl))
}

func (h *Handle) HandleShortRequest(rw http.ResponseWriter, req *http.Request) {
	url, err := io.ReadAll(req.Body)
	if err != nil {
		panic(errors.New("cannot read request body"))
	}
	lurl := string(url)
	surl, err := h.strg.GetShortURL(lurl)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(h.config.GetBaseAddr() + "/" + surl))
}

func (h *Handle) HandleShortRequestJSON(rw http.ResponseWriter, req *http.Request) {
	var rqJSON reqJSON
	var rwJSON resJSON

	url, err := io.ReadAll(req.Body)
	if err != nil {
		panic(errors.New("cannot read request body"))
	}
	err = json.Unmarshal(url, &rqJSON)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		rw.Write([]byte(body))
		return
	}
	rwJSON.Result, err = h.strg.GetShortURL(rqJSON.URL)
	rwJSON.Result = h.config.GetBaseAddr() + "/" + rwJSON.Result
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		rw.Write([]byte(body))
		return
	}
	rw.Header().Set("Location", rwJSON.Result)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	body, _ := json.Marshal(rwJSON)
	rw.Write([]byte(body))
}

func (h *Handle) Decompress() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Content-Encoding") == "gzip" {
				reader, err := gzip.NewReader(req.Body)
				if err != nil {
					rw.Header().Set("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(err.Error()))
					return
				}
				defer reader.Close()

				buf := new(strings.Builder)
				_, err = io.Copy(buf, reader)
				if err != nil {
					rw.Header().Set("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(err.Error()))
					return
				}
				req.Body = io.NopCloser(strings.NewReader(buf.String()))
				req.Header.Set("Content-Length", string(rune(len(buf.String()))))
			}
			next.ServeHTTP(rw, req)
		})
	}
}
