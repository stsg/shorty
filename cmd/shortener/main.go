package main

// TODO
// @rktkov Основное замечание ментора:
// Сейчас принимаю плоскую структуру,
// но в следующей раз без разбиения
// на логические уровни их связывания
// через интерфейсы код не приму.

import (
	"errors"
	"net/http"

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
	strg := storage.NewMapStorage()
	hndl := handle.NewHandle(conf, *strg)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(errors.New("cannot create logger"))
	}
	defer logger.Sync()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mylogger.ZapLogger(logger))
	r.Use(middleware.Recoverer)

	r.Post("/", hndl.HandleShortRequest)
	r.Get("/{id}", hndl.HandleShortID)

	err = http.ListenAndServe(conf.GetRunAddr(), r)
	if err != nil {
		panic(errors.New("cannot run server"))
	}
}
