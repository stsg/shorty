package logger

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "go.uber.org/zap"
)

func TestZapLogger(t *testing.T) {

    // Create a request and response recorder
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    w := httptest.NewRecorder()

    // Create a next handler that writes "hello world"
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("hello world"))
    })

    // Create logger middleware
    loggerMiddleware := ZapLogger()

    // Wrap next handler with logger middleware
    handler := loggerMiddleware(next)

    // Execute request
    handler.ServeHTTP(w, req)

    // Assert status code and response
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "hello world", w.Body.String())
}

func TestGetLogger(t *testing.T) {

    // Get logger
    logger := Get()

    // Assert it is not nil
    assert.NotNil(t, logger)

    // Assert it is a Zap logger
    assert.IsType(t, &zap.Logger{}, logger)
}

func TestLoggerContext(t *testing.T) {

    // Create context
    ctx := context.Background()

    // Get default logger from context - should be disabled
    logger := FromCtx(ctx)
    assert.IsType(t, &zap.Logger{}, logger)
    assert.True(t, logger.Core().Enabled(zap.InfoLevel))

    // Create a real logger
    realLogger, _ := zap.NewProduction()

    // Add logger to context
    ctx = WithCtx(ctx, realLogger)

    // Get logger from context again
    logger = FromCtx(ctx)

    // Assert it is the real logger
    assert.Equal(t, realLogger, logger)
}