package handlers

import (
	"context"
	"encoding/json"
	"github.com/LorezV/url-shorter.git/cmd/config"
	"github.com/LorezV/url-shorter.git/cmd/repository"
	"github.com/LorezV/url-shorter.git/cmd/utils"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"time"
)

func CreateURL(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(string(b)) == 0 {
		http.Error(w, "Cant handle empty body!", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(utils.ContextKey("userID")).(string)
	url, e := repository.MakeURL(string(b), userID)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	savedURL, saveError := repository.GlobalRepository.Save(url)
	if saveError != nil {
		http.Error(w, saveError.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(savedURL.Short))
}

func CreateURLJson(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(string(b)) == 0 {
		http.Error(w, "Can't handle empty body.", http.StatusBadRequest)
		return
	}

	var data struct {
		URL string `json:"url"`
	}

	err = json.Unmarshal(b, &data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(data.URL) == 0 {
		http.Error(w, "Can't handle empty url parameter in body.", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(utils.ContextKey("userID")).(string)
	url, e := repository.MakeURL(data.URL, userID)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	savedURL, saveError := repository.GlobalRepository.Save(url)
	if saveError != nil {
		http.Error(w, saveError.Error(), http.StatusInternalServerError)
		return
	}

	type ResponseData struct {
		Result string `json:"result"`
	}

	responseBody, marshalError := json.Marshal(ResponseData{Result: savedURL.Short})

	if marshalError != nil {
		http.Error(w, marshalError.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseBody)
}

func GetURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "The query parameter ID is missing", http.StatusBadRequest)
		return
	}

	if url, ok := repository.GlobalRepository.Get(id); ok {
		w.Header().Set("Location", url.Original)
		w.WriteHeader(307)
	} else {
		http.Error(w, "URL with this id not found!", http.StatusNotFound)
	}
}

func GetUserUrls(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(utils.ContextKey("userID")).(string)
	b, e := repository.GlobalRepository.GetAllByUser(userID)

	if e != nil {
		http.Error(w, "Can't get urls from repository.", http.StatusInternalServerError)
		return
	}

	if len(b) > 0 {
		type ResponseElement struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}
		v := make([]ResponseElement, len(b))

		for index, url := range b {
			v[index] = ResponseElement{OriginalURL: url.Original, ShortURL: url.Short}
		}

		j, err := json.Marshal(v)
		if err != nil {
			http.Error(w, "Can't marshal json.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(j)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func CheckPing(w http.ResponseWriter, r *http.Request) {

	if len(config.AppConfig.DatabaseDsn) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := config.DB.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func BatchURLJson(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var requestData []struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	err = json.Unmarshal(b, &requestData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(requestData) == 0 {
		http.Error(w, "Can't handle empty url array in body.", http.StatusBadRequest)
		return
	}

	type ResponseDataElement struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	var responseData = make([]ResponseDataElement, len(requestData))

	userID := r.Context().Value(utils.ContextKey("userID")).(string)

	for index, element := range requestData {
		var (
			error error
			url   repository.URL
		)
		url, error = repository.MakeURL(element.OriginalURL, userID)
		if error != nil {
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}

		url, error = repository.GlobalRepository.Save(url)
		if error != nil {
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}

		responseData[index] = ResponseDataElement{CorrelationID: element.CorrelationID, ShortURL: url.Short}
	}

	responseBody, marshalError := json.Marshal(responseData)

	if marshalError != nil {
		http.Error(w, marshalError.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseBody)
}
