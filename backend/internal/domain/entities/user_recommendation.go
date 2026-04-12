package entities

import (
	"github.com/google/uuid"
	"time"
)

type RecommendationType string

const (
	RecommendationTypeWorkout  RecommendationType = "workout"
	RecommendationTypeNutrition RecommendationType = "nutrition"
	RecommendationTypeRest     RecommendationType = "rest"
	RecommendationTypeGeneral  RecommendationType = "general"
)

type UserRecommendation struct {
	id          uuid.UUID
	userID      uuid.UUID
	recType     RecommendationType
	description string
	priority    int
	isRead      bool
	generatedAt time.Time
	createdAt   time.Time
}

func (r *UserRecommendation) ID() uuid.UUID                  { return r.id }
func (r *UserRecommendation) UserID() uuid.UUID               { return r.userID }
func (r *UserRecommendation) Type() RecommendationType        { return r.recType }
func (r *UserRecommendation) Description() string             { return r.description }
func (r *UserRecommendation) Priority() int                   { return r.priority }
func (r *UserRecommendation) IsRead() bool                    { return r.isRead }
func (r *UserRecommendation) GeneratedAt() time.Time          { return r.generatedAt }
func (r *UserRecommendation) CreatedAt() time.Time            { return r.createdAt }

func (r *UserRecommendation) MarkRead() {
	r.isRead = true
}

type UserRecommendationOption func(r *UserRecommendation)

func NewUserRecommendation(opt UserRecommendationOption) *UserRecommendation {
	r := new(UserRecommendation)
	opt(r)
	return r
}

type UserRecommendationInitSpec struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Type        RecommendationType
	Description string
	Priority    int
	GeneratedAt time.Time
}

type UserRecommendationRestoreSpec struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Type        RecommendationType
	Description string
	Priority    int
	IsRead      bool
	GeneratedAt time.Time
	CreatedAt   time.Time
}

func WithUserRecommendationInitSpec(s UserRecommendationInitSpec) UserRecommendationOption {
	return func(r *UserRecommendation) {
		r.id = s.ID
		r.userID = s.UserID
		r.recType = s.Type
		r.description = s.Description
		r.priority = s.Priority
		r.isRead = false
		r.generatedAt = s.GeneratedAt
		r.createdAt = time.Now()
	}
}

func WithUserRecommendationRestoreSpec(s UserRecommendationRestoreSpec) UserRecommendationOption {
	return func(r *UserRecommendation) {
		r.id = s.ID
		r.userID = s.UserID
		r.recType = s.Type
		r.description = s.Description
		r.priority = s.Priority
		r.isRead = s.IsRead
		r.generatedAt = s.GeneratedAt
		r.createdAt = s.CreatedAt
	}
}
