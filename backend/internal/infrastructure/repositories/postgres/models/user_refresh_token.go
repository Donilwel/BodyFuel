package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type UserRefreshTokenRow struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

func NewUserRefreshTokenRow(t *entities.UserRefreshToken) *UserRefreshTokenRow {
	return &UserRefreshTokenRow{
		ID:        t.ID(),
		UserID:    t.UserID(),
		TokenHash: t.TokenHash(),
		ExpiresAt: t.ExpiresAt(),
		CreatedAt: t.CreatedAt(),
	}
}

func (r *UserRefreshTokenRow) ToEntity() *entities.UserRefreshToken {
	return entities.NewUserRefreshToken(entities.WithUserRefreshTokenRestoreSpec(entities.UserRefreshTokenRestoreSpec{
		ID:        r.ID,
		UserID:    r.UserID,
		TokenHash: r.TokenHash,
		ExpiresAt: r.ExpiresAt,
		CreatedAt: r.CreatedAt,
	}))
}
