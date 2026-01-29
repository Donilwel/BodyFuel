package postgres

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	errs "backend/internal/errors"
	"backend/internal/infrastructure/repositories/postgres/builders"
	"backend/internal/infrastructure/repositories/postgres/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
	queryCreateWorkout = `
		INSERT INTO bodyfuel.workout (
			"id",
			"user_id",
			"level",
			"status",
			"total_calories",
			"prediction_calories",
			"duration",
			"created_at",
			"updated_at"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	queryUpdateWorkout = `
		UPDATE bodyfuel.workout SET
			level = :level,
			status = :status,
			total_calories = :total_calories,
			prediction_calories = :prediction_calories,
			duration = :duration,
			updated_at = :updated_at
		WHERE id = :id
	`
)

type WorkoutRepository struct {
	getter dbClientGetter
}

func NewWorkoutRepository(db *sqlx.DB) *WorkoutRepository {
	return &WorkoutRepository{getter: dbClientGetter{db: db}}
}

func (r *WorkoutRepository) Get(ctx context.Context, f dto.WorkoutsFilter, withBlock bool) (*entities.Workout, error) {
	selectBuilder := builders.NewWorkoutSelectBuilder().WithFilterSpecification(builders.NewWorkoutFilterSpecification(f))
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}
	query, args, err := selectBuilder.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.WorkoutRow
	if err := r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserWeightNotFound
		}
		return nil, fmt.Errorf("get context: %w", err)
	}

	return row.ToEntity(), nil
}

func (r *WorkoutRepository) TopListWithLimit(ctx context.Context, f dto.WorkoutsFilter, limit int, withBlock bool) ([]*entities.Workout, error) {
	var rows []*models.WorkoutRow

	selectBuilder := builders.NewWorkoutSelectBuilder()
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}

	query, args, err := selectBuilder.WithFilterSpecification(builders.NewWorkoutFilterSpecification(f)).Limit(limit).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select context: %w", err)
	}
	result := make([]*entities.Workout, len(rows))
	for i := range rows {
		result[i] = rows[i].ToEntity()
	}

	return result, nil
}

func (r *WorkoutRepository) Create(ctx context.Context, workout *entities.Workout) error {
	row := models.NewWorkoutRow(workout)

	_, err := r.getter.Get(ctx).ExecContext(
		ctx,
		queryCreateWorkout,
		row.ID,
		row.UserID,
		row.Level,
		row.Status,
		row.PredictionCalories,
		row.Duration,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}

func (r *WorkoutRepository) Update(ctx context.Context, workout *entities.Workout) error {
	row := models.NewWorkoutRow(workout)

	res, err := r.getter.Get(ctx).NamedExecContext(ctx, queryUpdateWorkout, row)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", errs.ErrUserWeightNotFound)
	}

	if rowAffected == 0 {
		return fmt.Errorf("rows affected: %w", err)
	}

	return nil
}

func (r *WorkoutRepository) Delete(ctx context.Context, f dto.WorkoutsFilter) error {
	deleteBuilder := builders.NewWorkoutDeleteBuilder().WithFilterSpecification(builders.NewWorkoutFilterSpecification(f))
	query, args, err := deleteBuilder.ToSQL()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	result, err := r.getter.Get(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows deleted: %w", errs.ErrUserWeightAlreadyDeleted)
	}

	return nil
}
