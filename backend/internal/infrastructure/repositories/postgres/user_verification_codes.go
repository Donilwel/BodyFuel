package postgres

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/infrastructure/repositories/postgres/models"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const (
	queryCreateVerificationCode = `INSERT INTO bodyfuel.user_verification_codes
		(id, user_id, code_hash, code_type, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	queryMarkVerificationCodeUsed = `UPDATE bodyfuel.user_verification_codes
		SET used_at = NOW() WHERE id = $1`
)

type UserVerificationCodesRepo struct {
	getter dbClientGetter
}

func NewUserVerificationCodesRepository(db *sqlx.DB) *UserVerificationCodesRepo {
	return &UserVerificationCodesRepo{getter: dbClientGetter{db: db}}
}

func (r *UserVerificationCodesRepo) Create(ctx context.Context, c *entities.UserVerificationCode) error {
	row := models.NewUserVerificationCodeRow(c)
	_, err := r.getter.Get(ctx).ExecContext(ctx, queryCreateVerificationCode,
		row.ID, row.UserID, row.CodeHash, row.CodeType, row.ExpiresAt, row.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create verification code: %w", err)
	}
	return nil
}

func (r *UserVerificationCodesRepo) GetLatest(ctx context.Context, f dto.UserVerificationCodeFilter) (*entities.UserVerificationCode, error) {
	q := psq.Select("id", "user_id", "code_hash", "code_type", "expires_at", "used_at", "created_at").
		From("bodyfuel.user_verification_codes").
		OrderBy("created_at DESC").
		Limit(1)

	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}
	if f.CodeType != nil {
		q = q.Where(sq.Eq{"code_type": string(*f.CodeType)})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.UserVerificationCodeRow
	if err = r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		return nil, fmt.Errorf("get verification code: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *UserVerificationCodesRepo) MarkUsed(ctx context.Context, id interface{}) error {
	_, err := r.getter.Get(ctx).ExecContext(ctx, queryMarkVerificationCodeUsed, id)
	if err != nil {
		return fmt.Errorf("mark code used: %w", err)
	}
	return nil
}
