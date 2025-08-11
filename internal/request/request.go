package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/clementine-tw/httpfromtcp/internal/headers"
)

const (
	bufferSize = 8
	crlf       = "\r\n"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	state       requestState
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	req := &Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
	}
	buf := make([]byte, bufferSize)
	readToIndex := 0

	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			// buf is full, create new one with double length
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, errors.New("incompleted request")
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if numBytesParsed == 0 {
			continue
		}
		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	return req, nil
}

func (req *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for req.state != requestStateDone {
		n, err := req.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (req *Request) parseSingle(data []byte) (int, error) {
	switch req.state {
	case requestStateInitialized:
		requestLine, numBytesParsed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytesParsed == 0 {
			// need more data to parse
			return 0, nil
		}
		req.RequestLine = *requestLine
		req.state = requestStateParsingHeaders
		return numBytesParsed, nil

	case requestStateParsingHeaders:
		numBytesParsed, done, err := req.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			req.state = requestStateParsingBody
		}
		return numBytesParsed, nil

	case requestStateParsingBody:
		lengthString := req.Headers.Get("content-length")
		if lengthString == "" || lengthString == "0" {
			req.state = requestStateDone
			return 0, nil
		}
		length, err := strconv.Atoi(lengthString)
		if err != nil {
			return 0, err
		}
		if len(data) < length {
			return 0, nil
		}
		req.Body = make([]byte, length)
		copy(req.Body, data[:length])
		req.state = requestStateDone
		return length, nil

	case requestStateDone:
		return 0, errors.New("error: trying to read data in done state")
	default:
		return 0, errors.New("error: unknown state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	n := bytes.Index(data, []byte(crlf))
	if n == -1 {
		return nil, 0, nil
	}

	s := string(data[:n])
	req, err := requestLineFromString(s)
	if err != nil {
		return nil, 0, err
	}

	return req, n + 2, nil
}

func requestLineFromString(s string) (*RequestLine, error) {

	parts := strings.Split(s, " ")
	if len(parts) != 3 {
		return nil, errors.New("invalid number of parts in request line")
	}
	if parts[0] != strings.ToUpper(parts[0]) {
		return nil, errors.New("invalid method")
	}
	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, errors.New("invalid version format")
	}
	if versionParts[1] != "1.1" {
		return nil, errors.New("invalid version")
	}

	return &RequestLine{
		HttpVersion:   versionParts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}, nil
}
