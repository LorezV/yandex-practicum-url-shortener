package middlewares

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/LorezV/url-shorter.git/cmd/config"
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

func generateRandom(size int) ([]byte, error) {
	// генерируем случайную последовательность байт
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			isCookieValid bool
			userID        string
		)

		isCookieValid = false
		cookie, err := r.Cookie("userID")

		if err == nil {
			id := cookie.Value[:8]
			sign, e := hex.DecodeString(cookie.Value[8:])
			if e != nil {
				http.Error(w, "Can't decode string.", http.StatusInternalServerError)
				return
			}
			h := hmac.New(sha256.New, []byte(config.AppConfig.SecretKey))
			h.Write([]byte(id))
			dst := h.Sum(nil)
			isCookieValid = hmac.Equal(dst, sign)
			if isCookieValid {
				userID = id
			}
		}

		if !isCookieValid || err != nil {
			id, e := utils.GenerateID()
			if e != nil {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}

			h := hmac.New(sha256.New, []byte(config.AppConfig.SecretKey))
			h.Write([]byte(id))
			dst := h.Sum(nil)
			value := id + hex.EncodeToString(dst)
			cookie = &http.Cookie{Name: "userID", Value: value, MaxAge: 36000}
			http.SetCookie(w, cookie)
			userID = id
		}

		r = r.WithContext(context.WithValue(r.Context(), utils.ContextKey("userID"), userID))
		next.ServeHTTP(w, r)
	})
}
