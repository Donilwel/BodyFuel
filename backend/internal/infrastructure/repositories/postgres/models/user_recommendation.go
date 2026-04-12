package models

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type UserRecommendationRow struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	Type        string    `db:"type"`
	Description string    `db:"description"`
	Priority    int       `db:"priority"`
	IsRead      bool      `db:"is_read"`
	GeneratedAt time.Time `db:"generated_at"`
	CreatedAt   time.Time `db:"created_at"`
}

func NewUserRecommendationRow(r *entities.UserRecommendation) *UserRecommendationRow {
	return &UserRecommendationRow{
		ID:          r.ID(),
		UserID:      r.UserID(),
		Type:        string(r.Type()),
		Description: r.Description(),
		Priority:    r.Priority(),
		IsRead:      r.IsRead(),
		GeneratedAt: r.GeneratedAt(),
		CreatedAt:   r.CreatedAt(),
	}
}

func (r *UserRecommendationRow) ToEntity() *entities.UserRecommendation {
	return entities.NewUserRecommendation(entities.WithUserRecommendationRestoreSpec(entities.UserRecommendationRestoreSpec{
		ID:          r.ID,
		UserID:      r.UserID,
		Type:        entities.RecommendationType(r.Type),
		Description: r.Description,
		Priority:    r.Priority,
		IsRead:      r.IsRead,
		GeneratedAt: r.GeneratedAt,
		CreatedAt:   r.CreatedAt,
	}))
}
