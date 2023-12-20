package handle

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/stsg/shorty/internal/config"
	"github.com/stsg/shorty/internal/storage"
)

type Handle struct {
	config  *config.Config
	storage storage.MapStorage
}

func NewHandle(config *config.Config, storage storage.MapStorage) Handle {
	hndl := Handle{}
	hndl.config = config
	hndl.storage = storage

	return hndl
}

func (h *Handle) HandleShortID(rw http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	id = strings.TrimSuffix(id, "/")
	lurl, err := h.storage.GetRealURL(id)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(err.Error()))

	}
	rw.WriteHeader(http.StatusTemporaryRedirect)
	rw.Write([]byte(lurl))
}

func (h *Handle) HandleShortRequest(rw http.ResponseWriter, req *http.Request) {
	url, err := io.ReadAll(req.Body)
	if err != nil {
		panic(errors.New("cannot read request body"))
	}
	lurl := string(url)
	surl, err := h.storage.GetShortURL(lurl)
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))

	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte("http://" + h.config.GetBaseAddr() + "/" + surl))
}
