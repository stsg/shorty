package app

import (
	"compress/gzip"
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

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

// TrustedSubnets returns a middleware function that checks if the requested URL is protected and if the client's IP address is trusted.
//
// The function takes an http.Handler as a parameter and returns an http.Handler.
// The returned http.Handler checks if the requested URL is protected by iterating over the protectedURLs slice.
// If the requested URL is not protected, the next http.Handler is called.
// If the requested URL is protected, the client's IP address is obtained from the request's RemoteAddr field.
// The client's IP address is then checked against the trusted subnet using the IsTrusted method of the Config struct.
// If the client's IP address is not trusted, a "Forbidden" response is sent with a status code of http.StatusForbidden.
// If the client's IP address is trusted, the next http.Handler is called.
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

			clientIP := net.ParseIP(req.Header.Get("X-Real-Ip"))
			logger.Debug("trusted ip", zap.String("ip", clientIP.String()))
			if app.Config.GetTrustedSubnet() != nil && app.Config.IsTrusted(clientIP.String()) {
				logger.Info("not trusted ip blocked", zap.String("ip", clientIP.String()))
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusForbidden)
				rw.Write([]byte("Forbidden"))
				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}

// GRPCRequestLogger logs the incoming gRPC request and its status code.
//
// It takes the context, request, server info, and the handler as input parameters.
// It returns the response and error.
func GRPCRequestLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger := logger.Get()

	logger.Info("grpc request", zap.String("method", info.FullMethod))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, err := handler(ctx, req)
	status, _ := status.FromError(err)

	logger.Info("got incoming gRPC request",
		zap.String("method", info.FullMethod),
		zap.String("status code", status.Code().String()),
	)

	return resp, err
}
