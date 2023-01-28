package utils

import (
	"io"
	"math/rand"
	"net/http"

	"github.com/LorezV/url-shorter.git/cmd/storage"
)

func GenerateID() string {
	runes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	id := make([]rune, 6)
	for i := range id {
		id[i] = runes[rand.Intn(len(runes))]
	}

	// Простая проверка на уникальность
	_, ok := storage.Repository.Get(string(id))
	if ok {
		return GenerateID()
	}

	return string(id)
}

type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w GzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
