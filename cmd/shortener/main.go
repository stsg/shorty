// Package main - main package for shorty service, URL shortener application
package main

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/stsg/shorty/internal/app"
	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/storage"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// main is the entry point of the program.
//
// It initializes the configuration, creates a new storage instance, and sets up the logger.
// Then it creates a new router and sets up middleware for request handling.
// After that, it mounts the debug routes and sets up the routes for handling different requests.
// Finally, it starts the HTTP server and listens for incoming requests.
func main() {
	fmt.Printf("shorty version: %s, build date: %s, build commit: %s\n", buildVersion, buildDate, buildCommit)
	conf := config.NewConfig()
	fmt.Println("storage type:", conf.GetStorageType())
	pStorage, err := storage.New(conf)
	if err != nil {
		panic(err)
	}
	shortyApp := app.NewApp(conf, pStorage)
	err = shortyApp.Run()
	if err != nil {
		panic(fmt.Errorf("application running error: %v", err))
	}

	// os.Exit(256)
}
