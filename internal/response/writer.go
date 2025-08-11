package response

import (
	"fmt"
	"io"

	"github.com/clementine-tw/httpfromtcp/internal/headers"
)

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var reason string
	switch statusCode {
	case StatusOK:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	}
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, reason)
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	for name, v := range h {
		// for value := range strings.SplitSeq(v, ",") {
		// 	if _, err := fmt.Fprintf(w, "%s: %s\r\n", name, value); err != nil {
		// 		return err
		// 	}
		// }
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", name, v); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	body := fmt.Appendf(nil, "%X\r\n%s\r\n", len(p), p)
	return w.Write(body)
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	return w.WriteHeaders(h)
}
