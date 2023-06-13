package auth

import (
	"context"
	"errors"
	"github.com/LorezV/url-shorter.git/internal/utils"
)

// ErrUnauthorized error when unauthorized.
var ErrUnauthorized = errors.New("unauthorized")

// ErrValueIsNotUUID error when value is not uuid.
var ErrValueIsNotUUID = errors.New("value is not uuid.UUID")

// GetUser returns user from context.
func GetUser(ctx context.Context) (user string, err error) {
	user, ok := ctx.Value(utils.ContextKey("userID")).(string)
	if !ok {
		return "", ErrValueIsNotUUID
	}

	return
}
