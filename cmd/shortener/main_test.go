package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/handle"
	"github.com/stsg/shorty/internal/storage"

	// "github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/assert"
)

var conf config.Config

func Test_getShortURL(t *testing.T) {
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

	conf = config.NewConfig()
	strg := storage.NewMapStorage()

	// for testing
	strg.SetShorURL("123456", "https://www.google.com")

	hndl := handle.NewHandle(conf, *strg)

	handler := http.HandlerFunc(hndl.HandleShortRequest)
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
				statusCode:  http.StatusNotFound,
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
				contentType: "text/plain; charset=utf-8",
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
				statusCode:  http.StatusNotFound,
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
				statusCode:  http.StatusNotFound,
				contentType: "text/plain",
				location:    "",
				response:    "",
			},
		},
	}

	//conf = config.NewConfig()
	strg := storage.NewMapStorage()

	// for testing
	strg.SetShorURL("123456", "https://www.google.com")

	hndl := handle.NewHandle(conf, *strg)

	handler := http.HandlerFunc(hndl.HandleShortID)
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

func Test_getShortURLJSON(t *testing.T) {
	type reqJSON struct {
		URL string `json:"url,omitempty"`
	}

	type resJSON struct {
		Result string `json:"result"`
	}

	type want struct {
		statusCode  int
		contentType string
		response    resJSON
	}

	tests := []struct {
		name    string
		method  string
		url     string
		request reqJSON
		want    want
	}{
		{
			name:   "getShortURLJSON #1",
			method: http.MethodPost,
			url:    "/api/shorten",
			request: reqJSON{
				URL: "https://kutt.su",
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				response: resJSON{
					Result: ``,
				},
			},
		},
		{
			name:   "getShortURLJSON #2",
			method: http.MethodPost,
			url:    "/api/shorten",
			request: reqJSON{
				URL: "https://ya.ru",
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				response: resJSON{
					Result: ``,
				},
			},
		},
		{
			name:   "getShortURLJSON #3",
			method: http.MethodPost,
			url:    "/api/shorten",
			request: reqJSON{
				URL: "https://ya.ru",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response: resJSON{
					Result: ``,
				},
			},
		},
		{
			name:   "getShortURLJSON #4",
			method: http.MethodPost,
			url:    "/api/shorten",
			request: reqJSON{
				URL: "https://www.google.com",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response: resJSON{
					Result: ``,
				},
			},
		},
	}

	// conf = config.NewConfig()
	strg := storage.NewMapStorage()

	// for testing
	strg.SetShorURL("123456", "https://www.google.com")

	hndl := handle.NewHandle(conf, *strg)

	handler := http.HandlerFunc(hndl.HandleShortRequestJSON)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	var rsJSON resJSON

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = test.method
			body, _ := json.Marshal(test.request)
			req.SetBody(body)
			req.URL = srv.URL

			resp, err := req.Send()
			assert.NoError(t, err, "Error making HTTP request!")

			assert.Equal(t, test.want.statusCode, resp.StatusCode())
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, resp.Header().Get("Content-Type"))
			}
			if test.want.response.Result != "" {
				err = json.Unmarshal(resp.Body(), &rsJSON)
				assert.NoError(t, err, "Error unmarshal response")
				assert.Equal(t, test.want.response, rsJSON)
			}
		})
	}
}
