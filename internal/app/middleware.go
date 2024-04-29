package app

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/stsg/shorty/internal/logger"
)

// Decompress returns a middleware that decompresses request bodies if they are gzipped.
//
// It takes an http.Handler as a parameter and returns an http.Handler.
func (app *App) Decompress() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Content-Encoding") == "gzip" {
				reader, err := gzip.NewReader(req.Body)
				if err != nil {
					rw.Header().Set("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(err.Error()))
					return
				}
				defer reader.Close()

				buf := new(strings.Builder)
				_, err = io.Copy(buf, reader)
				if err != nil {
					rw.Header().Set("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(err.Error()))
					return
				}
				req.Body = io.NopCloser(strings.NewReader(buf.String()))
				req.Header.Set("Content-Length", string(rune(len(buf.String()))))
			}
			next.ServeHTTP(rw, req)
		})
	}
}

func (app *App) TrustedSubnets() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			var isURLProtected = false

			logger := logger.Get()

			for _, v := range protectedURLs {
				if strings.Contains(req.URL.Path, v) {
					isURLProtected = true
					break
				}
			}
			if !isURLProtected {
				next.ServeHTTP(rw, req)
				return
			}

			clientIP := strings.Split(req.RemoteAddr, ":")[0]
			// clientIP := net.ParseIP(req.Header.Get("X-Real-Ip"))
			logger.Debug("trusted ip", zap.String("ip", clientIP))
			if app.Config.GetTrustedSubnet() != nil && app.Config.IsTrusted(clientIP) {
				logger.Info("not trusted ip blocked", zap.String("ip", clientIP))
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusForbidden)
				rw.Write([]byte("Forbidden"))
				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}
