package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	var repository = MemoryRepository{storage: make(map[string]URL)}

	if len(config.AppConfig.FileStoragePath) > 0 {
		filePath, err := filepath.Abs(config.AppConfig.FileStoragePath)
		if err != nil {
			log.Fatalf("Can't get absolute path of  %s", config.AppConfig.FileStoragePath)
			return nil
		}

		repository.filePath = filePath

		err = repository.LoadFromFile()

		if err != nil {
			log.Fatalf(err.Error())
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
			return err
		}

		if ok := r.Add(url); !ok {
			return fmt.Errorf("can't pass url in memory")
		}
	}

	err = scanner.Err()
	return
}

func (r MemoryRepository) Save(url URL) (URL, error) {
	r.Add(url)
	fmt.Println(url)

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

		file.Write([]byte(string(data) + "\n"))
	}

	return url, nil
}

func (r MemoryRepository) Add(url URL) bool {
	_, ok := r.storage[url.ID]
	if !ok {
		r.storage[url.ID] = url
	}

	return !ok
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
