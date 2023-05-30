package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/LorezV/url-shorter.git/internal/config"
	"github.com/LorezV/url-shorter.git/internal/utils"
)

// GlobalRepository it's repository variable.
var GlobalRepository Repository

// Repository is a interface with contains methods for CRUD operations with data.
type Repository interface {
	Insert(ctx context.Context, url URL) (URL, error)
	InsertMany(ctx context.Context, urls []URL) ([]URL, error)
	Get(ctx context.Context, id string) (URL, bool)
	GetAllByUser(ctx context.Context, userID string) ([]URL, error)
	DeleteManyByUser(ctx context.Context, urlIDs []string, userID string) bool
	Close() error
}

// URL entity represent database table url
type URL struct {
	ID        string `json:"id"`
	Original  string `json:"original_url"`
	Short     string `json:"short_url"`
	UserID    string `json:"user_id"`
	IsDeleted bool   `json:"-"`
}

// MakeURL is a constructor from URL type.
func MakeURL(original string, userID string) (URL, error) {
	url := URL{Original: original, UserID: userID}

	id, err := utils.GenerateID()
	if err != nil {
		return url, err
	}

	url.ID = id
	url.Short = fmt.Sprintf("%s/%s", config.AppConfig.BaseURL, id)
	url.IsDeleted = false

	return url, nil
}

// ErrorURLDuplicate is error which returning when url with id already exists in database.
var ErrorURLDuplicate = errors.New("url already exists")
