package entity

import (
	"time"

	"github.com/google/uuid"
)

type TokenBlacklist struct {
	ID        uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewTokenBlacklist(tokenHash string, expiresAt time.Time) *TokenBlacklist {
	return &TokenBlacklist{
		ID:        uuid.New(),
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}
