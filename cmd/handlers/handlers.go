package handlers

import (
	"fmt"
	"github.com/LorezV/url-shorter.git/cmd/storage"
	"github.com/LorezV/url-shorter.git/cmd/utils"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

func URLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		b, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "Unknown error", http.StatusBadRequest)
		}

		if len(string(b)) == 0 {
			http.Error(w, "Cant handle empty body!", http.StatusBadRequest)
		}

		id := utils.GenerateID()
		url := storage.URL{ID: id, Original: string(b), Short: fmt.Sprintf("http://%s/%s", r.Host, id)}

		if storage.Repository.Add(url) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(url.Short))
		} else {
			http.Error(w, "Can't add new url to storage.", http.StatusInternalServerError)
		}
	case http.MethodGet:
		id := chi.URLParam(r, "id")

		fmt.Println(r.URL)

		if id == "" {
			http.Error(w, "The query parameter ID is missing", http.StatusBadRequest)
			return
		}

		if url, ok := storage.Repository.Get(id); ok {
			w.Header().Set("Location", url.Original)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "Url with this id not found!", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
	}
}
