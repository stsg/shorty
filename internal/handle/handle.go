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
	config  config.Config
	storage storage.Storage
}

type reqJSON struct {
	URL string `json:"url,omitempty"`
}

type resJSON struct {
	Result string `json:"result"`
}

type reqJSONBatch struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url,omitempty"`
}

type resJSONBatch struct {
	ID     string `json:"correlation_id"`
	Result string `json:"short_url,omitempty"`
}

func NewHandle(config config.Config, storage storage.Storage) Handle {
	handle := Handle{}
	handle.config = config
	handle.storage = storage

	return handle
}

func (h *Handle) HandlePing(rw http.ResponseWriter, req *http.Request) {
	ping := strings.TrimPrefix(req.URL.Path, "/")
	ping = strings.TrimSuffix(ping, "/")
	if !h.storage.IsReady() {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("storage not ready"))
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(ping + " - pong"))
}

func (h *Handle) HandleShortID(rw http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	id = strings.TrimSuffix(id, "/")
	longURL, err := h.storage.GetRealURL(id)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(err.Error()))
		return

	}
	rw.Header().Set("Location", longURL)
	rw.WriteHeader(http.StatusTemporaryRedirect)
	rw.Write([]byte(longURL))
}

func (h *Handle) HandleShortRequest(rw http.ResponseWriter, req *http.Request) {
	url, err := io.ReadAll(req.Body)
	if err != nil {
		panic(errors.New("cannot read request body"))
	}
	longURL := string(url)
	shortURL, err := h.storage.GetShortURL(longURL)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(h.config.GetBaseAddr() + "/" + shortURL))
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
	rwJSON.Result, err = h.storage.GetShortURL(rqJSON.URL)
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

func (h *Handle) HandleShortRequestJSONBatch(rw http.ResponseWriter, req *http.Request) {
	var rqJSON []reqJSONBatch
	var rwJSON []resJSONBatch

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
	for _, rqElemJSON := range rqJSON {
		shortURL, err := h.storage.GetShortURL(rqElemJSON.URL)
		shortURL = h.config.GetBaseAddr() + "/" + shortURL
		rwElemJSON := resJSONBatch{
			ID:     rqElemJSON.ID,
			Result: shortURL,
		}
		if err != nil {
			rwElemJSON.Result = err.Error()
		}
		rwJSON = append(rwJSON, rwElemJSON)
	}
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
