package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/clementine-tw/httpfromtcp/internal/request"
	"github.com/clementine-tw/httpfromtcp/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		handler:  handler,
	}

	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("error accepting new connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {

	defer conn.Close()

	resp := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		resp.WriteStatusLine(response.StatusBadRequest)
		body := fmt.Appendf(nil, "error parsing request: %v", err)
		resp.WriteHeaders(response.GetDefaultHeaders(len(body), "text/plain"))
		resp.WriteBody(body)
		return
	}

	s.handler(resp, req)
}
