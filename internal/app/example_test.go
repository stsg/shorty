package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/storage"
)

// ExampleHandlePing is a function that shows HandlePing usage.
func (app *App) ExampleHandlePing() {
	conf := config.NewConfig()
	fmt.Println("storage type:", conf.GetStorageType())
	pStorage, err := storage.New(conf)
	if err != nil {
		panic(err)
	}
	pHandle := NewApp(conf, pStorage)

	router := chi.NewRouter()

	router.Get("/ping", pHandle.HandlePing)

	err = http.ListenAndServe(conf.GetRunAddr(), router)
	if err != nil {
		panic(errors.New("cannot run server"))
	}
}
