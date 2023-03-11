package middlewares

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"encoding/hex"
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

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			isCookieValid bool
			userID        string
		)

		isCookieValid = false
		cookie, err := r.Cookie("userID")

		if err == nil && len(cookie.Value) >= 12 {
			id := cookie.Value[:12]
			sign, e := hex.DecodeString(cookie.Value[12:])
			if e != nil {
				http.Error(w, "Can't decode string.", http.StatusInternalServerError)
				return
			}

			isCookieValid = hmac.Equal(utils.EncodeUserID(id), sign)
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

			value := id + hex.EncodeToString(utils.EncodeUserID(id))
			cookie = &http.Cookie{Name: "userID", Value: value, MaxAge: 36000}
			http.SetCookie(w, cookie)
			userID = id
		}

		r = r.WithContext(context.WithValue(r.Context(), utils.ContextKey("userID"), userID))
		next.ServeHTTP(w, r)
	})
}
