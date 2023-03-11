package handlers

import (
	"context"
	"encoding/json"
	"errors"
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
	url, err := repository.MakeURL(string(b), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var status = http.StatusCreated

	savedURL, err := repository.GlobalRepository.Insert(r.Context(), url)
	if err != nil {
		if errors.Is(err, repository.ErrorURLDuplicate) {
			status = http.StatusConflict
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(status)
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
	url, err := repository.MakeURL(data.URL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var status = http.StatusCreated

	savedURL, err := repository.GlobalRepository.Insert(r.Context(), url)
	if err != nil {
		if errors.Is(err, repository.ErrorURLDuplicate) {
			status = http.StatusConflict
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	type ResponseData struct {
		Result string `json:"result"`
	}

	responseBody, err := json.Marshal(ResponseData{Result: savedURL.Short})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(responseBody)
}

func GetURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "The query parameter ID is missing", http.StatusBadRequest)
		return
	}

	url, ok := repository.GlobalRepository.Get(r.Context(), id)
	if !ok {
		http.Error(w, "URL with this id not found!", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url.Original)
	w.WriteHeader(307)
}

func GetUserUrls(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(utils.ContextKey("userID")).(string)
	b, err := repository.GlobalRepository.GetAllByUser(r.Context(), userID)

	if err != nil {
		http.Error(w, "Can't get urls from repository.", http.StatusInternalServerError)
		return
	}

	if len(b) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

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
}

func CheckPing(w http.ResponseWriter, r *http.Request) {

	if len(config.AppConfig.DatabaseDsn) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
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

	correlationIDs := make([]string, len(requestData))
	urls := make([]repository.URL, len(requestData))

	for index, element := range requestData {
		url, err := repository.MakeURL(element.OriginalURL, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urls[index] = url
		correlationIDs[index] = element.CorrelationID
	}

	urls, err = repository.GlobalRepository.InsertMany(r.Context(), urls)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for index, url := range urls {
		responseData[index] = ResponseDataElement{CorrelationID: correlationIDs[index], ShortURL: url.Short}
	}

	responseBody, err := json.Marshal(responseData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseBody)
}
