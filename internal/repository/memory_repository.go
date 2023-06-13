package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/LorezV/url-shorter.git/internal/config"
	"log"
	"os"
	"path/filepath"
)

// MemoryRepository is Repository implementation for working with Urls in memory and file.
type MemoryRepository struct {
	storage  map[string]URL
	filePath string
}

// MakeMemoryRepository is constructor for MemoryRepository.
func MakeMemoryRepository() Repository {
	var repository = MemoryRepository{storage: make(map[string]URL)}

	if len(config.AppConfig.FileStoragePath) > 0 {
		filePath, err := filepath.Abs(config.AppConfig.FileStoragePath)
		if err != nil {
			log.Fatalf("Can't get absolute path of  %s", config.AppConfig.FileStoragePath)
			return nil
		}

		repository.filePath = filePath

		//err = repository.LoadFromFile()
		//
		//if err != nil {
		//	log.Fatalf(err.Error())
		//	return nil
		//}
	}

	return repository
}

// DeleteManyByUser is constructor for MemoryRepository.
func (r MemoryRepository) DeleteManyByUser(ctx context.Context, urlIDs []string, userID string) bool {
	for _, id := range urlIDs {
		if url, ok := r.Get(ctx, id); ok && url.UserID == userID {
			url.IsDeleted = true
			r.storage[url.ID] = url
		}
	}

	return true
}

// LoadFromFile loads Urls from file.
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

		if ok := r.Add(context.Background(), url); !ok {
			return fmt.Errorf("can't pass url in memory")
		}
	}

	err = scanner.Err()
	return
}

// Insert adds row in file storage.
func (r MemoryRepository) Insert(context context.Context, url URL) (URL, error) {
	r.Add(context, url)

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

// InsertMany adds many rows in file storage.
func (r MemoryRepository) InsertMany(context context.Context, urls []URL) ([]URL, error) {
	var (
		rawData = ""
		err     error
	)

	for _, url := range urls {
		r.Add(context, url)

		data, err := json.Marshal(&url)
		if err != nil {
			return urls, err
		}

		rawData += string(data) + "\n"
	}

	if len(r.filePath) > 0 {
		file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)

		if err != nil {
			return urls, err
		}

		defer file.Close()

		file.Write([]byte(rawData))
	}

	return urls, err
}

// Add adds url in memory.
func (r MemoryRepository) Add(context context.Context, url URL) bool {
	_, ok := r.storage[url.ID]
	if !ok {
		r.storage[url.ID] = url
	}

	return !ok
}

// Get select row by id from file storage.
func (r MemoryRepository) Get(context context.Context, id string) (URL, bool) {
	val, ok := r.storage[id]
	return val, ok
}

// GetAllByUser select many rows by user_id from file storage.
func (r MemoryRepository) GetAllByUser(context context.Context, userID string) ([]URL, error) {
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

// Close prints close
func (r MemoryRepository) Close() error {
	fmt.Println("Close memory repository")
	return nil
}

// GetStats gets stats from memory repository.
func (r MemoryRepository) GetStats(ctx context.Context) (Stats, error) {
	userIDs := make(map[string]string)

	for _, v := range r.storage {
		userIDs[v.UserID] = v.UserID
	}

	return Stats{
		Urls:  len(r.storage),
		Users: len(userIDs),
	}, nil
}
