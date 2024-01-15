package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mylogger "github.com/stsg/shorty/internal/logger"

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/handle"
	"github.com/stsg/shorty/internal/storage"

	"go.uber.org/zap"
)

func main() {
	conf := config.NewConfig()
	fmt.Println("storage type:", conf.GetStorageType())
	pStorage, err := storage.New(conf)
	if err != nil {
		panic(err)
	}
	pHandle := handle.NewHandle(conf, pStorage)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(errors.New("cannot create logger"))
	}
	defer logger.Sync()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mylogger.ZapLogger(logger))
	r.Use(pHandle.Decompress())
	r.Use(middleware.Compress(5, "application/json", "text/html"))

	r.Post("/", pHandle.HandleShortRequest)
	r.Get("/ping", pHandle.HandlePing)
	r.Get("/{id}", pHandle.HandleShortID)
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", pHandle.HandleShortRequestJSON)
	})

	err = http.ListenAndServe(conf.GetRunAddr(), r)
	if err != nil {
		panic(errors.New("cannot run server"))
	}
}
