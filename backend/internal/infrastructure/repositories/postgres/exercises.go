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
	queryCreateExercise = `INSERT INTO bodyfuel.exercise (
                                    "id",
                                    "level_preparation",
                                    "name",
                                    "type_exercise",
                                    "description",
                                    "base_count_reps",
                                    "steps",
                                    "link_gif",
                               		"place_exercise",
                              		"avg_calories_per",
                              		"base_relax_time") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	queryUpdateExercise = `UPDATE bodyfuel.exercise SET
									level_preparation=:level_preparation,
									name=:name,
									type_exercise=:type_exercise,
									description=:description,
									base_count_reps=:base_count_reps,
									steps=:steps,
									link_gif=:link_gif,
									place_exercise=:place_exercise,
									avg_calories_per=:avg_calories_per,
									base_relax_time=:base_relax_time
									WHERE id=:id`
)

type ExerciseRepo struct {
	getter dbClientGetter
}

func NewExerciseRepository(db *sqlx.DB) *ExerciseRepo {
	return &ExerciseRepo{getter: dbClientGetter{db: db}}
}

func (r *ExerciseRepo) Get(ctx context.Context, f dto.ExerciseFilter, withBlock bool) (*entities.Exercise, error) {
	selectBuilder := builders.NewExerciseSelectBuilder().WithFilterSpecification(builders.NewExerciseFilterSpecification(f))
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}
	query, args, err := selectBuilder.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.ExerciseRow
	if err := r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrExerciseNotFound
		}
		return nil, fmt.Errorf("get context: %w", err)
	}

	return row.ToEntity(), nil
}

func (r *ExerciseRepo) Create(ctx context.Context, exercise *entities.Exercise) error {
	row := models.NewExerciseRow(exercise)

	_, err := r.getter.Get(ctx).ExecContext(ctx, queryCreateExercise,
		row.ID,
		row.LevelPreparation,
		row.Name,
		row.TypeExercise,
		row.Description,
		row.BaseCountReps,
		row.Steps,
		row.LinkGif,
		row.PlaceExercise,
		row.AvgCaloriesPer,
		row.BaseRelaxTime,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}
	return nil
}

func (r *ExerciseRepo) Update(ctx context.Context, exercise *entities.Exercise) error {
	row := models.NewExerciseRow(exercise)

	res, err := r.getter.Get(ctx).NamedExecContext(ctx, queryUpdateExercise, row)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", errs.ErrExerciseNotFound)
	}

	if rowAffected == 0 {
		return fmt.Errorf("rows affected: %w", err)
	}

	return nil
}

func (r *ExerciseRepo) Delete(ctx context.Context, f dto.ExerciseFilter) error {
	deleteBuilder := builders.NewExerciseDeleteBuilder().WithFilterSpecification(builders.NewExerciseFilterSpecification(f))
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
		return fmt.Errorf("no rows deleted: %w", errs.ErrExerciseAlreadyDeleted)
	}

	return nil
}

func (r *ExerciseRepo) List(ctx context.Context, f dto.ExerciseFilter, withBlock bool) ([]*entities.Exercise, error) {
	var rows []*models.ExerciseRow

	selectBuilder := builders.NewExerciseSelectBuilder()
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}

	query, args, err := selectBuilder.WithFilterSpecification(builders.NewExerciseFilterSpecification(f)).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select context: %w", err)
	}
	result := make([]*entities.Exercise, len(rows))
	for i := range rows {
		result[i] = rows[i].ToEntity()
	}

	return result, nil
}
