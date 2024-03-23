package app

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go.uber.org/zap"

	mylogger "github.com/stsg/shorty/internal/logger"
)

// This class definition defines a struct named App with the following fields:
//
// Config of type config.Config
// storage of type storage.Storage
// Session of type *Session (pointer to Session)
// delChan of type chan map[string]uint64 (channel of map[string]uint64)
//
// App holds main application
type App struct {
	Config  config.Config
	storage storage.Storage
	Session *Session
	delChan chan map[string]uint64
}

// Session is a struct that holds user session data.
type Session struct {
	storage     storage.Storage
	userSession map[string]uint64
	count       *atomic.Uint64
}

// Run runs the App.
//
// It initializes the logger and router, sets up middleware, mounts routes, and starts the server.
// It returns an error if there is any issue running the server.
func (app *App) Run() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(errors.New("cannot create logger"))
	}
	defer logger.Sync()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mylogger.ZapLogger(logger))
	router.Use(app.Decompress())
	router.Use(middleware.Compress(5, "application/json", "text/html"))

	router.Mount("/debug", middleware.Profiler())

	router.Post("/", app.HandleShortRequest)
	router.Get("/ping", app.HandlePing)
	router.Get("/{id}", app.HandleShortID)
	router.Route("/api", func(childRouter chi.Router) {
		childRouter.Post("/shorten", app.HandleShortRequestJSON)
		childRouter.Post("/shorten/batch", app.HandleShortRequestJSONBatch)
		childRouter.Get("/user/urls", app.HandleGetAllURLs)
		childRouter.Delete("/user/urls", app.HandleDeleteURLs)
	})

	err = http.ListenAndServe(app.Config.GetRunAddr(), router)
	if err != nil {
		panic(fmt.Errorf("cannot run server: %v", err))
	}

	return err
}

// NewSession creates a new session with the given storage.
//
// It retrieves the last ID from the storage and initializes a new atomic ID
// with that value. It then creates and returns a new Session object with the
// provided storage, an empty user session map, and the initialized atomic ID.
//
// Parameters:
// - storage: The storage to be used for the session.
//
// Returns:
// - *Session: A pointer to the newly created session.
func NewSession(storage storage.Storage) *Session {
	lastID, err := storage.GetLastID()
	if err != nil {
		return nil
	}

	newID := atomic.Uint64{}
	newID.Store(uint64(lastID))
	return &Session{
		storage:     storage,
		userSession: make(map[string]uint64),
		count:       &newID,
	}
}

// GetUserSessionID returns the user session ID associated with the given session ID.
//
// Parameters:
// - sessionID: The ID of the session.
//
// Returns:
// - uint64: The user session ID.
func (s *Session) GetUserSessionID(sessionID string) uint64 {
	return s.userSession[sessionID]
}

// AddUserSession adds a new user session to the Session struct.
//
// No parameters.
// Returns a string representing the session and a uint64 representing the count.
func (s *Session) AddUserSession() (session string, count uint64) {
	s.count.Add(1)
	session = uuid.New().String()
	s.userSession[session] = s.count.Load()
	return session, s.count.Load()
}

// SetSession sets a session for the Handle.
//
// It takes the http.ResponseWriter and session string as parameters and does not return anything.
func (app *App) SetSession(rw http.ResponseWriter, session string) {
	http.SetCookie(rw, &http.Cookie{
		Name:    "token",
		Value:   session,
		Expires: time.Now().Add(24 * time.Hour),
	})
}

// NewApp creates a new handle object with the provided configuration and storage.
// It returns the handle object along with a new session object.
func NewApp(config config.Config, storage storage.Storage) App {
	app := App{
		Config:  config,
		storage: storage,
		Session: NewSession(storage),
		delChan: make(chan map[string]uint64, 500),
	}

	go func() {
		for delURL := range app.delChan {
			err := storage.DeleteURL(delURL)
			if err != nil {
				continue
			}
		}
	}()

	return app
}

// HandlePing handles the ping request.
//
// It takes in the http.ResponseWriter and *http.Request as parameters.
// It does not return any values.
func (app *App) HandlePing(rw http.ResponseWriter, req *http.Request) {
	ping := strings.TrimPrefix(req.URL.Path, "/")
	ping = strings.TrimSuffix(ping, "/")
	if !app.storage.IsReady() {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("storage not ready"))
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(ping + " - pong"))
}

