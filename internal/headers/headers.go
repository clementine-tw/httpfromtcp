package headers

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	crlf = "\r\n"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	newLineIndex := bytes.Index(data, []byte(crlf))
	if newLineIndex == -1 {
		// not enough data
		return 0, false, nil
	}
	if newLineIndex == 0 {
		// end of field-line
		return 2, true, nil
	}

	s := string(data[:newLineIndex])
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, ":", 2)

	if len(parts) != 2 {
		return 0, false, fmt.Errorf("malformed header: %v", s)
	}
	// key
	key := parts[0]
	if !validToken([]byte(key)) {
		return 0, false, errors.New("invalid header token found")
	}

	// values
	values := strings.Split(parts[1], ";")
	if len(values) == 0 {
		return 0, false, errors.New("malformed header values")
	}

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		h.Set(key, trimmed)
	}

	return newLineIndex + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if v, ok := h[key]; ok {
		h[key] = v + "," + value
		return
	}
	h[key] = value
}

func (h Headers) Get(key string) string {
	key = strings.ToLower(key)
	return h[key]
}

var tokenChars = []byte{
	'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~',
}

func isTokenChar(c byte) bool {
	if c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' ||
		c >= '0' && c <= '9' {

		return true
	}

	return slices.Contains(tokenChars, c)
}

func validToken(data []byte) bool {
	for _, c := range data {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}
