// Package app provides main application logic.
package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	mylogger "github.com/stsg/shorty/internal/logger"
)

const certSerialMaxInt = 1024

var protectedURLs = []string{
	"/api/internal/stats",
}

// App class definition defines a struct named App with the following fields:
//
// Config of type config.Config
// storage of type storage.Storage
// Session of type *Session (pointer to Session)
// delChan of type chan map[string]uint64 (channel of map[string]uint64)
//
// App holds main application
type App struct {
	storage storage.Storage
	Session *Session
	delChan chan map[string]uint64
	Config  config.Config
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
func (app *App) Run(ctx context.Context) error {
	logger := mylogger.Get()
	defer func() {
		logger.Sync()
		if x := recover(); x != nil {
			logger.Sugar().Error(x)
			panic(x)
		}
	}()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mylogger.ZapLogger())
	router.Use(app.Decompress())
	router.Use(middleware.Compress(5, "application/json", "text/html"))
	router.Use(app.TrustedSubnets())

	router.Mount("/debug", middleware.Profiler())

	router.Post("/", app.HandleShortRequest)
	router.Get("/ping", app.HandlePing)
	router.Get("/{id}", app.HandleShortID)
	router.Route("/api", func(childRouter chi.Router) {
		childRouter.Post("/shorten", app.HandleShortRequestJSON)
		childRouter.Post("/shorten/batch", app.HandleShortRequestJSONBatch)
		childRouter.Get("/user/urls", app.HandleGetAllURLs)
		childRouter.Delete("/user/urls", app.HandleDeleteURLs)
		childRouter.Get("/internal/stats", app.HandleInternalStats)
	})

	srv := &http.Server{
		Addr:              app.Config.GetRunAddr(),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       time.Second,
	}

	go func() {
		<-ctx.Done()
		if srv != nil {
			if err := srv.Close(); err != nil {
				logger.Error("shutting down by signal")
			}
		}
	}()

	if app.Config.GetEnableHTTPS() {
		logger.Info("Creating certificate...")
		err := app.createCertificate()
		if err != nil {
			panic(fmt.Sprintf("cannot create certificate: %v", err))
		}
		logger.Info("Certificate created.")
		err = srv.ListenAndServeTLS("./data/cert/cert.pem", "./data/cert/key.pem")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("cannot run https server: %v", err))
		}
	} else {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("cannot run http server: %v", err))
		}
	}

	return nil
}

// createCertificate generates a certificate and private key, saving them to disk.
//
// No parameters.
// Returns an error.
func (app *App) createCertificate() error {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(int64(certSerialMaxInt)),
		Subject: pkix.Name{
			Organization:       []string{"Localhost Ent."},
			OrganizationalUnit: []string{"Shorty Server"},
			CommonName:         "localhost",
			Country:            []string{"RU"},
		},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("error generate key %w", err)
	}

	certData, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("error create certificate %w", err)
	}

	var certDataBytes bytes.Buffer
	err = pem.Encode(&certDataBytes, &pem.Block{Type: "CERTIFICATE", Bytes: certData})
	if err != nil {
		return fmt.Errorf("error encode certificate %w", err)
	}
	err = os.WriteFile("./data/cert/cert.pem", certDataBytes.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error write certificate %w", err)
	}

	var privateKeyBytes bytes.Buffer
	err = pem.Encode(&privateKeyBytes, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err != nil {
		return fmt.Errorf("error encode private key %w", err)
	}
	err = os.WriteFile("./data/cert/key.pem", privateKeyBytes.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error write private key %w", err)
	}

	return nil
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
