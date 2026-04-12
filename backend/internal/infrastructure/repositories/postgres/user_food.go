package postgres

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/infrastructure/repositories/postgres/models"
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserFoodRepo struct {
	getter dbClientGetter
}

func NewUserFoodRepository(db *sqlx.DB) *UserFoodRepo {
	return &UserFoodRepo{getter: dbClientGetter{db: db}}
}

func (r *UserFoodRepo) Create(ctx context.Context, f *entities.UserFood) error {
	row := models.NewUserFoodRow(f)
	_, err := r.getter.Get(ctx).ExecContext(ctx,
		`INSERT INTO bodyfuel.user_food
		(id, user_id, description, calories, protein, carbs, fat, meal_type, photo_url, date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		row.ID, row.UserID, row.Description, row.Calories, row.Protein,
		row.Carbs, row.Fat, row.MealType, row.PhotoURL, row.Date, row.CreatedAt, row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user food: %w", err)
	}
	return nil
}

func (r *UserFoodRepo) Get(ctx context.Context, f dto.UserFoodFilter) (*entities.UserFood, error) {
	q := psq.Select("id","user_id","description","calories","protein","carbs","fat","meal_type","photo_url","date","created_at","updated_at").
		From("bodyfuel.user_food")
	q = applyUserFoodFilter(q, f)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.UserFoodRow
	if err := r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		return nil, fmt.Errorf("get user food: %w", err)
	}
	return row.ToEntity(), nil
}

func (r *UserFoodRepo) List(ctx context.Context, f dto.UserFoodFilter) ([]*entities.UserFood, error) {
	q := psq.Select("id","user_id","description","calories","protein","carbs","fat","meal_type","photo_url","date","created_at","updated_at").
		From("bodyfuel.user_food").
		OrderBy("date DESC, created_at DESC")
	q = applyUserFoodFilter(q, f)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var rows []models.UserFoodRow
	if err := r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("list user food: %w", err)
	}

	result := make([]*entities.UserFood, len(rows))
	for i, row := range rows {
		result[i] = row.ToEntity()
	}
	return result, nil
}

func (r *UserFoodRepo) Update(ctx context.Context, f *entities.UserFood) error {
	row := models.NewUserFoodRow(f)
	_, err := r.getter.Get(ctx).ExecContext(ctx,
		`UPDATE bodyfuel.user_food SET
		description=$1, calories=$2, protein=$3, carbs=$4, fat=$5,
		meal_type=$6, photo_url=$7, date=$8, updated_at=$9
		WHERE id=$10 AND user_id=$11`,
		row.Description, row.Calories, row.Protein, row.Carbs, row.Fat,
		row.MealType, row.PhotoURL, row.Date, time.Now(), row.ID, row.UserID,
	)
	if err != nil {
		return fmt.Errorf("update user food: %w", err)
	}
	return nil
}

func (r *UserFoodRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.getter.Get(ctx).ExecContext(ctx,
		`DELETE FROM bodyfuel.user_food WHERE id=$1 AND user_id=$2`, id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete user food: %w", err)
	}
	return nil
}

func applyUserFoodFilter(q sq.SelectBuilder, f dto.UserFoodFilter) sq.SelectBuilder {
	if f.ID != nil {
		q = q.Where(sq.Eq{"id": *f.ID})
	}
	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}
	if f.MealType != nil {
		q = q.Where(sq.Eq{"meal_type": *f.MealType})
	}
	if f.Date != nil {
		q = q.Where(sq.Eq{"date::date": f.Date.Format("2006-01-02")})
	}
	if f.StartDate != nil {
		q = q.Where(sq.GtOrEq{"date": *f.StartDate})
	}
	if f.EndDate != nil {
		q = q.Where(sq.LtOrEq{"date": *f.EndDate})
	}
	return q
}
