package builders

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const (
	workoutsExerciseTable = "bodyfuel.workouts_exercise"
)

type WorkoutsExerciseFilterSpecification struct {
	WorkoutID       *uuid.UUID
	ExerciseID      *uuid.UUID
	ModifyReps      *int
	ModifyRelaxTime *int
	Calories        *int
	Status          *entities.ExerciseStatus
	UpdateAt        *time.Time
}

func NewWorkoutsExerciseFilterSpecification(f *dto.WorkoutsExerciseFilter) *WorkoutsExerciseFilterSpecification {
	return &WorkoutsExerciseFilterSpecification{
		WorkoutID:       f.WorkoutID,
		ExerciseID:      f.ExerciseID,
		ModifyReps:      f.ModifyReps,
		ModifyRelaxTime: f.ModifyRelaxTime,
		Calories:        f.Calories,
		Status:          f.Status,
		UpdateAt:        f.UpdateAt,
	}
}

func (spec *WorkoutsExerciseFilterSpecification) Predicates() []sq.Sqlizer {
	var predicates []sq.Sqlizer

	if v := spec.WorkoutID; v != nil {
		predicates = append(predicates, sq.Eq{"workouts_exercise.workout_id": v})
	}

	if v := spec.ExerciseID; v != nil {
		predicates = append(predicates, sq.Eq{"workouts_exercise.exercise_id": v})
	}

	if v := spec.ModifyReps; v != nil {
		predicates = append(predicates, sq.Eq{"workouts_exercise.modify_reps": v})
	}
	if v := spec.ModifyRelaxTime; v != nil {
		predicates = append(predicates, sq.Eq{"workouts_exercise.modify_relax_time": v})
	}
	if v := spec.Calories; v != nil {
		predicates = append(predicates, sq.Eq{"workouts_exercise.calories": v})
	}
	if v := spec.Status; v != nil {
		predicates = append(predicates, sq.Eq{"workouts_exercise.status": v})
	}
	if v := spec.UpdateAt; v != nil {
		predicates = append(predicates, sq.Eq{"workouts_exercise.updated_at": v})
	}

	return predicates
}

type WorkoutsExerciseSelectBuilder struct {
	b sq.SelectBuilder
}

func NewWorkoutsExerciseSelectBuilder() *WorkoutsExerciseSelectBuilder {
	selectBuilder := newQueryBuilder().Select(
		"workouts_exercise.workout_id",
		"workouts_exercise.user_id",
		"workouts_exercise.exercise_id",
		"workouts_exercise.modify_reps",
		"workouts_exercise.modify_relax_time",
		"workouts_exercise.calories",
		"workouts_exercise.status",
		"workouts_exercise.updated_at").From(workoutsExerciseTable)
	return &WorkoutsExerciseSelectBuilder{b: selectBuilder}
}

func (a *WorkoutsExerciseSelectBuilder) WithFilterSpecification(spec *WorkoutsExerciseFilterSpecification) *WorkoutsExerciseSelectBuilder {
	a.b = ApplyFilter(a.b, spec)
	return a
}

func (a *WorkoutsExerciseSelectBuilder) Offset(offset int) *WorkoutsExerciseSelectBuilder {
	a.b = a.b.Offset(uint64(offset))
	return a
}

func (a *WorkoutsExerciseSelectBuilder) WithBlock() *WorkoutsExerciseSelectBuilder {
	a.b = a.b.Suffix("FOR UPDATE OF workouts_exercise")
	return a
}

func (a *WorkoutsExerciseSelectBuilder) ToSql() (string, []any, error) {
	return a.b.ToSql()
}

type WorkoutsExerciseDeleteBuilder struct {
	b sq.DeleteBuilder
}

func NewWorkoutsExerciseDeleteBuilder() *WorkoutsExerciseDeleteBuilder {
	deleteBuilder := newDeleteQueryBuilder().
		Delete(workoutsExerciseTable)

	return &WorkoutsExerciseDeleteBuilder{b: deleteBuilder}
}

func (a *WorkoutsExerciseDeleteBuilder) WithFilterSpecification(spec *WorkoutsExerciseFilterSpecification) *WorkoutsExerciseDeleteBuilder {
	a.b = ApplyFilter(a.b, spec)
	return a
}

func (a *WorkoutsExerciseDeleteBuilder) ToSql() (string, []any, error) {
	return a.b.ToSql()
}
