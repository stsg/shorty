package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

func Test_getShortURL(t *testing.T) {

	Shorty["123456"] = "https://www.google.com"

	type want struct {
		statusCode  int
		contentType string
		response    string
	}
	tests := []struct {
		name    string
		method  string
		url     string
		request string
		want    want
	}{
		{
			name:    "getShortURL #1",
			method:  http.MethodPost,
			url:     "/",
			request: "https://practicum.yandex.ru",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				response:    "",
			},
		},
		{
			name:    "getShortURL #2",
			method:  http.MethodPost,
			url:     "/",
			request: "https://ya.ru",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				response:    "",
			},
		},
		{
			name:    "getShortURL #3",
			method:  http.MethodPost,
			url:     "/",
			request: "https://ya.ru",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
				response:    "",
			},
		},
		{
			name:    "getShortURL #4",
			method:  http.MethodPost,
			url:     "/",
			request: "https://www.google.com",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
				response:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.url, strings.NewReader(test.request))
			w := httptest.NewRecorder()
			getShortURL(w, req)
			res := w.Result()
			_, err := io.ReadAll(res.Body)
			// assert.Equal(t, test.want.statusCode, res.StatusCode)
			assert.Equal(t, res.StatusCode, test.want.statusCode)
			defer res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, w.Header().Get("Content-Type"), test.want.contentType)
			require.NoError(t, res.Body.Close())
		})
	}
}

func Test_getRealURL(t *testing.T) {

	Shorty["123456"] = "https://www.google.com"

	type want struct {
		statusCode  int
		contentType string
		location    string
		response    string
	}

	tests := []struct {
		name    string
		method  string
		url     string
		request string
		want    want
	}{
		{
			name:    "getRealURL #1",
			method:  http.MethodGet,
			url:     "/654321",
			request: "/654321",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain",
				location:    "",
				response:    "",
			},
		},
		{
			name:    "getRealURL #2",
			method:  http.MethodGet,
			url:     "/123456",
			request: "/123456",
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: "text/plain",
				location:    "https://www.google.com",
				response:    "https://www.google.com",
			},
		},
		{
			name:    "getRealURL #3",
			method:  http.MethodGet,
			url:     "/djsakhjkashhsjkhsadjkhsajkhdjkashdjkashdjkhaskjdhaskjhdjkashdkjashdjkashdkjhsakdhjkashdjkashdkjashdjkhasjkdhasjkhdkjashdjkashdjkhasdjkhdasjk/",
			request: "/djsakhjkashhsjkhsadjkhsajkhdjkashdjkashdjkhaskjdhaskjhdjkashdkjashdjkashdkjhsakdhjkashdjkashdkjashdjkhasjkdhasjkhdkjashdjkashdjkhasdjkhdasjk/",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain",
				location:    "",
				response:    "",
			},
		},
		{
			name:    "getRealURL #4",
			method:  http.MethodGet,
			url:     "/djsakhjk/ashhsjkhsadjkhsajkhdjka/ashdjkashdjkha/678",
			request: "/djsakhjk/ashhsjkhsadjkhsajkhdjka/ashdjkashdjkha/678",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain",
				location:    "",
				response:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.url, strings.NewReader(test.request))
			// req := httptest.NewRequest(test.method, test.url, nil)
			w := httptest.NewRecorder()
			getRealURL(w, req)
			res := w.Result()
			_, err := io.ReadAll(res.Body)
			assert.Equal(t, res.StatusCode, test.want.statusCode)
			defer res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, w.Header().Get("Content-Type"), test.want.contentType)
			assert.Equal(t, w.Header().Get("Location"), test.want.location)
			require.NoError(t, res.Body.Close())
		})
	}
}
