package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
)

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

func NewExerciseRow(exercise *entities.Exercise) *ExerciseRow {
	return &ExerciseRow{
		ID:               exercise.ID(),
		Name:             exercise.Name(),
		BaseCountReps:    exercise.BaseCountReps(),
		PlaceExercise:    exercise.PlaceExercise(),
		TypeExercise:     exercise.TypeExercise(),
		AvgCaloriesPer:   exercise.AvgCaloriesPer(),
		BaseRelaxTime:    exercise.BaseRelaxTime(),
		Description:      exercise.Description(),
		Steps:            exercise.Steps(),
		LinkGif:          exercise.LinkGif(),
		LevelPreparation: exercise.LevelPreparation(),
	}
}

func (u *ExerciseRow) ToEntity() *entities.Exercise {
	return entities.NewExercise(
		entities.WithExerciseRestoreSpec(entities.ExerciseRestoreSpec{
			ID:               u.ID,
			Name:             u.Name,
			BaseCountReps:    u.BaseCountReps,
			PlaceExercise:    u.PlaceExercise,
			TypeExercise:     u.TypeExercise,
			AvgCaloriesPer:   u.AvgCaloriesPer,
			BaseRelaxTime:    u.BaseRelaxTime,
			Description:      u.Description,
			Steps:            u.Steps,
			LinkGif:          u.LinkGif,
			LevelPreparation: u.LevelPreparation,
		}),
	)
}
