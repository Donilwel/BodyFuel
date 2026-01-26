package recomendation

import "time"

type RecommendationEngine struct {
	userProfiles map[string]UserProfile
}

type UserProfile struct {
	FitnessLevel string          `json:"fitness_level"` // beginner, intermediate, advanced
	Goals        []Goal          `json:"goals"`         // weight_loss, muscle_gain, endurance
	Preferences  Preferences     `json:"preferences"`
	Restrictions []string        `json:"restrictions"` // dietary restrictions
	Progress     ProgressHistory `json:"progress"`
}

func (e *RecommendationEngine) GenerateWorkoutPlan(userID string, data UserData) WorkoutPlan {
	// Анализ данных пользователя
	calorieDeficit := e.calculateCalorieDeficit(userID)
	fitnessLevel := e.assessFitnessLevel(userID)
	recoveryStatus := e.assessRecovery(userID)

	// Генерация плана
	plan := WorkoutPlan{
		UserID:    userID,
		Date:      time.Now(),
		Exercises: []Exercise{},
	}

	if recoveryStatus == "needs_rest" {
		plan.RestDay = true
		plan.Recommendation = "День отдыха. Ваш организм нуждается в восстановлении."
	} else if calorieDeficit > 500 {
		// Большой дефицит - легкие кардио
		plan.Exercises = e.generateLightCardioPlan()
	} else {
		// Нормальный режим - силовая + кардио
		plan.Exercises = e.generateBalancedPlan(fitnessLevel)
	}

	return plan
}
