package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/clementine-tw/httpfromtcp/internal/headers"
	"github.com/clementine-tw/httpfromtcp/internal/request"
	"github.com/clementine-tw/httpfromtcp/internal/response"
	"github.com/clementine-tw/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {

	var code response.StatusCode
	var body []byte

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, req)
		return
	}

	switch req.RequestLine.RequestTarget {
	case "/video":
		handlerVideo(w, req)
		return

	case "/yourproblem":

		code = response.StatusBadRequest
		body = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

	case "/myproblem":
		handler500(w, req)
		return

	default:
		code = response.StatusOK
		body = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)

	}

	w.WriteStatusLine(code)
	w.WriteHeaders(response.GetDefaultHeaders(len(body), "text/html"))
	w.WriteBody(body)
}

func handler500(w *response.Writer, _ *request.Request) {
	code := response.StatusInternalServerError
	body := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)

	w.WriteStatusLine(code)
	w.WriteHeaders(response.GetDefaultHeaders(len(body), "text/html"))
	w.WriteBody(body)
}

const httpbinURL = "https://httpbin.org"

func proxyHandler(w *response.Writer, req *request.Request) {

	route := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	url := fmt.Sprintf("%s%s", httpbinURL, route)
	resp, err := http.Get(url)
	if err != nil {
		w.WriteStatusLine(response.StatusInternalServerError)
		w.WriteHeaders(response.GetDefaultHeaders(0, "text/plain"))
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOK)
	h := headers.NewHeaders()
	h.Set("Content-Type", "text/plain")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "x-content-sha256, x-content-length")
	w.WriteHeaders(h)

	hash := sha256.New()
	buf := make([]byte, 1024)
	totalBytes := 0
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, err = hash.Write(buf[:n])
			if err != nil {
				log.Printf("error writing chunked body hash: %v", err)
				break
			}
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				log.Printf("error writing chunked body: %v", err)
				break
			}
			totalBytes += n
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			log.Printf("error reading from httpbin: %v", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Printf("error writing chunked body done: %v", err)
	}

	trailer := headers.NewHeaders()
	trailer.Set("x-content-sha256", fmt.Sprintf("%x", hash.Sum(nil)))
	trailer.Set("x-content-length", fmt.Sprintf("%d", totalBytes))
	err = w.WriteTrailers(trailer)
	if err != nil {
		log.Printf("error writing trailers: %v", err)
	}
}

func handlerVideo(w *response.Writer, req *request.Request) {
	video, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		handler500(w, req)
		return
	}

	h := response.GetDefaultHeaders(len(video), "video/mp4")
	err = w.WriteStatusLine(response.StatusOK)
	if err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}
	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("error writing header: %v", err)
		return
	}
	_, err = w.WriteBody(video)
	if err != nil {
		log.Printf("error writing body: %v", err)
		return
	}
}
