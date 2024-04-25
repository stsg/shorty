// Package logger provides a middleware to log HTTP requests and responses using Zap logger.
package logger

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ctxKey struct{}

var once sync.Once

var logger *zap.Logger

// ZapLogger returns a function that can be used as middleware to log HTTP requests and responses using Zap logger.
//
// It takes a *zap.Logger as input parameter and returns a function that takes http.Handler as input parameter and returns http.Handler.
func ZapLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := Get()
			var fields []zapcore.Field
			var headers []string

			for k := range r.Header {
				headers = append(headers, k)
			}

			for _, h := range headers {
				fields = append(fields, zap.String(h, r.Header.Get(h)))
			}
			logger.Info("header", fields...)

			fields = []zap.Field{
				zap.String("Method", r.Method),
				zap.String("Host", r.Host),
				zap.String("RequestURI", r.RequestURI),
				zap.String("Proto", r.Proto),
				zap.String("RemoteAddr", r.RemoteAddr),
				zap.String("UserAgent", r.UserAgent()),
				zap.Int64("ContentLength", r.ContentLength),
			}
			logger.Info(
				"http request",
				fields...)

			then := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			var respHeaders []string
			var rFields []zapcore.Field

			for k := range ww.Header() {
				respHeaders = append(respHeaders, k)
			}

			for _, h := range respHeaders {
				rFields = append(rFields, zap.String(h, ww.Header().Get(h)))
			}
			logger.Info("respHeader", rFields...)

			dur := time.Since(then)
			status := ww.Status()
			var responseFields = []zapcore.Field{
				zap.Int("Status", status),
				zap.Int("Bytes", ww.BytesWritten()),
				zap.Duration("Duration", dur),
			}
			if status < http.StatusOK || status >= http.StatusInternalServerError {
				logger.Error(
					"http response",
					responseFields...)
			} else {
				logger.Info(
					"http response",
					responseFields...)
			}
		})
	}
}

// Get initializes a zap.Logger instance if it has not been initialized
// already and returns the same instance for subsequent calls.
func Get() *zap.Logger {
	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		file := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    5,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		})

		level := zap.InfoLevel
		levelEnv := os.Getenv("LOG_LEVEL")
		if levelEnv != "" {
			levelFromEnv, err := zapcore.ParseLevel(levelEnv)
			if err != nil {
				log.Println(
					fmt.Errorf("invalid level, defaulting to INFO: %w", err),
				)
			}

			level = levelFromEnv
		}

		logLevel := zap.NewAtomicLevelAt(level)

		productionCfg := zap.NewProductionEncoderConfig()
		productionCfg.TimeKey = "timestamp"
		productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		developmentCfg := zap.NewDevelopmentEncoderConfig()
		developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

		consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
		fileEncoder := zapcore.NewJSONEncoder(productionCfg)

		var gitRevision string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					gitRevision = v.Value
					break
				}
			}
		}

		// log to multiple destinations (console and file)
		// extra fields are added to the JSON output alone
		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, stdout, logLevel),
			zapcore.NewCore(fileEncoder, file, logLevel).
				With(
					[]zapcore.Field{
						zap.String("git_revision", gitRevision),
						zap.String("go_version", buildInfo.GoVersion),
					},
				),
		)

		logger = zap.New(core)
	})

	return logger
}

// FromCtx returns the Logger associated with the ctx. If no logger
// is associated, the default logger is returned, unless it is nil
// in which case a disabled logger is returned.
func FromCtx(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	} else if l := logger; l != nil {
		return l
	}

	return zap.NewNop()
}

// WithCtx returns a copy of ctx with the Logger attached.
func WithCtx(ctx context.Context, l *zap.Logger) context.Context {
	if lp, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		if lp == l {
			// Do not store same logger.
			return ctx
		}
	}

	return context.WithValue(ctx, ctxKey{}, l)
}
