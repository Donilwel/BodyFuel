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
	queryCreateWorkoutsExercise = `INSERT INTO bodyfuel.workouts_exercise (
		"workout_id",
		"exercise_id",
		"modify_reps",
		"modify_relax_time",
		"calories",
		"status",
		"updated_at"
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	queryUpdateWorkoutsExercise = `UPDATE bodyfuel.workouts_exercise SET
		"modify_reps" = :modify_reps,
		"modify_relax_time" = :modify_relax_time,
		"calories" = :calories,
		"status" = :status,
		"updated_at" = :updated_at
	WHERE workout_id = :workout_id AND exercise_id = :exercise_id`
)

type WorkoutsExerciseRepo struct {
	getter dbClientGetter
}

func NewWorkoutsExerciseRepository(db *sqlx.DB) *WorkoutsExerciseRepo {
	return &WorkoutsExerciseRepo{getter: dbClientGetter{db: db}}
}

func (r *WorkoutsExerciseRepo) Get(ctx context.Context, f dto.WorkoutsExerciseFilter, withBlock bool) (*entities.WorkoutsExercise, error) {
	selectBuilder := builders.NewWorkoutsExerciseSelectBuilder().
		WithFilterSpecification(builders.NewWorkoutsExerciseFilterSpecification(&f))

	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.WorkoutsExerciseRow
	if err := r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrWorkoutsExerciseNotFound
		}
		return nil, fmt.Errorf("get context: %w", err)
	}

	return row.ToEntity(), nil
}

func (r *WorkoutsExerciseRepo) List(ctx context.Context, f dto.WorkoutsExerciseFilter, withBlock bool) ([]*entities.WorkoutsExercise, error) {
	var rows []*models.WorkoutsExerciseRow

	selectBuilder := builders.NewWorkoutsExerciseSelectBuilder().
		WithFilterSpecification(builders.NewWorkoutsExerciseFilterSpecification(&f))

	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select context: %w", err)
	}

	result := make([]*entities.WorkoutsExercise, len(rows))
	for i := range rows {
		result[i] = rows[i].ToEntity()
	}

	return result, nil
}

func (r *WorkoutsExerciseRepo) Create(ctx context.Context, workoutsExercise *entities.WorkoutsExercise) error {
	row := models.NewWorkoutsExerciseRow(workoutsExercise)

	_, err := r.getter.Get(ctx).ExecContext(ctx, queryCreateWorkoutsExercise,
		row.WorkoutID,
		row.ExerciseID,
		row.ModifyReps,
		row.ModifyRelaxTime,
		row.Calories,
		row.Status,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}
	return nil
}

func (r *WorkoutsExerciseRepo) Update(ctx context.Context, workoutsExercise *entities.WorkoutsExercise) error {
	row := models.NewWorkoutsExerciseRow(workoutsExercise)

	res, err := r.getter.Get(ctx).NamedExecContext(ctx, queryUpdateWorkoutsExercise, row)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rows affected: %w", errs.ErrWorkoutsExerciseNotFound)
	}

	return nil
}

func (r *WorkoutsExerciseRepo) Delete(ctx context.Context, f dto.WorkoutsExerciseFilter) error {
	deleteBuilder := builders.NewWorkoutsExerciseDeleteBuilder().
		WithFilterSpecification(builders.NewWorkoutsExerciseFilterSpecification(&f))

	query, args, err := deleteBuilder.ToSql()
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
		return fmt.Errorf("no rows deleted: %w", errs.ErrWorkoutsExerciseAlreadyDeleted)
	}

	return nil
}
