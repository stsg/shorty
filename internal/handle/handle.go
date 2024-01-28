package handle

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/storage"
)

type Handle struct {
	Config  config.Config
	storage storage.Storage
	Session *Session
}

type Session struct {
	userSession map[string]uint64
	count       uint64
}

func NewSession() *Session {
	return &Session{
		userSession: make(map[string]uint64),
		count:       0,
	}
}

func (s *Session) GetCount() uint64 {
	return s.count
}

func (s *Session) AddCount() {
	s.count++
}

func (s *Session) GetUserSession(userID string) uint64 {
	return s.userSession[userID]
}

func (s *Session) AddUserSession() (session string, count uint64) {
	s.AddCount()
	session = uuid.New().String()
	s.userSession[session] = s.count
	return session, s.count
}

func (h *Handle) SetSession(rw http.ResponseWriter, session string) {
	http.SetCookie(rw, &http.Cookie{
		Name:    "token",
		Value:   session,
		Expires: time.Now().Add(24 * time.Hour),
	})
}

func NewHandle(config config.Config, storage storage.Storage) Handle {
	handle := Handle{}
	handle.Config = config
	handle.storage = storage
	handle.Session = NewSession()

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

	userID := uint64(0)
	session := ""
	userIDToken, err := req.Cookie("token")
	if err != nil {
		userID = h.Session.GetUserSession(userIDToken.Value)
	} else {
		session, userID = h.Session.AddUserSession()
		h.SetSession(rw, session)
	}

	shortURL, err := h.storage.GetShortURL(userID, longURL)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		if errors.Is(err, storage.ErrUniqueViolation) {
			rw.WriteHeader(http.StatusConflict)
			rw.Write([]byte(h.Config.GetBaseAddr() + "/" + shortURL))
			return
		}
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(h.Config.GetBaseAddr() + "/" + shortURL))
}

func (h *Handle) HandleShortRequestJSON(rw http.ResponseWriter, req *http.Request) {
	var rqJSON storage.ReqJSON
	var rwJSON storage.ResJSON

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

	userID := uint64(0)
	session := ""
	userIDToken, err := req.Cookie("token")
	if err != nil {
		userID = h.Session.GetUserSession(userIDToken.Value)
	} else {
		session, userID = h.Session.AddUserSession()
		h.SetSession(rw, session)
	}

	rwJSON.Result, err = h.storage.GetShortURL(userID, rqJSON.URL)
	rwJSON.Result = h.Config.GetBaseAddr() + "/" + rwJSON.Result
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		if errors.Is(err, storage.ErrUniqueViolation) {
			rw.WriteHeader(http.StatusConflict)
			body, _ := json.Marshal(rwJSON)
			rw.Write([]byte(body))
			return
		}
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
	var rqJSON []storage.ReqJSONBatch

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

	userID := uint64(0)
	session := ""
	userIDToken, err := req.Cookie("token")
	if err != nil {
		userID = h.Session.GetUserSession(userIDToken.Value)
	} else {
		session, userID = h.Session.AddUserSession()
		h.SetSession(rw, session)
	}

	rwJSON, err := h.storage.GetShortURLBatch(userID, h.Config.GetBaseAddr(), rqJSON)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		rw.Write([]byte(body))
		return
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

func (h *Handle) HandleGetAllURLs(rw http.ResponseWriter, req *http.Request) {
	var resJSON []storage.ResJSONURL

	userIDToken, err := req.Cookie("token")
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusNoContent)
		rw.Write([]byte(err.Error()))
		return
	}
	userID := h.Session.GetUserSession(userIDToken.Value)
	resJSON, err = h.storage.GetAllURLs(userID)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	body, _ := json.Marshal(resJSON)
	rw.Write([]byte(body))
}
