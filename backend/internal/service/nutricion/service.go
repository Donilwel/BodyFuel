package nutricion

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/ai"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type (
	UserFoodRepository interface {
		Create(ctx context.Context, f *entities.UserFood) error
		Get(ctx context.Context, f dto.UserFoodFilter) (*entities.UserFood, error)
		List(ctx context.Context, f dto.UserFoodFilter) ([]*entities.UserFood, error)
		Update(ctx context.Context, f *entities.UserFood) error
		Delete(ctx context.Context, id, userID uuid.UUID) error
	}

	AIClient interface {
		AnalyzeNutritionPhoto(ctx context.Context, imageURL string) (*ai.NutritionAnalysis, error)
	}
)

type NutritionDiary struct {
	Date          time.Time
	Entries       []*entities.UserFood
	TotalCalories int
	TotalProtein  float64
	TotalCarbs    float64
	TotalFat      float64
}

type NutritionReport struct {
	From          time.Time
	To            time.Time
	Entries       []*entities.UserFood
	TotalCalories int
	TotalProtein  float64
	TotalCarbs    float64
	TotalFat      float64
	AvgCalories   float64
	Days          int
}

type Service struct {
	foodRepo UserFoodRepository
	ai       AIClient
}

type Config struct {
	UserFoodRepository UserFoodRepository
	AIClient           AIClient
}

func NewService(c *Config) *Service {
	return &Service{
		foodRepo: c.UserFoodRepository,
		ai:       c.AIClient,
	}
}

// AnalyzePhoto sends the image URL to OpenAI Vision and returns nutritional estimates.
func (s *Service) AnalyzePhoto(ctx context.Context, imageURL string) (*ai.NutritionAnalysis, error) {
	result, err := s.ai.AnalyzeNutritionPhoto(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("analyze photo: %w", err)
	}
	return result, nil
}

// CreateFoodEntry creates a new food diary entry.
func (s *Service) CreateFoodEntry(ctx context.Context, spec entities.UserFoodInitSpec) error {
	entry := entities.NewUserFood(entities.WithUserFoodInitSpec(spec))
	if err := s.foodRepo.Create(ctx, entry); err != nil {
		return fmt.Errorf("create food entry: %w", err)
	}
	return nil
}

// GetFoodEntry retrieves a single food entry by ID for a user.
func (s *Service) GetFoodEntry(ctx context.Context, id, userID uuid.UUID) (*entities.UserFood, error) {
	entry, err := s.foodRepo.Get(ctx, dto.UserFoodFilter{ID: &id, UserID: &userID})
	if err != nil {
		return nil, fmt.Errorf("get food entry: %w", err)
	}
	return entry, nil
}

// UpdateFoodEntry updates a food entry.
func (s *Service) UpdateFoodEntry(ctx context.Context, id, userID uuid.UUID, params entities.UserFoodUpdateParams) error {
	entry, err := s.foodRepo.Get(ctx, dto.UserFoodFilter{ID: &id, UserID: &userID})
	if err != nil {
		return fmt.Errorf("update food entry: not found: %w", err)
	}
	entry.Update(params)
	if err := s.foodRepo.Update(ctx, entry); err != nil {
		return fmt.Errorf("update food entry: %w", err)
	}
	return nil
}

// DeleteFoodEntry removes a food entry.
func (s *Service) DeleteFoodEntry(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.foodRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete food entry: %w", err)
	}
	return nil
}

// GetDiary returns all food entries for a specific date with aggregated totals.
func (s *Service) GetDiary(ctx context.Context, userID uuid.UUID, date time.Time) (*NutritionDiary, error) {
	entries, err := s.foodRepo.List(ctx, dto.UserFoodFilter{UserID: &userID, Date: &date})
	if err != nil {
		return nil, fmt.Errorf("get diary: %w", err)
	}

	diary := &NutritionDiary{
		Date:    date,
		Entries: entries,
	}
	for _, e := range entries {
		diary.TotalCalories += e.Calories()
		diary.TotalProtein += e.Protein()
		diary.TotalCarbs += e.Carbs()
		diary.TotalFat += e.Fat()
	}
	return diary, nil
}

// GetReport returns all food entries for a date range with aggregated totals and daily averages.
func (s *Service) GetReport(ctx context.Context, userID uuid.UUID, from, to time.Time) (*NutritionReport, error) {
	entries, err := s.foodRepo.List(ctx, dto.UserFoodFilter{UserID: &userID, StartDate: &from, EndDate: &to})
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}

	report := &NutritionReport{
		From:    from,
		To:      to,
		Entries: entries,
	}
	for _, e := range entries {
		report.TotalCalories += e.Calories()
		report.TotalProtein += e.Protein()
		report.TotalCarbs += e.Carbs()
		report.TotalFat += e.Fat()
	}

	days := int(to.Sub(from).Hours()/24) + 1
	if days < 1 {
		days = 1
	}
	report.Days = days
	if days > 0 {
		report.AvgCalories = float64(report.TotalCalories) / float64(days)
	}

	return report, nil
}
