package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid 2 headers
	headers = NewHeaders()
	data = []byte("Host: localhost:1234\r\nUser-Agent: curl/8.7.1\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:1234", headers["host"])
	assert.Equal(t, 22, n)
	require.False(t, done)
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, "curl/8.7.1", headers["user-agent"])
	assert.Equal(t, 24, n)
	require.False(t, done)

	// Test: Valid 2 headers with same key
	headers = NewHeaders()
	data = []byte("Host: localhost:1234\r\nHost: localhost:5678\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:1234", headers["host"])
	assert.Equal(t, 22, n)
	require.False(t, done)
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, "localhost:1234,localhost:5678", headers["host"])
	assert.Equal(t, 22, n)
	require.False(t, done)

	// Test: Valid 2 headers with same key and tailing ';'
	headers = NewHeaders()
	data = []byte("Host: localhost:1234;\r\nHost: localhost:5678;\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:1234", headers["host"])
	assert.Equal(t, 23, n)
	require.False(t, done)
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, "localhost:1234,localhost:5678", headers["host"])
	assert.Equal(t, 23, n)
	require.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Key is always lower case
	headers = NewHeaders()
	data = []byte("Host: localhost:1234\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:1234", headers["host"])
	assert.Equal(t, "", headers["Host"])
	assert.Equal(t, 22, n)
	assert.False(t, done)

	// Test: Invalid character
	headers = NewHeaders()
	data = []byte("H©st: localhost:1234\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
