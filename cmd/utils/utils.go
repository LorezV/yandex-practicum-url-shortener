package utils

import (
	"encoding/hex"
	"io"
	"math/rand"
	"net/http"
)

type ContextKey string

func GenerateID() (string, error) {
	b, err := GenerateRandom(6)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w GzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
