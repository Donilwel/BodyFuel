package entities

import (
	"github.com/google/uuid"
	"time"
)

type UserRefreshToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	tokenHash string
	expiresAt time.Time
	createdAt time.Time
}

func (t *UserRefreshToken) ID() uuid.UUID        { return t.id }
func (t *UserRefreshToken) UserID() uuid.UUID     { return t.userID }
func (t *UserRefreshToken) TokenHash() string     { return t.tokenHash }
func (t *UserRefreshToken) ExpiresAt() time.Time  { return t.expiresAt }
func (t *UserRefreshToken) CreatedAt() time.Time  { return t.createdAt }
func (t *UserRefreshToken) IsExpired() bool       { return time.Now().After(t.expiresAt) }

type UserRefreshTokenInitSpec struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

type UserRefreshTokenRestoreSpec struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type UserRefreshTokenOption func(t *UserRefreshToken)

func NewUserRefreshToken(opt UserRefreshTokenOption) *UserRefreshToken {
	t := new(UserRefreshToken)
	opt(t)
	return t
}

func WithUserRefreshTokenInitSpec(s UserRefreshTokenInitSpec) UserRefreshTokenOption {
	return func(t *UserRefreshToken) {
		t.id = s.ID
		t.userID = s.UserID
		t.tokenHash = s.TokenHash
		t.expiresAt = s.ExpiresAt
		t.createdAt = time.Now()
	}
}

func WithUserRefreshTokenRestoreSpec(s UserRefreshTokenRestoreSpec) UserRefreshTokenOption {
	return func(t *UserRefreshToken) {
		t.id = s.ID
		t.userID = s.UserID
		t.tokenHash = s.TokenHash
		t.expiresAt = s.ExpiresAt
		t.createdAt = s.CreatedAt
	}
}
