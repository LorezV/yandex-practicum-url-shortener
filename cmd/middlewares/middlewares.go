package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/LorezV/url-shorter.git/cmd/utils"
)

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				io.WriteString(w, err.Error())
				return
			}
			r.Body = reader
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			//	Заменяем Writer на GzipWriter
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				io.WriteString(w, err.Error())
				return
			}
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			w = utils.GzipWriter{ResponseWriter: w, Writer: gz}
		}

		next.ServeHTTP(w, r)
	})
}
