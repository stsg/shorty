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

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/handle"
	"github.com/stsg/shorty/internal/storage"
)

func main() {

	conf := config.NewConfig()
	strg := storage.NewMapStorage()
	hndl := handle.NewHandle(conf, *strg)

	r := chi.NewRouter()

	r.Post("/", hndl.HandleShortRequest)
	r.Get("/{id}", hndl.HandleShortId)

	err := http.ListenAndServe(conf.GetRunAddr(), r)
	if err != nil {
		panic(errors.New("cannot run server"))
	}
}
