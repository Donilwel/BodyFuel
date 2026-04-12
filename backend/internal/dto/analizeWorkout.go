package dto

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
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
	TargetWorkoutsPerWeek        int
	LastTimeGenerateWorkout      time.Time
	SkipGeneration               bool
	SkipReason                   string

	// Nutrition context for today
	TodayCalories  int
	TargetCalories int
	CalorieBalance int // consumed - target (positive = surplus, negative = deficit)

	// Weight progress toward goal
	CurrentWeight float64
	TargetWeight  float64
	WeightDelta   float64 // current - target (positive = need to lose, negative = need to gain)
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
