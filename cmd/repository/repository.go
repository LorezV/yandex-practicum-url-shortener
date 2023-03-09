package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/LorezV/url-shorter.git/cmd/config"
	"github.com/LorezV/url-shorter.git/cmd/utils"
)

var GlobalRepository Repository

type Repository interface {
	Insert(ctx context.Context, url URL) (URL, error)
	InsertMany(ctx context.Context, urls []URL) ([]URL, error)
	Get(ctx context.Context, id string) (URL, bool)
	GetAllByUser(ctx context.Context, userID string) ([]URL, error)
}

type URL struct {
	ID       string `json:"id"`
	Original string `json:"original_url"`
	Short    string `json:"short_url"`
	UserID   string `json:"user_id"`
}

func MakeURL(original string, userID string) (URL, error) {
	url := URL{Original: original, UserID: userID}

	id, err := utils.GenerateID()
	if err != nil {
		return url, err
	}

	url.ID = id
	url.Short = fmt.Sprintf("%s/%s", config.AppConfig.BaseURL, id)

	return url, nil
}

var ErrorURLDuplicate = errors.New("url already exists")
