package server

import (
	"github.com/clementine-tw/httpfromtcp/internal/request"
	"github.com/clementine-tw/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    []byte
}

type Handler func(w *response.Writer, req *request.Request)

func WriteErrorResponse(w *response.Writer, handlerError *HandlerError) error {

	err := w.WriteStatusLine(handlerError.StatusCode)
	if err != nil {
		return err
	}

	err = w.WriteHeaders(response.GetDefaultHeaders(len(handlerError.Message), "text/plain"))
	if err != nil {
		return err
	}

	_, err = w.WriteBody(handlerError.Message)
	if err != nil {
		return err
	}

	return nil
}
