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

func (tb *TokenBlacklist) IsExpired() bool {
	return time.Now().After(tb.ExpiresAt)
}
