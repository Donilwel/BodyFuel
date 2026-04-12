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
	"strings"
	"time"

	"github.com/google/uuid"
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
		"updated_at",
		"created_at"
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

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
		row.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}
	return nil
}

func (r *WorkoutsExerciseRepo) CreateBulk(ctx context.Context, workoutExercises []entities.WorkoutsExercise) error {
	if len(workoutExercises) == 0 {
		return nil
	}

	fmt.Printf("Creating %d workout exercises\n", len(workoutExercises))

	workoutID := workoutExercises[0].WorkoutID()
	for _, we := range workoutExercises {
		if we.WorkoutID() != workoutID {
			return fmt.Errorf("all exercises must belong to the same workout")
		}
	}

	const numFields = 8

	valueStrings := make([]string, 0, len(workoutExercises))
	valueArgs := make([]interface{}, 0, len(workoutExercises)*numFields)

	for i, we := range workoutExercises {
		row := models.NewWorkoutsExerciseRow(&we)

		placeholders := make([]string, numFields)
		for j := 0; j < numFields; j++ {
			placeholders[j] = fmt.Sprintf("$%d", i*numFields+j+1)
		}
		valueStrings = append(valueStrings, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))

		valueArgs = append(valueArgs,
			row.WorkoutID,
			row.ExerciseID,
			row.ModifyReps,
			row.ModifyRelaxTime,
			row.Calories,
			row.Status,
			row.UpdatedAt,
			row.CreatedAt,
		)
	}

	query := fmt.Sprintf(`INSERT INTO bodyfuel.workouts_exercise (
		"workout_id",
		"exercise_id",
		"modify_reps",
		"modify_relax_time",
		"calories",
		"status",
		"updated_at",
		"created_at"
	) VALUES %s`, strings.Join(valueStrings, ","))

	_, err := r.getter.Get(ctx).ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("bulk insert exec context: %w, query: %s", err, query)
	}

	return nil
}

// CreateBulkSimplified - упрощенная версия с использованием NamedExec
func (r *WorkoutsExerciseRepo) CreateBulkSimplified(ctx context.Context, workoutExercises []entities.WorkoutsExercise) error {
	if len(workoutExercises) == 0 {
		return nil
	}

	// Конвертируем в слайс интерфейсов для NamedExec
	rows := make([]interface{}, len(workoutExercises))
	for i, we := range workoutExercises {
		rows[i] = models.NewWorkoutsExerciseRow(&we)
	}

	// Используем NamedExec для массовой вставки
	_, err := r.getter.Get(ctx).NamedExecContext(ctx, `
		INSERT INTO bodyfuel.workouts_exercise (
			workout_id,
			exercise_id,
			modify_reps,
			modify_relax_time,
			calories,
			status,
			updated_at,
			created_at
		) VALUES (
			:workout_id,
			:exercise_id,
			:modify_reps,
			:modify_relax_time,
			:calories,
			:status,
			:updated_at,
			:created_at
		)
	`, rows)

	if err != nil {
		return fmt.Errorf("bulk insert named exec: %w", err)
	}

	return nil
}

// CreateBulkWithReturning создает несколько упражнений и возвращает созданные записи
func (r *WorkoutsExerciseRepo) CreateBulkWithReturning(ctx context.Context, workoutExercises []entities.WorkoutsExercise) ([]*entities.WorkoutsExercise, error) {
	if len(workoutExercises) == 0 {
		return []*entities.WorkoutsExercise{}, nil
	}

	// Проверяем, что все упражнения принадлежат одной тренировке
	workoutID := workoutExercises[0].WorkoutID()
	for _, we := range workoutExercises {
		if we.WorkoutID() != workoutID {
			return nil, fmt.Errorf("all exercises must belong to the same workout")
		}
	}

	const numFields = 8

	// Строим массовый INSERT запрос с RETURNING
	valueStrings := make([]string, 0, len(workoutExercises))
	valueArgs := make([]interface{}, 0, len(workoutExercises)*numFields)

	for i, we := range workoutExercises {
		row := models.NewWorkoutsExerciseRow(&we)

		placeholders := make([]string, numFields)
		for j := 0; j < numFields; j++ {
			placeholders[j] = fmt.Sprintf("$%d", i*numFields+j+1)
		}
		valueStrings = append(valueStrings, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))

		valueArgs = append(valueArgs,
			row.WorkoutID,
			row.ExerciseID,
			row.ModifyReps,
			row.ModifyRelaxTime,
			row.Calories,
			row.Status,
			row.UpdatedAt,
			row.CreatedAt,
		)
	}

	query := fmt.Sprintf(`INSERT INTO bodyfuel.workouts_exercise (
		"workout_id",
		"exercise_id",
		"modify_reps",
		"modify_relax_time",
		"calories",
		"status",
		"updated_at",
		"created_at"
	) VALUES %s 
	RETURNING 
		workout_id,
		exercise_id,
		modify_reps,
		modify_relax_time,
		calories,
		status,
		updated_at,
		created_at`, strings.Join(valueStrings, ","))

	// Выполняем запрос и сканируем результаты
	var rows []*models.WorkoutsExerciseRow
	err := r.getter.Get(ctx).SelectContext(ctx, &rows, query, valueArgs...)
	if err != nil {
		return nil, fmt.Errorf("bulk insert with returning: %w", err)
	}

	// Конвертируем в entity
	result := make([]*entities.WorkoutsExercise, len(rows))
	for i := range rows {
		result[i] = rows[i].ToEntity()
	}

	return result, nil
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

// ListSkippedExercises returns skip aggregates (per exercise_id) for a given user since the provided time.
// Useful for skip-tracking logic in workout generation.
func (r *WorkoutsExerciseRepo) ListSkippedExercises(ctx context.Context, userID uuid.UUID, since time.Time) ([]dto.SkippedExerciseInfo, error) {
	const query = `
		SELECT we.exercise_id, COUNT(*) AS skip_count, MAX(we.updated_at) AS last_skipped_at
		FROM bodyfuel.workouts_exercise we
		JOIN bodyfuel.workouts w ON w.id = we.workout_id
		WHERE w.user_id = $1
		  AND we.status = 'skipped'
		  AND we.updated_at > $2
		GROUP BY we.exercise_id`

	type row struct {
		ExerciseID    uuid.UUID `db:"exercise_id"`
		SkipCount     int       `db:"skip_count"`
		LastSkippedAt time.Time `db:"last_skipped_at"`
	}

	var rows []row
	if err := r.getter.Get(ctx).SelectContext(ctx, &rows, query, userID, since); err != nil {
		return nil, fmt.Errorf("list skipped exercises: %w", err)
	}

	result := make([]dto.SkippedExerciseInfo, len(rows))
	for i, row := range rows {
		result[i] = dto.SkippedExerciseInfo{
			ExerciseID:    row.ExerciseID,
			SkipCount:     row.SkipCount,
			LastSkippedAt: row.LastSkippedAt,
		}
	}
	return result, nil
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
