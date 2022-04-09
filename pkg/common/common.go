package common

import (
	"errors"
	"io"
)

const (
	GlobalRequestRetries = 3
	PGDatetimeFmt        = `2006-01-02 15:04:05`
)

var (
	ErrUnauthenticated      = errors.New("err user failed to authenticate")
	ErrInvalidSigningMethod = errors.New("err invalid signing method")
	ErrInvalidAccessToken   = errors.New("err invalid access token")
	ErrInvalidPhoneNumber   = errors.New("err invalid phone number")
	ErrPhoneNotFound        = errors.New("err phone not found")
)

type CountingReader struct {
	Reader    io.Reader
	BytesRead int
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.BytesRead += n
	return n, err
}
