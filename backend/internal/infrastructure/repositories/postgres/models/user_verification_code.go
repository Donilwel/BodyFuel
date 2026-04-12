package models

import (
	"backend/internal/domain/entities"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type UserVerificationCodeRow struct {
	ID        uuid.UUID            `db:"id"`
	UserID    uuid.UUID            `db:"user_id"`
	CodeHash  string               `db:"code_hash"`
	CodeType  string               `db:"code_type"`
	ExpiresAt time.Time            `db:"expires_at"`
	UsedAt    sql.NullTime         `db:"used_at"`
	CreatedAt time.Time            `db:"created_at"`
}

func NewUserVerificationCodeRow(c *entities.UserVerificationCode) *UserVerificationCodeRow {
	row := &UserVerificationCodeRow{
		ID:        c.ID(),
		UserID:    c.UserID(),
		CodeHash:  c.CodeHash(),
		CodeType:  string(c.CodeType()),
		ExpiresAt: c.ExpiresAt(),
		CreatedAt: c.CreatedAt(),
	}
	if c.UsedAt() != nil {
		row.UsedAt = sql.NullTime{Time: *c.UsedAt(), Valid: true}
	}
	return row
}

func (r *UserVerificationCodeRow) ToEntity() *entities.UserVerificationCode {
	spec := entities.UserVerificationCodeRestoreSpec{
		ID:        r.ID,
		UserID:    r.UserID,
		CodeHash:  r.CodeHash,
		CodeType:  entities.VerificationCodeType(r.CodeType),
		ExpiresAt: r.ExpiresAt,
		CreatedAt: r.CreatedAt,
	}
	if r.UsedAt.Valid {
		spec.UsedAt = &r.UsedAt.Time
	}
	return entities.NewUserVerificationCode(entities.WithUserVerificationCodeRestoreSpec(spec))
}
