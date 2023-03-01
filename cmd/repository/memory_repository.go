package repository

import (
	"bufio"
	"encoding/json"
	"github.com/LorezV/url-shorter.git/cmd/config"
	"log"
	"os"
	"path/filepath"
)

type MemoryRepository struct {
	storage  map[string]URL
	filePath string
}

func MakeMemoryRepository() Repository {
	filePath, err := filepath.Abs(config.AppConfig.FileStoragePath)
	if err != nil {
		log.Fatalf("Can't get absolute path of  %s", config.AppConfig.FileStoragePath)
		return nil
	}
	var repository = MemoryRepository{storage: make(map[string]URL), filePath: filePath}

	if len(filePath) > 0 {
		err := repository.LoadFromFile()

		if err != nil {
			log.Fatalf("Error loading repository from file %s", config.AppConfig.FileStoragePath)
			return nil
		}
	}

	return repository
}

func (r MemoryRepository) LoadFromFile() (err error) {
	var file *os.File

	file, err = os.OpenFile(r.filePath, os.O_RDONLY, 0777)

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

		r.Save(url)
	}

	err = scanner.Err()
	return
}

func (r MemoryRepository) Save(url URL) (URL, error) {
	_, ok := r.storage[url.ID]
	if !ok {
		r.storage[url.ID] = url
	}

	if len(r.filePath) > 0 {
		file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)

		if err != nil {
			return url, err
		}

		defer file.Close()

		data, errMarshal := json.Marshal(&url)
		if errMarshal != nil {
			return url, errMarshal
		}

		file.Write(data)
	}

	return url, nil
}

func (r MemoryRepository) Get(id string) (URL, bool) {
	val, ok := r.storage[id]
	return val, ok
}

func (r MemoryRepository) GetAllByUser(userID string) ([]URL, error) {
	result := make([]URL, len(r.storage))
	i := 0

	for _, value := range r.storage {
		if value.UserID == userID {
			result[i] = value
			i++
		}
	}

	return result[:i], nil
}
