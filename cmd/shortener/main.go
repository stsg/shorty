package main

// TODO

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/handle"
	mylogger "github.com/stsg/shorty/internal/logger"
	"github.com/stsg/shorty/internal/storage"

	"go.uber.org/zap"
)

func main() {
	conf := config.NewConfig()
	strg, err := storage.New(conf)
	if err != nil {
		panic(err)
	}
	hndl := handle.NewHandle(conf, strg)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(errors.New("cannot create logger"))
	}
	defer logger.Sync()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mylogger.ZapLogger(logger))
	r.Use(hndl.Decompress())
	r.Use(middleware.Compress(5, "application/json", "text/html"))

	r.Post("/", hndl.HandleShortRequest)
	r.Get("/{id}", hndl.HandleShortID)
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", hndl.HandleShortRequestJSON)
	})

	err = http.ListenAndServe(conf.GetRunAddr(), r)
	if err != nil {
		panic(errors.New("cannot run server"))
	}
}
