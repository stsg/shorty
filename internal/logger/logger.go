package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ZapLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
