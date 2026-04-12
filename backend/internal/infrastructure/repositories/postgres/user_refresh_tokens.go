package postgres

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/infrastructure/repositories/postgres/models"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	queryCreateRefreshToken = `INSERT INTO bodyfuel.user_refresh_tokens
		(id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	queryDeleteRefreshTokensByUser = `DELETE FROM bodyfuel.user_refresh_tokens WHERE user_id = $1`
)

type UserRefreshTokensRepo struct {
	getter dbClientGetter
}

func NewUserRefreshTokensRepository(db *sqlx.DB) *UserRefreshTokensRepo {
	return &UserRefreshTokensRepo{getter: dbClientGetter{db: db}}
}

func (r *UserRefreshTokensRepo) Create(ctx context.Context, t *entities.UserRefreshToken) error {
	row := models.NewUserRefreshTokenRow(t)
	_, err := r.getter.Get(ctx).ExecContext(ctx, queryCreateRefreshToken,
		row.ID, row.UserID, row.TokenHash, row.ExpiresAt, row.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *UserRefreshTokensRepo) Get(ctx context.Context, f dto.UserRefreshTokenFilter) (*entities.UserRefreshToken, error) {
	q := psq.Select("id", "user_id", "token_hash", "expires_at", "created_at").
		From("bodyfuel.user_refresh_tokens")

	if f.ID != nil {
		q = q.Where(sq.Eq{"id": *f.ID})
	}
	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}
	if f.TokenHash != nil {
		q = q.Where(sq.Eq{"token_hash": *f.TokenHash})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.UserRefreshTokenRow
	if err = r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *UserRefreshTokensRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.getter.Get(ctx).ExecContext(ctx, queryDeleteRefreshTokensByUser, userID)
	if err != nil {
		return fmt.Errorf("delete refresh tokens: %w", err)
	}
	return nil
}

func (r *UserRefreshTokensRepo) Delete(ctx context.Context, f dto.UserRefreshTokenFilter) error {
	q := psq.Delete("bodyfuel.user_refresh_tokens")
	if f.ID != nil {
		q = q.Where(sq.Eq{"id": *f.ID})
	}
	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}
	if f.TokenHash != nil {
		q = q.Where(sq.Eq{"token_hash": *f.TokenHash})
	}
	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}
	_, err = r.getter.Get(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}
