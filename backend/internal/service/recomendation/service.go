package recomendation

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/ai"
	"backend/pkg/cache"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type (
	UserRecommendationRepository interface {
		Create(ctx context.Context, r *entities.UserRecommendation) error
		Get(ctx context.Context, f dto.UserRecommendationFilter) (*entities.UserRecommendation, error)
		List(ctx context.Context, f dto.UserRecommendationFilter) ([]*entities.UserRecommendation, error)
		MarkRead(ctx context.Context, id, userID uuid.UUID) error
		DeleteByUser(ctx context.Context, userID uuid.UUID) error
	}

	UserParamsRepository interface {
		Get(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error)
	}

	UserWeightRepository interface {
		List(ctx context.Context, f dto.UserWeightFilter, withBlock bool) ([]*entities.UserWeight, error)
	}

	AIClient interface {
		GenerateRecommendations(ctx context.Context, profile ai.UserProfile) ([]ai.RecommendationItem, error)
	}

	// RecommendationCache is used to enforce a cooldown between AI generation calls.
	RecommendationCache interface {
		Get(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key, value string, ttl time.Duration) error
	}
)

// refreshCooldown is the minimum time between two OpenAI calls for the same user.
const refreshCooldown = 6 * time.Hour

type Service struct {
	recRepo        UserRecommendationRepository
	userParamsRepo UserParamsRepository
	userWeightRepo UserWeightRepository
	ai             AIClient
	cache          RecommendationCache // optional, nil means no cooldown
}

type Config struct {
	RecommendationRepository UserRecommendationRepository
	UserParamsRepository     UserParamsRepository
	UserWeightRepository     UserWeightRepository
	AIClient                 AIClient
	RecommendationCache      RecommendationCache // optional
}

func NewService(c *Config) *Service {
	return &Service{
		recRepo:        c.RecommendationRepository,
		userParamsRepo: c.UserParamsRepository,
		userWeightRepo: c.UserWeightRepository,
		ai:             c.AIClient,
		cache:          c.RecommendationCache,
	}
}

// cooldownKey returns the Redis key used to track the last AI call for a user.
func cooldownKey(userID uuid.UUID) string {
	return fmt.Sprintf("rec_cooldown:%s", userID)
}

// List returns paginated recommendations for a user.
func (s *Service) List(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.UserRecommendation, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	recs, err := s.recRepo.List(ctx, dto.UserRecommendationFilter{
		UserID: &userID,
		Limit:  &limit,
		Offset: &offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list recommendations: %w", err)
	}
	return recs, nil
}

// Refresh generates new recommendations via OpenAI, replacing existing ones.
// If Redis is configured, calls are throttled to once per refreshCooldown (6 hours).
// During the cooldown the existing recommendations stored in postgres are returned instead.
func (s *Service) Refresh(ctx context.Context, userID uuid.UUID) ([]*entities.UserRecommendation, error) {
	// Cooldown check: if a recent generation exists, return stored recommendations.
	if s.cache != nil {
		_, err := s.cache.Get(ctx, cooldownKey(userID))
		if err == nil {
			// Key exists → still within cooldown → return existing from DB.
			return s.List(ctx, userID, 1, 50)
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			// Redis error — fall through and call OpenAI anyway.
			_ = err
		}
	}

	profile := ai.UserProfile{
		Goal:          "general fitness",
		ActivityLevel: "moderate",
	}

	// Enrich profile with user params if available.
	params, err := s.userParamsRepo.Get(ctx, dto.UserParamsFilter{UserID: &userID}, false)
	if err == nil && params != nil {
		profile.Weight = params.CurrentWeight()
		profile.Height = float64(params.Height())
		profile.TargetWeight = params.TargetWeight()
		if params.Want() != "" {
			profile.Goal = string(params.Want())
		}
		if params.Lifestyle() != "" {
			profile.ActivityLevel = string(params.Lifestyle())
		}
	}

	// Enrich with most recent logged weight for accurate progress tracking.
	if s.userWeightRepo != nil {
		weights, werr := s.userWeightRepo.List(ctx, dto.UserWeightFilter{UserID: &userID}, false)
		if werr == nil && len(weights) > 0 {
			latest := weights[0]
			for _, w := range weights[1:] {
				if w.Date().After(latest.Date()) {
					latest = w
				}
			}
			profile.Weight = latest.Weight()
		}
	}

	items, err := s.ai.GenerateRecommendations(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("refresh recommendations: generate: %w", err)
	}

	// Persist cooldown marker so the next call within 6 hours is served from DB.
	if s.cache != nil {
		_ = s.cache.Set(ctx, cooldownKey(userID), "1", refreshCooldown)
	}

	// Delete old recommendations and insert new ones.
	_ = s.recRepo.DeleteByUser(ctx, userID)

	result := make([]*entities.UserRecommendation, 0, len(items))
	now := time.Now()
	for _, item := range items {
		rec := entities.NewUserRecommendation(entities.WithUserRecommendationInitSpec(entities.UserRecommendationInitSpec{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        entities.RecommendationType(item.Type),
			Description: item.Description,
			Priority:    item.Priority,
			GeneratedAt: now,
		}))
		if err := s.recRepo.Create(ctx, rec); err != nil {
			return nil, fmt.Errorf("refresh recommendations: save: %w", err)
		}
		result = append(result, rec)
	}

	return result, nil
}

// MarkRead marks a recommendation as read.
func (s *Service) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.recRepo.MarkRead(ctx, id, userID); err != nil {
		return fmt.Errorf("mark recommendation read: %w", err)
	}
	return nil
}
