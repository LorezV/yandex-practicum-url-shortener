package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/LorezV/url-shorter.git/cmd/config"
	"log"
	"os"
)

var Repository URLRepository

type URLRepository struct {
	storage map[string]URL
}

func MakeRepository() URLRepository {
	repository := URLRepository{storage: make(map[string]URL)}

	if len(config.AppConfig.FileStoragePath) > 0 {
		repository.Load()
	}

	return repository
}

func (r URLRepository) Load() {
	file, err := os.OpenFile(config.AppConfig.FileStoragePath, os.O_RDONLY, 644)

	defer file.Close()

	if err != nil {
		log.Fatal(fmt.Sprintf("Can't open file by path %s", config.AppConfig.FileStoragePath))
		return
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var url URL

		if json.Unmarshal([]byte(scanner.Text()), &url) != nil {
			log.Fatal(fmt.Sprintf("Can't unmarshal %s", scanner.Text()))
			return
		}

		if !r.Add(url) {
			log.Fatal("Can't add url to repository.")
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (r URLRepository) Save(url URL) bool {
	if len(config.AppConfig.FileStoragePath) > 0 {
		file, err := os.OpenFile(config.AppConfig.FileStoragePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 644)

		if err != nil {
			log.Fatal(fmt.Sprintf("Can't open file by path %s", "path"))
			return false
		}

		defer file.Close()

		data, errMarshal := json.Marshal(&url)
		if errMarshal != nil {
			log.Fatal(err)
			return false
		}

		file.Write(data)
	}

	return r.Add(url)
}

func (r URLRepository) Get(id string) (URL, bool) {
	val, ok := r.storage[id]
	return val, ok
}

func (r URLRepository) Add(url URL) bool {
	_, ok := r.storage[url.ID]
	if !ok {
		r.storage[url.ID] = url
	}
	return !ok
}

type URL struct {
	ID       string
	Original string
	Short    string
}
