package entity

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Action    AuditAction
	IPAddress string
	UserAgent string
	Metadata  map[string]interface{}
	CreatedAt time.Time
}

type AuditAction string

const (
	AuditActionRegister        AuditAction = "register"
	AuditActionLogin           AuditAction = "login"
	AuditActionLoginFailed     AuditAction = "login_failed"
	AuditActionLogout          AuditAction = "logout"
	AuditActionTokenRefresh    AuditAction = "token_refresh"
	AuditActionPasswordChange  AuditAction = "password_change"
	AuditActionAccountLocked   AuditAction = "account_locked"
	AuditActionAccountVerified AuditAction = "account_verified"
)

func NewAuditLog(userID uuid.UUID, action AuditAction, ipAddress, userAgent string) *AuditLog {
	return &AuditLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

func (a *AuditLog) AddMetadata(key string, value interface{}) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	a.Metadata[key] = value
}
