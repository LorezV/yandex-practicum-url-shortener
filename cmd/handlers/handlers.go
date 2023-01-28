package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/LorezV/url-shorter.git/cmd/config"
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
	url := storage.URL{ID: id, Original: string(b), Short: fmt.Sprintf("%s/%s", config.AppConfig.BaseURL, id)}

	if storage.Repository.Save(url) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(url.Short))
	} else {
		http.Error(w, "Can't add new url to storage.", http.StatusInternalServerError)
	}
}

func CreateURLJson(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Can,t read body.", http.StatusBadRequest)
		return
	}

	if len(string(b)) == 0 {
		http.Error(w, "Can't handle empty body.", http.StatusBadRequest)
		return
	}

	var data struct {
		URL string `json:"url"`
	}

	jsonErr := json.Unmarshal(b, &data)

	if jsonErr != nil {
		http.Error(w, "Can't unmarshal json from body.", http.StatusBadRequest)
		return
	}

	if len(data.URL) == 0 {
		http.Error(w, "Can't handle empty url parameter in body.", http.StatusBadRequest)
		return
	}

	id := utils.GenerateID()
	url := storage.URL{ID: id, Original: data.URL, Short: fmt.Sprintf("%s/%s", config.AppConfig.BaseURL, id)}

	if storage.Repository.Save(url) {
		type ResponseData struct {
			Result string `json:"result"`
		}

		responseBody, err := json.Marshal(ResponseData{Result: url.Short})
		responseBody = []byte(fmt.Sprintf("%s\n%s\n%s", responseBody, responseBody, responseBody))

		if err != nil {
			http.Error(w, "Can't send response.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responseBody)
	}
}

func GetURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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
