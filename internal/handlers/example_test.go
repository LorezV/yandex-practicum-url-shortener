package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func ExampleGetURL() {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/1244543", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleCreateURL() {
	r, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/", strings.NewReader("yandex.lyceum"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleCreateURLJson() {
	r, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/shorten", strings.NewReader(`{"url": "yandex.lyceum"}`))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleBatchURLJson() {
	type RequestElement struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	urls := []RequestElement{
		{
			CorrelationID: "SOMEID1",
			OriginalURL:   "yandex.lyceum",
		},
		{
			CorrelationID: "SOMEID1",
			OriginalURL:   "google.com",
		},
	}

	body, err := json.Marshal(urls)
	if err != nil {
		log.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/shorten/batch", strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleGetUserUrls() {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/user/urls", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleDeleteUserUrls() {
	r, err := http.NewRequest(http.MethodDelete, "http://127.0.0.1:8080/api/user/urls", strings.NewReader(`["URLID1", "URLID2", "URLID3"]`))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleCheckPing() {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/ping", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}
