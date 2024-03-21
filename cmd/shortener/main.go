package main

import (
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go.uber.org/zap"

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/handle"
	mylogger "github.com/stsg/shorty/internal/logger"
	"github.com/stsg/shorty/internal/storage"
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

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mylogger.ZapLogger(logger))
	router.Use(pHandle.Decompress())
	router.Use(middleware.Compress(5, "application/json", "text/html"))

	router.Mount("/debug", middleware.Profiler())

	router.Post("/", pHandle.HandleShortRequest)
	router.Get("/ping", pHandle.HandlePing)
	router.Get("/{id}", pHandle.HandleShortID)
	router.Route("/api", func(childRouter chi.Router) {
		childRouter.Post("/shorten", pHandle.HandleShortRequestJSON)
		childRouter.Post("/shorten/batch", pHandle.HandleShortRequestJSONBatch)
		childRouter.Get("/user/urls", pHandle.HandleGetAllURLs)
		childRouter.Delete("/user/urls", pHandle.HandleDeleteURLs)
	})

	err = http.ListenAndServe(conf.GetRunAddr(), router)
	if err != nil {
		panic(errors.New("cannot run server"))
	}
}
