package nutricion

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/ai"
	"backend/pkg/cache"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		AnalyzeNutritionPhotoData(ctx context.Context, data []byte, mediaType string) (*ai.NutritionAnalysis, error)
		GenerateRecipes(ctx context.Context, intake ai.DailyIntake) ([]ai.RecipeItem, error)
	}

	// StorageService handles uploading food photos to object storage.
	StorageService interface {
		UploadFoodPhoto(ctx context.Context, userID, objectName, contentType string, data io.Reader) (string, error)
	}

	// RecipeCache stores/retrieves AI-generated recipes to avoid repeated OpenAI calls.
	RecipeCache interface {
		Get(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key, value string, ttl time.Duration) error
		Del(ctx context.Context, keys ...string) error
	}
)

const recipeCacheTTL = 2 * time.Hour

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
	storage  StorageService
	cache    RecipeCache // optional, nil means no caching
}

type Config struct {
	UserFoodRepository UserFoodRepository
	AIClient           AIClient
	StorageService     StorageService
	RecipeCache        RecipeCache // optional
}

func NewService(c *Config) *Service {
	return &Service{
		foodRepo: c.UserFoodRepository,
		ai:       c.AIClient,
		storage:  c.StorageService,
		cache:    c.RecipeCache,
	}
}

// recipeCacheKey returns the Redis key for a user's recipe suggestions on a given date.
// Key format: recipes:{userID}:{YYYY-MM-DD}
func recipeCacheKey(userID uuid.UUID, date time.Time) string {
	return fmt.Sprintf("recipes:%s:%s", userID, date.Format("2006-01-02"))
}

// AnalyzePhoto sends the image URL to OpenAI Vision and returns nutritional estimates.
func (s *Service) AnalyzePhoto(ctx context.Context, imageURL string) (*ai.NutritionAnalysis, error) {
	result, err := s.ai.AnalyzeNutritionPhoto(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("analyze photo: %w", err)
	}
	return result, nil
}

// UploadPhotoResult is returned by UploadAndAnalyzePhoto.
type UploadPhotoResult struct {
	Analysis *ai.NutritionAnalysis
	PhotoURL string
}

// UploadAndAnalyzePhoto uploads the food photo to S3 and analyzes it with OpenAI Vision.
// Bytes are read once into memory so they can be used for both upload and base64 analysis
// (MinIO public URL is localhost-only and not reachable by OpenAI).
func (s *Service) UploadAndAnalyzePhoto(ctx context.Context, userID, filename, contentType string, data io.Reader) (*UploadPhotoResult, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("storage service not configured")
	}

	imgBytes, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("read photo data: %w", err)
	}

	photoURL, err := s.storage.UploadFoodPhoto(ctx, userID, filename, contentType, bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("upload food photo: %w", err)
	}

	analysis, err := s.ai.AnalyzeNutritionPhotoData(ctx, imgBytes, contentType)
	if err != nil {
		return nil, fmt.Errorf("analyze uploaded photo: %w", err)
	}

	return &UploadPhotoResult{
		Analysis: analysis,
		PhotoURL: photoURL,
	}, nil
}

// CreateFoodEntry creates a new food diary entry and invalidates the recipe cache for that day.
func (s *Service) CreateFoodEntry(ctx context.Context, spec entities.UserFoodInitSpec) error {
	entry := entities.NewUserFood(entities.WithUserFoodInitSpec(spec))
	if err := s.foodRepo.Create(ctx, entry); err != nil {
		return fmt.Errorf("create food entry: %w", err)
	}
	s.invalidateRecipeCache(ctx, spec.UserID, spec.Date)
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

// UpdateFoodEntry updates a food entry and invalidates the recipe cache for that day.
func (s *Service) UpdateFoodEntry(ctx context.Context, id, userID uuid.UUID, params entities.UserFoodUpdateParams) error {
	entry, err := s.foodRepo.Get(ctx, dto.UserFoodFilter{ID: &id, UserID: &userID})
	if err != nil {
		return fmt.Errorf("update food entry: not found: %w", err)
	}
	entry.Update(params)
	if err := s.foodRepo.Update(ctx, entry); err != nil {
		return fmt.Errorf("update food entry: %w", err)
	}
	s.invalidateRecipeCache(ctx, userID, entry.Date())
	return nil
}

// DeleteFoodEntry removes a food entry and invalidates the recipe cache for that day.
func (s *Service) DeleteFoodEntry(ctx context.Context, id, userID uuid.UUID) error {
	// Fetch before delete to know the date for cache invalidation.
	entry, err := s.foodRepo.Get(ctx, dto.UserFoodFilter{ID: &id, UserID: &userID})
	if err == nil {
		defer s.invalidateRecipeCache(ctx, userID, entry.Date())
	}
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

// RecommendRecipes returns AI-generated recipe suggestions based on what the user has eaten today.
// Results are cached in Redis for 2 hours. The cache is invalidated automatically whenever
// the user creates, updates, or deletes a food entry for the same date.
func (s *Service) RecommendRecipes(ctx context.Context, userID uuid.UUID, date time.Time) ([]ai.RecipeItem, error) {
	diary, err := s.GetDiary(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("recommend recipes: %w", err)
	}

	intake := ai.DailyIntake{
		ConsumedCalories: diary.TotalCalories,
		ConsumedProtein:  diary.TotalProtein,
		ConsumedCarbs:    diary.TotalCarbs,
		ConsumedFat:      diary.TotalFat,
	}

	if s.cache != nil {
		key := recipeCacheKey(userID, date)

		if cached, err := s.cache.Get(ctx, key); err == nil {
			var items []ai.RecipeItem
			if json.Unmarshal([]byte(cached), &items) == nil {
				return items, nil
			}
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			// Redis is down — log and fall through to OpenAI.
			_ = err
		}
	}

	recipes, err := s.ai.GenerateRecipes(ctx, intake)
	if err != nil {
		return nil, fmt.Errorf("recommend recipes: %w", err)
	}

	if s.cache != nil {
		if data, jerr := json.Marshal(recipes); jerr == nil {
			_ = s.cache.Set(ctx, recipeCacheKey(userID, date), string(data), recipeCacheTTL)
		}
	}

	return recipes, nil
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

// invalidateRecipeCache deletes the recipe cache for a given user+date.
// Errors are swallowed — a stale cache is better than a failed request.
func (s *Service) invalidateRecipeCache(ctx context.Context, userID uuid.UUID, date time.Time) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Del(ctx, recipeCacheKey(userID, date))
}
