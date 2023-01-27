package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/LorezV/url-shorter.git/cmd/config"
)

var Repository URLRepository

type URLRepository struct {
	storage map[string]URL
}

func MakeRepository() URLRepository {
	repository := URLRepository{storage: make(map[string]URL)}

	//if len(config.AppConfig.FileStoragePath) > 0 {
	//	err := repository.Load()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}

	return repository
}

func (r URLRepository) Load() (err error) {
	var file *os.File
	file, err = os.OpenFile(config.AppConfig.FileStoragePath, os.O_RDONLY, 0777)

	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	if err != nil {
		return
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var url URL
		err = json.Unmarshal([]byte(scanner.Text()), &url)

		if err != nil {
			return
		}

		r.Add(url)
	}

	err = scanner.Err()
	return
}

func (r URLRepository) Save(url URL) bool {
	if len(config.AppConfig.FileStoragePath) > 0 {
		file, err := os.OpenFile(config.AppConfig.FileStoragePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)

		if err != nil {
			log.Fatalf("Can't open file by path %s", "path")
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
