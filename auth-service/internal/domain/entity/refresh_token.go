package entity

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TokenHash     string
	TokenFamilyID uuid.UUID
	ExpiresAt     time.Time
	IsRevoked     bool
	CreatedAt     time.Time
	RevokedAt     *time.Time
}

func NewRefreshToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) *RefreshToken {
	familyID := uuid.New()
	return &RefreshToken{
		ID:            uuid.New(),
		UserID:        userID,
		TokenHash:     tokenHash,
		TokenFamilyID: familyID,
		ExpiresAt:     expiresAt,
		IsRevoked:     false,
		CreatedAt:     time.Now(),
	}
}

func NewRefreshTokenWithFamily(userID uuid.UUID, tokenHash string, expiresAt time.Time, familyID uuid.UUID) *RefreshToken {
	return &RefreshToken{
		ID:            uuid.New(),
		UserID:        userID,
		TokenHash:     tokenHash,
		TokenFamilyID: familyID,
		ExpiresAt:     expiresAt,
		IsRevoked:     false,
		CreatedAt:     time.Now(),
	}
}

func (rt *RefreshToken) IsValid() bool {
	if rt.IsRevoked {
		return false
	}
	return time.Now().Before(rt.ExpiresAt)
}

func (rt *RefreshToken) Revoke() {
	rt.IsRevoked = true
	now := time.Now()
	rt.RevokedAt = &now
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}
