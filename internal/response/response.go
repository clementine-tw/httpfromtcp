package response

import (
	"strconv"

	"github.com/clementine-tw/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLength int, contentType string) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLength))
	h.Set("Connection", "close")
	h.Set("Content-Type", contentType)
	return h
}
