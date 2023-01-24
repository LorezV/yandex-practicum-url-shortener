package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"

	"github.com/LorezV/url-shorter.git/cmd/storage"
	"github.com/LorezV/url-shorter.git/cmd/utils"
)

func CreateURL(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Can't read body!", http.StatusBadRequest)
		return
	}

	if len(string(b)) == 0 {
		http.Error(w, "Cant handle empty body!", http.StatusBadRequest)
		return
	}

	id := utils.GenerateID()
	url := storage.URL{ID: id, Original: string(b), Short: fmt.Sprintf("http://%s/%s", r.Host, id)}

	if storage.Repository.Add(url) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(url.Short))
	} else {
		http.Error(w, "Can't add new url to storage.", http.StatusInternalServerError)
	}
}

func CreateURLJson(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Can't read body!", http.StatusBadRequest)
		return
	}

	var data struct {
		URL string `json:"url"`
	}

	if err := json.Unmarshal(b, &data); err != nil {
		http.Error(w, "Cant parse json data!", http.StatusBadRequest)
		return
	}

	if len(data.URL) == 0 {
		http.Error(w, "Request json body must contain url parameter!", http.StatusBadRequest)
		return
	}

	id := utils.GenerateID()
	url := storage.URL{ID: id, Original: data.URL, Short: fmt.Sprintf("http://%s/%s", r.Host, id)}

	if storage.Repository.Add(url) {
		rBody, _ := json.Marshal(struct {
			Response string `json:"response"`
		}{Response: url.Short})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(rBody)
	} else {
		http.Error(w, "Can't add new url to storage.", http.StatusInternalServerError)
	}
}

func GetURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	fmt.Println(id)

	if id == "" {
		http.Error(w, "The query parameter ID is missing", http.StatusBadRequest)
		return
	}

	if url, ok := storage.Repository.Get(id); ok {
		w.Header().Set("Location", url.Original)
		w.WriteHeader(307)
	} else {
		http.Error(w, "URL with this id not found!", http.StatusNotFound)
	}
}
