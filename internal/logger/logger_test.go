package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// Logs HTTP requests and responses using Zap logger.
func TestZapLogger_LogsRequestsAndResponses(t *testing.T) {
	// Initialize a mock logger
	// logger := zap.NewExample()
	observedZapCore, observerLogs := observer.New(zap.InfoLevel)
	logger := zap.New(observedZapCore)

	// Create a request with headers
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	// Create a response writer
	rw := httptest.NewRecorder()

	// Create a handler function
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create the ZapLogger middleware
	middleware := ZapLogger(logger)

	// Invoke the middleware with the handler
	middleware(handler).ServeHTTP(rw, req)

	logs := observerLogs.All()
	assert.Equal(t, 4, len(logs), "There should be 4 logs")

	rLog0 := logs[0]
	assert.Equal(t, "header", rLog0.Message)
	assert.Equal(t, 2, len(rLog0.Context))
	if rLog0.Context[0].Key == "Authorization" {
		assert.Equal(t, "Content-Type", rLog0.Context[1].Key)
	} else {
		assert.Equal(t, "Authorization", rLog0.Context[1].Key)
	}
	assert.Equal(t, "Authorization", rLog0.Context[1].Key)
	assert.Equal(t, "Content-Type", rLog0.Context[0].Key)

	rLog1 := logs[1]
	assert.Equal(t, "http request", rLog1.Message)
	assert.Equal(t, 7, len(rLog1.Context))
	assert.Equal(t, "Method", rLog1.Context[0].Key)
	assert.Equal(t, "Host", rLog1.Context[1].Key)
	assert.Equal(t, "RequestURI", rLog1.Context[2].Key)

	rLog2 := logs[2]
	assert.Equal(t, "respHeader", rLog2.Message)
	assert.Equal(t, 0, len(rLog2.Context))

	rLog3 := logs[3]
	assert.Equal(t, "http response", rLog3.Message)
	assert.Equal(t, 3, len(rLog3.Context))
	assert.Equal(t, "Status", rLog3.Context[0].Key)
	assert.Equal(t, "Bytes", rLog3.Context[1].Key)
	assert.Equal(t, "Duration", rLog3.Context[2].Key)
}

// Handles HTTP requests with no headers.
func TestZapLogger_HandlesRequestsWithNoHeaders(t *testing.T) {
	// Initialize a mock logger
	// logger := zap.NewExample()
	observedZapCore, observerLogs := observer.New(zap.InfoLevel)
	logger := zap.New(observedZapCore)

	// Create a request with no headers
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response writer
	rw := httptest.NewRecorder()

	// Create a handler function
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create the ZapLogger middleware
	middleware := ZapLogger(logger)

	// Invoke the middleware with the handler
	middleware(handler).ServeHTTP(rw, req)

	// Verify that the logger recorded the request and response
	logs := observerLogs.All()
	assert.Equal(t, 4, len(logs))

	// Verify the request log
	requestLog := logs[0]
	assert.Equal(t, "header", requestLog.Message)
	assert.Equal(t, 0, len(requestLog.Context))

	// Verify the response log

	rLog0 := logs[0]
	assert.Equal(t, "header", rLog0.Message)
	assert.Equal(t, 0, len(rLog0.Context))

	rLog1 := logs[1]
	assert.Equal(t, "http request", rLog1.Message)
	assert.Equal(t, 7, len(rLog1.Context))
	assert.Equal(t, "Method", rLog1.Context[0].Key)
	assert.Equal(t, "Host", rLog1.Context[1].Key)
	assert.Equal(t, "RequestURI", rLog1.Context[2].Key)

	rLog2 := logs[2]
	assert.Equal(t, "respHeader", rLog2.Message)
	assert.Equal(t, 0, len(rLog2.Context))

	rLog3 := logs[3]
	assert.Equal(t, "http response", rLog3.Message)
	assert.Equal(t, 3, len(rLog3.Context))
	assert.Equal(t, "Status", rLog3.Context[0].Key)
	assert.Equal(t, "Bytes", rLog3.Context[1].Key)
	assert.Equal(t, "Duration", rLog3.Context[2].Key)
}
