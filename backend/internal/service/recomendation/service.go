package recomendation

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

	AIClient interface {
		GenerateRecommendations(ctx context.Context, profile ai.UserProfile) ([]ai.RecommendationItem, error)
	}
)

type Service struct {
	recRepo        UserRecommendationRepository
	userParamsRepo UserParamsRepository
	ai             AIClient
}

type Config struct {
	RecommendationRepository UserRecommendationRepository
	UserParamsRepository     UserParamsRepository
	AIClient                 AIClient
}

func NewService(c *Config) *Service {
	return &Service{
		recRepo:        c.RecommendationRepository,
		userParamsRepo: c.UserParamsRepository,
		ai:             c.AIClient,
	}
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

// Refresh generates new recommendations for a user using OpenAI, replacing existing ones.
func (s *Service) Refresh(ctx context.Context, userID uuid.UUID) ([]*entities.UserRecommendation, error) {
	profile := ai.UserProfile{
		Goal:          "general fitness",
		ActivityLevel: "moderate",
	}

	// Enrich profile with user params if available.
	params, err := s.userParamsRepo.Get(ctx, dto.UserParamsFilter{UserID: &userID}, false)
	if err == nil && params != nil {
		profile.Weight = params.CurrentWeight()
		profile.Height = float64(params.Height())
		if params.Want() != "" {
			profile.Goal = string(params.Want())
		}
		if params.Lifestyle() != "" {
			profile.ActivityLevel = string(params.Lifestyle())
		}
	}

	items, err := s.ai.GenerateRecommendations(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("refresh recommendations: generate: %w", err)
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
