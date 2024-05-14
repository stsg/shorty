package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/stsg/shorty/internal/storage"
)

// HandlePing handles the ping request.
//
// It takes in the http.ResponseWriter and *http.Request as parameters.
// It does not return any values.
func (app *App) HandlePing(rw http.ResponseWriter, req *http.Request) {
	ping := strings.TrimPrefix(req.URL.Path, "/")
	ping = strings.TrimSuffix(ping, "/")
	if !app.storage.IsReady() {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		http.Error(rw, "storage not ready", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(ping + " - pong"))
}

// HandleShortID handles the shortened URL request and redirects the client to the corresponding long URL.
//
// Parameters:
// - rw: http.ResponseWriter - the response writer used to write the response.
// - req: *http.Request - the HTTP request object containing the URL path.
//
// Returns: None.
func (app *App) HandleShortID(rw http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	id = strings.TrimSuffix(id, "/")
	longURL, err := app.storage.GetRealURL(id)
	if errors.Is(err, storage.ErrURLDeleted) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusGone)
		rw.Write([]byte(err.Error()))
		return
	}
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(err.Error()))
		return

	}
	rw.Header().Set("Location", longURL)
	rw.WriteHeader(http.StatusTemporaryRedirect)
	rw.Write([]byte(longURL))
}

// HandleShortRequest handles the short URL request and generates a short URL for the given long URL.
//
// The parameters are rw for http.ResponseWriter and req for http.Request. It does not return anything.
func (app *App) HandleShortRequest(rw http.ResponseWriter, req *http.Request) {
	var userID uint64
	var session string

	url, err := io.ReadAll(req.Body)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	longURL := string(url)
	if longURL == "" {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("url is empty"))
		return
	}

	userIDToken, err := req.Cookie("token")
	if err == nil {
		userID = app.Session.GetUserSessionID(userIDToken.Value)
	} else {
		session, userID = app.Session.AddUserSession()
		app.SetSession(rw, session)
	}

	shortURL, err := app.storage.GetShortURL(userID, longURL)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		if errors.Is(err, storage.ErrUniqueViolation) {
			rw.WriteHeader(http.StatusConflict)
			rw.Write([]byte(app.Config.GetBaseAddr() + "/" + shortURL))
			return
		}
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(app.Config.GetBaseAddr() + "/" + shortURL))
}

// HandleShortRequestJSON handles short request JSON and generates a short URL.
//
// Parameters:
// - rw: http.ResponseWriter for writing response.
// - req: *http.Request for incoming request.
func (app *App) HandleShortRequestJSON(rw http.ResponseWriter, req *http.Request) {
	var rqJSON storage.ReqJSON
	var rwJSON storage.ResJSON
	var userID uint64
	var session string

	url, err := io.ReadAll(req.Body)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(url, &rqJSON)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		rw.Write([]byte(body))
		return
	}

	userIDToken, err := req.Cookie("token")
	if err == nil {
		userID = app.Session.GetUserSessionID(userIDToken.Value)
	} else {
		session, userID = app.Session.AddUserSession()
		app.SetSession(rw, session)
	}

	rwJSON.Result, err = app.storage.GetShortURL(userID, rqJSON.URL)
	rwJSON.Result = app.Config.GetBaseAddr() + "/" + rwJSON.Result
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		if errors.Is(err, storage.ErrUniqueViolation) {
			rw.WriteHeader(http.StatusConflict)
			body, _ := json.Marshal(rwJSON)
			rw.Write([]byte(body))
			return
		}
		rw.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		rw.Write([]byte(body))
		return
	}
	rw.Header().Set("Location", rwJSON.Result)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	body, _ := json.Marshal(rwJSON)
	rw.Write([]byte(body))
}

// HandleShortRequestJSONBatch handles a JSON batch request and processes it accordingly.
//
// Parameters:
//
//	rw http.ResponseWriter - the http response writer for sending responses.
//	req *http.Request - the http request containing the JSON batch.
func (app *App) HandleShortRequestJSONBatch(rw http.ResponseWriter, req *http.Request) {
	var rqJSON []storage.ReqJSONBatch
	var userID uint64
	var session string

	url, err := io.ReadAll(req.Body)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(url, &rqJSON)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		rw.Write([]byte(body))
		return
	}

	userIDToken, err := req.Cookie("token")
	if err == nil {
		userID = app.Session.GetUserSessionID(userIDToken.Value)
	} else {
		session, userID = app.Session.AddUserSession()
		app.SetSession(rw, session)
	}

	rwJSON, err := app.storage.GetShortURLBatch(userID, app.Config.GetBaseAddr(), rqJSON)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		rw.Write([]byte(body))
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	body, _ := json.Marshal(rwJSON)
	rw.Write([]byte(body))
}

// HandleGetAllURLs handles the GET request to retrieve all URLs for a user.
//
// It takes in the http.ResponseWriter and http.Request as parameters.
// It does not return any value.
func (app *App) HandleGetAllURLs(rw http.ResponseWriter, req *http.Request) {
	var resJSON []storage.ResJSONURL
	var userID uint64

	userIDToken, err := req.Cookie("token")
	if err == nil {
		userID = app.Session.GetUserSessionID(userIDToken.Value)
	} else {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte(err.Error()))
		return
	}

	resJSON, err = app.storage.GetAllURLs(userID, app.Config.GetBaseAddr())
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	if len(resJSON) == 0 {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusNoContent)
		rw.Write([]byte("no content for this user"))
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	// body, _ := json.Marshal(resJSON)
	body, _ := json.MarshalIndent(resJSON, "", "    ")
	rw.Write([]byte(body))
}

// HandleDeleteURLs handles the deletion of URLs.
//
// It takes in an http.ResponseWriter and an http.Request as parameters.
// The function reads the request body to get the URLs to be deleted.
// If there is an error reading the request body, it sets the response header to "text/plain" and writes the error message with a status code of http.StatusInternal
func (app *App) HandleDeleteURLs(rw http.ResponseWriter, req *http.Request) {
	var delURLs []string
	var userID uint64

	urls, err := io.ReadAll(req.Body)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(urls, &delURLs)
	if err != nil {
		rw.Header().Set("Content-Type", "application/text")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	userIDToken, err := req.Cookie("token")
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte(err.Error()))
		return

	}
	userID = app.Session.GetUserSessionID(userIDToken.Value)

	for _, url := range delURLs {
		go func(url string, userID uint64) {
			app.delChan <- map[string]uint64{
				url: userID,
			}
		}(url, userID)

	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusAccepted)
	rw.Write([]byte("Accepted"))
}

// HandleInternalStats handles the internal stats request.
//
// It retrieves the statistics from the storage and sends them as a JSON response.
// If there is an error retrieving the statistics, it sends a JSON response with the error message.
//
// Parameters:
// - rw: http.ResponseWriter - the response writer to send the response.
// - req: *http.Request - the request object.
//
// Return type: None.
func (app *App) HandleInternalStats(rw http.ResponseWriter, req *http.Request) {
	resJSON, err := app.storage.GetStats()
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		rw.Write([]byte(body))
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	body, _ := json.MarshalIndent(resJSON, "", "    ")
	rw.Write([]byte(body))
}
