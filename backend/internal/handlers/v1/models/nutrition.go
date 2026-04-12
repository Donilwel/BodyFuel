package models

import (
	"backend/internal/domain/entities"
	"backend/internal/service/nutricion"
	"backend/pkg/ai"
	"time"

	"github.com/google/uuid"
)

// --- Requests ---

type AnalyzePhotoRequest struct {
	ImageURL string `json:"image_url" validate:"required,url"`
}

type CreateFoodEntryRequest struct {
	Description string    `json:"description" validate:"required,max=500"`
	Calories    int       `json:"calories" validate:"required,min=0,max=10000"`
	Protein     float64   `json:"protein" validate:"omitempty,min=0,max=500"`
	Carbs       float64   `json:"carbs" validate:"omitempty,min=0,max=500"`
	Fat         float64   `json:"fat" validate:"omitempty,min=0,max=500"`
	MealType    string    `json:"meal_type" validate:"required,oneof=breakfast lunch dinner snack"`
	PhotoURL    string    `json:"photo_url" validate:"omitempty,url"`
	Date        time.Time `json:"date"`
}

type UpdateFoodEntryRequest struct {
	Description *string    `json:"description" validate:"omitempty,max=500"`
	Calories    *int       `json:"calories" validate:"omitempty,min=0,max=10000"`
	Protein     *float64   `json:"protein" validate:"omitempty,min=0,max=500"`
	Carbs       *float64   `json:"carbs" validate:"omitempty,min=0,max=500"`
	Fat         *float64   `json:"fat" validate:"omitempty,min=0,max=500"`
	MealType    *string    `json:"meal_type" validate:"omitempty,oneof=breakfast lunch dinner snack"`
	PhotoURL    *string    `json:"photo_url" validate:"omitempty,url"`
	Date        *time.Time `json:"date"`
}

// --- Responses ---

type NutritionAnalysisResponse struct {
	Description string  `json:"description"`
	Calories    int     `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
}

func NewNutritionAnalysisResponse(a *ai.NutritionAnalysis) NutritionAnalysisResponse {
	return NutritionAnalysisResponse{
		Description: a.Description,
		Calories:    a.Calories,
		Protein:     a.Protein,
		Carbs:       a.Carbs,
		Fat:         a.Fat,
	}
}

// UploadPhotoAnalysisResponse is returned by POST /nutrition/analyze/upload.
type UploadPhotoAnalysisResponse struct {
	PhotoURL    string  `json:"photo_url"`
	Description string  `json:"description"`
	Calories    int     `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
}

type FoodEntryResponse struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Calories    int       `json:"calories"`
	Protein     float64   `json:"protein"`
	Carbs       float64   `json:"carbs"`
	Fat         float64   `json:"fat"`
	MealType    string    `json:"meal_type"`
	PhotoURL    string    `json:"photo_url,omitempty"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewFoodEntryResponse(f *entities.UserFood) FoodEntryResponse {
	return FoodEntryResponse{
		ID:          f.ID(),
		Description: f.Description(),
		Calories:    f.Calories(),
		Protein:     f.Protein(),
		Carbs:       f.Carbs(),
		Fat:         f.Fat(),
		MealType:    string(f.MealType()),
		PhotoURL:    f.PhotoURL(),
		Date:        f.Date(),
		CreatedAt:   f.CreatedAt(),
	}
}

func NewFoodEntryResponseList(list []*entities.UserFood) []FoodEntryResponse {
	result := make([]FoodEntryResponse, len(list))
	for i, f := range list {
		result[i] = NewFoodEntryResponse(f)
	}
	return result
}

type NutritionDiaryResponse struct {
	Date          string              `json:"date"`
	Entries       []FoodEntryResponse `json:"entries"`
	TotalCalories int                 `json:"total_calories"`
	TotalProtein  float64             `json:"total_protein"`
	TotalCarbs    float64             `json:"total_carbs"`
	TotalFat      float64             `json:"total_fat"`
}

func NewNutritionDiaryResponse(d *nutricion.NutritionDiary) NutritionDiaryResponse {
	return NutritionDiaryResponse{
		Date:          d.Date.Format("2006-01-02"),
		Entries:       NewFoodEntryResponseList(d.Entries),
		TotalCalories: d.TotalCalories,
		TotalProtein:  d.TotalProtein,
		TotalCarbs:    d.TotalCarbs,
		TotalFat:      d.TotalFat,
	}
}

type NutritionReportResponse struct {
	From          string              `json:"from"`
	To            string              `json:"to"`
	Days          int                 `json:"days"`
	Entries       []FoodEntryResponse `json:"entries"`
	TotalCalories int                 `json:"total_calories"`
	TotalProtein  float64             `json:"total_protein"`
	TotalCarbs    float64             `json:"total_carbs"`
	TotalFat      float64             `json:"total_fat"`
	AvgCalories   float64             `json:"avg_calories_per_day"`
}

type MacroNutrientsResponse struct {
	Protein float64 `json:"protein"`
	Fat     float64 `json:"fat"`
	Carbs   float64 `json:"carbs"`
}

type RecipeResponse struct {
	ID              uuid.UUID             `json:"id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Macros          MacroNutrientsResponse `json:"macros"`
	PreparationTime int                   `json:"preparation_time"`
}

func NewRecipeResponseList(items []ai.RecipeItem) []RecipeResponse {
	result := make([]RecipeResponse, len(items))
	for i, item := range items {
		result[i] = RecipeResponse{
			ID:   uuid.New(),
			Name: item.Name,
			Description: item.Description,
			Macros: MacroNutrientsResponse{
				Protein: item.Macros.Protein,
				Fat:     item.Macros.Fat,
				Carbs:   item.Macros.Carbs,
			},
			PreparationTime: item.PreparationTime,
		}
	}
	return result
}

func NewNutritionReportResponse(r *nutricion.NutritionReport) NutritionReportResponse {
	return NutritionReportResponse{
		From:          r.From.Format("2006-01-02"),
		To:            r.To.Format("2006-01-02"),
		Days:          r.Days,
		Entries:       NewFoodEntryResponseList(r.Entries),
		TotalCalories: r.TotalCalories,
		TotalProtein:  r.TotalProtein,
		TotalCarbs:    r.TotalCarbs,
		TotalFat:      r.TotalFat,
		AvgCalories:   r.AvgCalories,
	}
}
