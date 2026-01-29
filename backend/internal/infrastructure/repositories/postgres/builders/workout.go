package builders

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"time"
)

const (
	workoutTable = "bodyfuel.workout"
)

type WorkoutFilterSpecification struct {
	ID                 *uuid.UUID
	UserID             *uuid.UUID
	Level              *entities.WorkoutsLevel
	TotalCalories      *int
	PredictionCalories *int
	Status             *entities.WorkoutsStatus
	Duration           *time.Duration
	CreatedAt          *time.Time
	UpdatedAt          *time.Time
}

func NewWorkoutFilterSpecification(f dto.WorkoutsFilter) *WorkoutFilterSpecification {
	return &WorkoutFilterSpecification{
		ID:                 f.ID,
		UserID:             f.UserID,
		Level:              f.Level,
		TotalCalories:      f.TotalCalories,
		PredictionCalories: f.PredictionCalories,
		Status:             f.Status,
		Duration:           f.Duration,
		CreatedAt:          f.CreatedAt,
		UpdatedAt:          f.UpdatedAt,
	}
}

func (spec *WorkoutFilterSpecification) Predicates() []sq.Sqlizer {
	var predicates []sq.Sqlizer

	if v := spec.ID; v != nil {
		predicates = append(predicates, sq.Eq{"workout.id": v})
	}

	if v := spec.UserID; v != nil {
		predicates = append(predicates, sq.Eq{"workout.user_id": v})
	}

	if v := spec.Level; v != nil {
		predicates = append(predicates, sq.Eq{"workout.level": v})
	}

	if v := spec.TotalCalories; v != nil {
		predicates = append(predicates, sq.Eq{"workout.total_calories": v})
	}

	if v := spec.PredictionCalories; v != nil {
		predicates = append(predicates, sq.Eq{"workout.prediction_calories": v})
	}

	if v := spec.Status; v != nil {
		predicates = append(predicates, sq.Eq{"workout.status": v})
	}

	if v := spec.Duration; v != nil {
		predicates = append(predicates, sq.Eq{"workout.duration": v})
	}

	if v := spec.CreatedAt; v != nil {
		predicates = append(predicates, sq.Eq{"workout.created_at": v})
	}

	if v := spec.UpdatedAt; v != nil {
		predicates = append(predicates, sq.Eq{"workout.updated_at": v})
	}

	return predicates
}

type WorkoutSelectBuilder struct {
	b sq.SelectBuilder
}

func NewWorkoutSelectBuilder() *WorkoutSelectBuilder {
	selectBuilder := newQueryBuilder().Select(
		"workout.id",
		"workout.user_id",
		"workout.level",
		"workout.status",
		"workout.total_calories",
		"workout.prediction_calories",
		"workout.duration",
		"workout.created_at",
		"workout.updated_at",
	).From(workoutTable)

	return &WorkoutSelectBuilder{b: selectBuilder}
}

func (a *WorkoutSelectBuilder) WithFilterSpecification(spec *WorkoutFilterSpecification) *WorkoutSelectBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *WorkoutSelectBuilder) Limit(limit int) *WorkoutSelectBuilder {
	if limit > 0 {
		a.b = a.b.Limit(uint64(limit))
	}

	return a
}

func (a *WorkoutSelectBuilder) Offset(offset int) *WorkoutSelectBuilder {
	a.b = a.b.Offset(uint64(offset))

	return a
}

func (a *WorkoutSelectBuilder) WithBlock() *WorkoutSelectBuilder {
	a.b = a.b.Suffix("FOR UPDATE OF workout")

	return a
}

func (a *WorkoutSelectBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}

type WorkoutDeleteBuilder struct {
	b sq.DeleteBuilder
}

func NewWorkoutDeleteBuilder() *WorkoutDeleteBuilder {
	deleteBuilder := newDeleteQueryBuilder().
		Delete(workoutTable)

	return &WorkoutDeleteBuilder{b: deleteBuilder}
}

func (a *WorkoutDeleteBuilder) WithFilterSpecification(spec *WorkoutFilterSpecification) *WorkoutDeleteBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *WorkoutDeleteBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}