// HandleShortID handles the shortened URL request and redirects the client to the corresponding long URL.
//
// Parameters:
// - rw: http.ResponseWriter - the response writer used to write the response.
// - req: *http.Request - the HTTP request object containing the URL path.
//
// Returns: None.
func (app *App) HandleShortID(rw http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	id = strings.TrimSuffix(id, "/")
	longURL, err := app.storage.GetRealURL(id)
	if errors.Is(err, storage.ErrURLDeleted) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusGone)
		rw.Write([]byte(err.Error()))
		return
	}
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

func (app *App) HandleShortRequest(rw http.ResponseWriter, req *http.Request) {
	var userID uint64
	var session string

	url, err := io.ReadAll(req.Body)
	if err != nil {
		panic(errors.New("cannot read request body"))
	}

	longURL := string(url)
	if longURL == "" {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("url is empty"))
		return
	}

	userIDToken, err := req.Cookie("token")
	if err == nil {
		userID = app.Session.GetUserSessionID(userIDToken.Value)
	} else {
		session, userID = app.Session.AddUserSession()
		app.SetSession(rw, session)
	}

	shortURL, err := app.storage.GetShortURL(userID, longURL)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		if errors.Is(err, storage.ErrUniqueViolation) {
			rw.WriteHeader(http.StatusConflict)
			rw.Write([]byte(app.Config.GetBaseAddr() + "/" + shortURL))
			return
		}
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(app.Config.GetBaseAddr() + "/" + shortURL))
}

func (app *App) HandleShortRequestJSON(rw http.ResponseWriter, req *http.Request) {
	var rqJSON storage.ReqJSON
	var rwJSON storage.ResJSON
	var userID uint64
	var session string

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

	userIDToken, err := req.Cookie("token")
	if err == nil {
		userID = app.Session.GetUserSessionID(userIDToken.Value)
	} else {
		session, userID = app.Session.AddUserSession()
		app.SetSession(rw, session)
	}

	rwJSON.Result, err = app.storage.GetShortURL(userID, rqJSON.URL)
	rwJSON.Result = app.Config.GetBaseAddr() + "/" + rwJSON.Result
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

func (app *App) HandleShortRequestJSONBatch(rw http.ResponseWriter, req *http.Request) {
	var rqJSON []storage.ReqJSONBatch
	var userID uint64
	var session string

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

	userIDToken, err := req.Cookie("token")
	if err == nil {
		userID = app.Session.GetUserSessionID(userIDToken.Value)
	} else {
		session, userID = app.Session.AddUserSession()
		app.SetSession(rw, session)
	}

	rwJSON, err := app.storage.GetShortURLBatch(userID, app.Config.GetBaseAddr(), rqJSON)
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

func (app *App) Decompress() func(http.Handler) http.Handler {
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

func (app *App) HandleGetAllURLs(rw http.ResponseWriter, req *http.Request) {
	var resJSON []storage.ResJSONURL
	var userID uint64

	userIDToken, err := req.Cookie("token")
	if err == nil {
		userID = app.Session.GetUserSessionID(userIDToken.Value)
	} else {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte(err.Error()))
		return
	}

	resJSON, err = app.storage.GetAllURLs(userID, app.Config.GetBaseAddr())
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	if len(resJSON) == 0 {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusNoContent)
		rw.Write([]byte("no content for this user"))
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	// body, _ := json.Marshal(resJSON)
	body, _ := json.MarshalIndent(resJSON, "", "    ")
	rw.Write([]byte(body))
}

func (app *App) HandleDeleteURLs(rw http.ResponseWriter, req *http.Request) {
	var delURLs []string
	var userID uint64

	urls, err := io.ReadAll(req.Body)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(urls, &delURLs)
	if err != nil {
		rw.Header().Set("Content-Type", "application/text")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	userIDToken, err := req.Cookie("token")
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte(err.Error()))
		return

	}
	userID = app.Session.GetUserSessionID(userIDToken.Value)

	for _, url := range delURLs {
		go func(url string, userID uint64) {
			app.delChan <- map[string]uint64{
				url: userID,
			}
		}(url, userID)

	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusAccepted)
	rw.Write([]byte("Accepted"))
}
