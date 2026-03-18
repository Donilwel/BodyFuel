package models

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type WorkoutsExerciseRow struct {
	WorkoutID       uuid.UUID               `db:"workout_id"`
	ExerciseID      uuid.UUID               `db:"exercise_id"`
	ModifyReps      int                     `db:"modify_reps"`
	ModifyRelaxTime int                     `db:"modify_relax_time"`
	Calories        int                     `db:"calories"`
	Status          entities.ExerciseStatus `db:"status"`
	UpdatedAt       time.Time               `db:"updated_at"`
	CreatedAt       time.Time               `db:"created_at"`
}

func NewWorkoutsExerciseRow(workoutsExercise *entities.WorkoutsExercise) *WorkoutsExerciseRow {
	return &WorkoutsExerciseRow{
		WorkoutID:       workoutsExercise.WorkoutID(),
		ExerciseID:      workoutsExercise.ExerciseID(),
		ModifyReps:      workoutsExercise.ModifyReps(),
		ModifyRelaxTime: workoutsExercise.ModifyRelaxTime(),
		Calories:        workoutsExercise.Calories(),
		Status:          workoutsExercise.Status(),
		UpdatedAt:       workoutsExercise.UpdatedAt(),
		CreatedAt:       workoutsExercise.CreatedAt(),
	}
}

func (w *WorkoutsExerciseRow) ToEntity() *entities.WorkoutsExercise {
	return entities.NewWorkoutsExercise(
		entities.WithWorkoutsExerciseRestoreSpec(entities.WorkoutsExerciseRestoreSpec{
			WorkoutID:       w.WorkoutID,
			ExerciseID:      w.ExerciseID,
			ModifyReps:      w.ModifyReps,
			ModifyRelaxTime: w.ModifyRelaxTime,
			Calories:        w.Calories,
			Status:          w.Status,
			UpdatedAt:       w.UpdatedAt,
			CreatedAt:       w.CreatedAt,
		}),
	)
}
