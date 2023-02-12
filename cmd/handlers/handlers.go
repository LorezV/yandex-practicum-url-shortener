package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"

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

	id, e := utils.GenerateID()
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	userId := r.Context().Value("userId").(string)
	url := storage.URL{ID: id, Original: string(b), Short: fmt.Sprintf("%s/%s", config.AppConfig.BaseURL, id), UserId: userId}

	if storage.Repository.Save(url) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(url.Short))
	} else {
		http.Error(w, "Can't add new url to storage.", http.StatusInternalServerError)
	}
}

func CreateURLJson(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Can't read body.", http.StatusBadRequest)
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

	id, e := utils.GenerateID()
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	userId := r.Context().Value("userId").(string)
	url := storage.URL{ID: id, Original: data.URL, Short: fmt.Sprintf("%s/%s", config.AppConfig.BaseURL, id), UserId: userId}

	if storage.Repository.Save(url) {
		type ResponseData struct {
			Result string `json:"result"`
		}

		responseBody, err := json.Marshal(ResponseData{Result: url.Short})

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

func GetUserUrls(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(string)
	b := storage.Repository.GetAllByUser(userId)

	if len(b) > 0 {
		j, err := json.Marshal(b)
		if err != nil {
			http.Error(w, "Can't marshal urls.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(j)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
