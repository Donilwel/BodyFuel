package builders

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const (
	exerciseTable = "bodyfuel.exercise"
)

type ExerciseFilterSpecification struct {
	ID               *uuid.UUID
	LevelPreparation *entities.LevelPreparation
	Name             *string
	TypeExercise     *entities.ExerciseType
	Description      *string
	BaseCountReps    *int
	Steps            *int
	LinkGif          *string
	PlaceExercise    *entities.PlaceExercise
	AvgCaloriesPer   *float64
	BaseRelaxTime    *int
}

func NewExerciseFilterSpecification(f dto.ExerciseFilter) *ExerciseFilterSpecification {
	return &ExerciseFilterSpecification{
		ID:               f.ID,
		LevelPreparation: f.LevelPreparation,
		Name:             f.Name,
		TypeExercise:     f.TypeExercise,
		Description:      f.Description,
		BaseCountReps:    f.BaseCountReps,
		Steps:            f.Steps,
		LinkGif:          f.LinkGif,
		PlaceExercise:    f.PlaceExercise,
		AvgCaloriesPer:   f.AvgCaloriesPer,
		BaseRelaxTime:    f.BaseRelaxTime,
	}
}

type ExerciseRow struct {
	ID               uuid.UUID                 `db:"id"`
	LevelPreparation entities.LevelPreparation `db:"level_preparation"`
	Name             string                    `db:"name"`
	TypeExercise     entities.ExerciseType     `db:"type_exercise"`
	Description      string                    `db:"description"`
	BaseCountReps    int                       `db:"base_count_reps"`
	Steps            int                       `db:"steps"`
	LinkGif          string                    `db:"link_gif"`
	PlaceExercise    entities.PlaceExercise    `db:"place_exercise"`
	AvgCaloriesPer   float64                   `db:"avg_calories_per"`
	BaseRelaxTime    int                       `db:"base_relax_time"`
}

func (spec *ExerciseFilterSpecification) Predicates() []sq.Sqlizer {
	var predicates []sq.Sqlizer

	if v := spec.ID; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.id": v})
	}

	if v := spec.LevelPreparation; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.level_preparation": v})
	}

	if v := spec.Name; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.name": v})
	}

	if v := spec.TypeExercise; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.type_exercise": v})
	}

	if v := spec.Description; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.description": v})
	}

	if v := spec.BaseCountReps; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.base_count_reps": v})
	}

	if v := spec.Steps; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.steps": v})
	}

	if v := spec.LinkGif; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.link_gif": v})
	}

	if v := spec.PlaceExercise; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.place_exercise": v})
	}

	if v := spec.AvgCaloriesPer; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.avg_calories_per": v})
	}

	if v := spec.BaseRelaxTime; v != nil {
		predicates = append(predicates, sq.Eq{"exercise.base_relax_time": v})
	}

	return predicates
}

type ExerciseSelectBuilder struct {
	b sq.SelectBuilder
}

func NewExerciseSelectBuilder() *ExerciseSelectBuilder {
	selectBuilder := newQueryBuilder().Select(
		"exercise.id",
		"exercise.level_preparation",
		"exercise.name",
		"exercise.type_exercise",
		"exercise.description",
		"exercise.base_count_reps",
		"exercise.steps",
		"exercise.link_gif",
		"exercise.place_exercise",
		"exercise.avg_calories_per",
		"exercise.base_relax_time",
	).From(exerciseTable)

	return &ExerciseSelectBuilder{b: selectBuilder}
}

func (a *ExerciseSelectBuilder) WithFilterSpecification(spec *ExerciseFilterSpecification) *ExerciseSelectBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *ExerciseSelectBuilder) Limit(limit int) *ExerciseSelectBuilder {
	if limit > 0 {
		a.b = a.b.Limit(uint64(limit))
	}

	return a
}

func (a *ExerciseSelectBuilder) Offset(offset int) *ExerciseSelectBuilder {
	a.b = a.b.Offset(uint64(offset))

	return a
}

func (a *ExerciseSelectBuilder) WithBlock() *ExerciseSelectBuilder {
	a.b = a.b.Suffix("FOR UPDATE OF exercise")

	return a
}

func (a *ExerciseSelectBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}

type ExerciseDeleteBuilder struct {
	b sq.DeleteBuilder
}

func NewExerciseDeleteBuilder() *ExerciseDeleteBuilder {
	deleteBuilder := newDeleteQueryBuilder().
		Delete(exerciseTable)

	return &ExerciseDeleteBuilder{b: deleteBuilder}
}

func (a *ExerciseDeleteBuilder) WithFilterSpecification(spec *ExerciseFilterSpecification) *ExerciseDeleteBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *ExerciseDeleteBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}
