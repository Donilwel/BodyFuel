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

type UserRecommendationsRepo struct {
	getter dbClientGetter
}

func NewUserRecommendationsRepository(db *sqlx.DB) *UserRecommendationsRepo {
	return &UserRecommendationsRepo{getter: dbClientGetter{db: db}}
}

func (r *UserRecommendationsRepo) Create(ctx context.Context, rec *entities.UserRecommendation) error {
	row := models.NewUserRecommendationRow(rec)
	_, err := r.getter.Get(ctx).ExecContext(ctx,
		`INSERT INTO bodyfuel.user_recommendation
		(id, user_id, type, description, priority, is_read, generated_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		row.ID, row.UserID, row.Type, row.Description, row.Priority, row.IsRead, row.GeneratedAt, row.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create recommendation: %w", err)
	}
	return nil
}

func (r *UserRecommendationsRepo) Get(ctx context.Context, f dto.UserRecommendationFilter) (*entities.UserRecommendation, error) {
	q := psq.Select("id","user_id","type","description","priority","is_read","generated_at","created_at").
		From("bodyfuel.user_recommendation")
	q = applyRecFilter(q, f)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.UserRecommendationRow
	if err := r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		return nil, fmt.Errorf("get recommendation: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *UserRecommendationsRepo) List(ctx context.Context, f dto.UserRecommendationFilter) ([]*entities.UserRecommendation, error) {
	q := psq.Select("id","user_id","type","description","priority","is_read","generated_at","created_at").
		From("bodyfuel.user_recommendation").
		OrderBy("priority ASC, generated_at DESC")
	q = applyRecFilter(q, f)

	if f.Limit != nil {
		q = q.Limit(uint64(*f.Limit))
	}
	if f.Offset != nil {
		q = q.Offset(uint64(*f.Offset))
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var rows []models.UserRecommendationRow
	if err := r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("list recommendations: %w", err)
	}

	result := make([]*entities.UserRecommendation, len(rows))
	for i, row := range rows {
		result[i] = row.ToEntity()
	}
	return result, nil
}

func (r *UserRecommendationsRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.getter.Get(ctx).ExecContext(ctx,
		`UPDATE bodyfuel.user_recommendation SET is_read=true WHERE id=$1 AND user_id=$2`, id, userID,
	)
	if err != nil {
		return fmt.Errorf("mark recommendation read: %w", err)
	}
	return nil
}

func (r *UserRecommendationsRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.getter.Get(ctx).ExecContext(ctx,
		`DELETE FROM bodyfuel.user_recommendation WHERE user_id=$1`, userID,
	)
	if err != nil {
		return fmt.Errorf("delete recommendations: %w", err)
	}
	return nil
}

func applyRecFilter(q sq.SelectBuilder, f dto.UserRecommendationFilter) sq.SelectBuilder {
	if f.ID != nil {
		q = q.Where(sq.Eq{"id": *f.ID})
	}
	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}
	if f.IsRead != nil {
		q = q.Where(sq.Eq{"is_read": *f.IsRead})
	}
	return q
}
