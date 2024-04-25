// Package main - main package for shorty service, URL shortener application
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/zap"

	"github.com/stsg/shorty/internal/app"
	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/logger"
	"github.com/stsg/shorty/internal/storage"
)

const maxDupmSize = 5 * 1024 * 1024

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
	// logger, _ := zap.NewDevelopment()
	logger := logger.Get()
	// zFields := []zap.Field{
	// 	zap.String("version", buildVersion),
	// 	zap.String("commit", buildCommit),
	// 	zap.String("date", buildDate),
	// }
	// observedZapCore, observerLogs := observer.New(zap.InfoLevel)
	logger.Info("starting shorty", zap.String("version", buildVersion), zap.String("date", buildDate), zap.String("commit", buildCommit))
	// logger.Info("starting shorty", zFields...)

	// , zap.String("date", buildDate)
	// fmt.Printf("shorty version: %s, build date: %s, build commit: %s\n", buildVersion, buildDate, buildCommit)
	conf := config.NewConfig()
	// fmt.Println("storage type:", conf.GetStorageType())
	logger.Info("config:", zap.String("storage type", conf.GetStorageType()))
	// fmt.Println("https:", conf.GetEnableHTTPS())
	logger.Info("config:", zap.Bool("https", conf.GetEnableHTTPS()))
	pStorage, err := storage.New(conf)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// catch signal for graceful shutdown
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-stop
		// fmt.Println("shutting down by signal")
		logger.Info("shutting down by signal")
		// fmt.Printf("stacktrace:\n%s\n", getDump())
		logger.Info("stacktrace:", zap.String("stacktrace", getDump()))
		cancel()
	}()

	shortyApp := app.NewApp(conf, pStorage)
	err = shortyApp.Run(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("application running error: %v", err))
	}
}

func getDump() string {
	stacktrace := make([]byte, maxDupmSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxDupmSize {
		length = maxDupmSize
	}
	return string(stacktrace[:length])
}
