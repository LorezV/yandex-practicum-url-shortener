package utils

import (
	"encoding/hex"
	"io"
	"log"
	"math/rand"
	"net/http"
)

func GenerateID() (string, error) {
	b, err := GenerateRandom(4)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func GenerateRandom(size int) ([]byte, error) {
	// генерируем случайную последовательность байт
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
