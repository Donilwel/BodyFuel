package dto

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type AnalyzeWorkoutStats struct {
	IDUser                       uuid.UUID
	TotalWorkouts                int
	TotalFinished                int
	TotalCancelled               int
	TotalNew                     int
	TotalFinishedWorkoutsForWeek int
	AWGLevel                     entities.WorkoutsLevel
	PopularExerciseType          entities.ExerciseType
	PopularPlaceExercise         entities.PlaceExercise
	LastTimeGenerateWorkout      time.Time
	SkipGeneration               bool
	SkipReason                   string
}

//	uuid_user
//	[]exercise
//	level
//	time.Duration()
//	total_calories
//	predictiction_calories
//	user_condition->user_info
//	status="new, in_progress, canceled, finished"
//	Date
