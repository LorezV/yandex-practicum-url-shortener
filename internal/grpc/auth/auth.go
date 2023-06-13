package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"github.com/LorezV/url-shorter.git/internal/config"
	"github.com/LorezV/url-shorter.git/internal/utils"
	"github.com/google/uuid"
)

// ErrUnauthorized error when unauthorized.
var ErrUnauthorized = errors.New("unauthorized")

// ErrValueIsNotUUID error when value is not uuid.
var ErrValueIsNotUUID = errors.New("value is not uuid.UUID")

// Load - loads user from string
func Load(s string) (user uuid.UUID, err error) {
	payload, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(payload) < 16 {
		return uuid.Nil, ErrUnauthorized
	}

	h := hmac.New(sha256.New, []byte(config.AppConfig.SecretKey))
	h.Write(payload[:16])
	sign := h.Sum(nil)

	if !hmac.Equal(sign, payload[16:]) {
		return uuid.Nil, ErrUnauthorized
	}

	user, err = uuid.FromBytes(payload[:16])
	if err != nil {
		return uuid.Nil, ErrUnauthorized
	}

	return user, nil
}

// Gen - generates new user
func Gen() (user uuid.UUID, signed string) {
	user = uuid.New()

	b, _ := user.MarshalBinary()

	h := hmac.New(sha256.New, []byte(config.AppConfig.SecretKey))
	h.Write(b)
	sign := h.Sum(nil)

	signed = base64.StdEncoding.EncodeToString(append(b, sign...))

	return user, signed
}

// GetUser returns user from context.
func GetUser(ctx context.Context) (user string, err error) {
	user, ok := ctx.Value(utils.ContextKey("userID")).(string)
	if !ok {
		return "", ErrValueIsNotUUID
	}

	return
}
