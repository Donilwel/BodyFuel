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
	queryCreateUserCalories = `INSERT INTO bodyfuel.user_calories
		(id, user_id, calories, description, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	queryUpdateUserCalories = `UPDATE bodyfuel.user_calories SET
		calories   = :calories,
		description = :description,
		date        = :date,
		updated_at  = :updated_at
		WHERE id = :id AND user_id = :user_id`
)

type UserCaloriesRepo struct {
	getter dbClientGetter
}

func NewUserCaloriesRepository(db *sqlx.DB) *UserCaloriesRepo {
	return &UserCaloriesRepo{getter: dbClientGetter{db: db}}
}

func (r *UserCaloriesRepo) Create(ctx context.Context, uc *entities.UserCalories) error {
	row := models.NewUserCaloriesRow(uc)
	_, err := r.getter.Get(ctx).ExecContext(ctx, queryCreateUserCalories,
		row.ID, row.UserID, row.Calories, row.Description, row.Date, row.CreatedAt, row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}
	return nil
}

func (r *UserCaloriesRepo) Get(ctx context.Context, f dto.UserCaloriesFilter) (*entities.UserCalories, error) {
	q := psq.Select("id", "user_id", "calories", "description", "date", "created_at", "updated_at").
		From("bodyfuel.user_calories")

	if f.ID != nil {
		q = q.Where(sq.Eq{"id": *f.ID})
	}
	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.UserCaloriesRow
	if err = r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		return nil, fmt.Errorf("get context: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *UserCaloriesRepo) List(ctx context.Context, f dto.UserCaloriesFilter) ([]*entities.UserCalories, error) {
	q := psq.Select("id", "user_id", "calories", "description", "date", "created_at", "updated_at").
		From("bodyfuel.user_calories").
		OrderBy("date DESC")

	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}
	if f.StartDate != nil {
		q = q.Where(sq.GtOrEq{"date": *f.StartDate})
	}
	if f.EndDate != nil {
		q = q.Where(sq.LtOrEq{"date": *f.EndDate})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var rows []models.UserCaloriesRow
	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select context: %w", err)
	}

	result := make([]*entities.UserCalories, len(rows))
	for i := range rows {
		result[i] = rows[i].ToEntity()
	}
	return result, nil
}

func (r *UserCaloriesRepo) Update(ctx context.Context, uc *entities.UserCalories) error {
	row := models.NewUserCaloriesRow(uc)
	_, err := r.getter.Get(ctx).NamedExecContext(ctx, queryUpdateUserCalories, row)
	if err != nil {
		return fmt.Errorf("named exec context: %w", err)
	}
	return nil
}

func (r *UserCaloriesRepo) Delete(ctx context.Context, f dto.UserCaloriesFilter) error {
	q := psq.Delete("bodyfuel.user_calories")

	if f.ID != nil {
		q = q.Where(sq.Eq{"id": *f.ID})
	}
	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.getter.Get(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}
	return nil
}
