package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	// "github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/assert"
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
				contentType: "text/plain; charset=utf-8",
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
				contentType: "text/plain; charset=utf-8",
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

	handler := http.HandlerFunc(getShortURL)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = test.method
			req.SetBody(test.request)
			req.URL = srv.URL

			resp, err := req.Send()
			assert.NoError(t, err, "Error making HTTP request!")

			assert.Equal(t, test.want.statusCode, resp.StatusCode())
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, resp.Header().Get("Content-Type"))
			}
			if test.want.response != "" {
				assert.Equal(t, test.want.response, string(resp.Body()))
			}
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
				response:    "",
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

	handler := http.HandlerFunc(getRealURL)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := resty.New()
			client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}))
			req := client.R()
			req.Method = test.method
			req.URL = srv.URL + test.request
			resp, err := req.Send()

			assert.NoError(t, err, "Error making HTTP request!")
			assert.Equal(t, test.want.statusCode, resp.StatusCode())
			assert.Equal(t, test.want.contentType, resp.Header().Get("Content-Type"))
			if test.want.response != "" {
				assert.Equal(t, test.want.response, string(resp.Body()))
			}
		})
	}
}
