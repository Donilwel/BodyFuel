package models

import (
	"backend/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type RecommendationResponse struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	IsRead      bool      `json:"is_read"`
	GeneratedAt time.Time `json:"generated_at"`
}

func NewRecommendationResponse(r *entities.UserRecommendation) RecommendationResponse {
	return RecommendationResponse{
		ID:          r.ID(),
		Type:        string(r.Type()),
		Description: r.Description(),
		Priority:    r.Priority(),
		IsRead:      r.IsRead(),
		GeneratedAt: r.GeneratedAt(),
	}
}

func NewRecommendationResponseList(list []*entities.UserRecommendation) []RecommendationResponse {
	result := make([]RecommendationResponse, len(list))
	for i, r := range list {
		result[i] = NewRecommendationResponse(r)
	}
	return result
}
