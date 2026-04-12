package recomendation

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/ai"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── mocks ──────────────────────────────────────────────────────

type mockRecRepo struct{ mock.Mock }

func (m *mockRecRepo) Create(ctx context.Context, r *entities.UserRecommendation) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mockRecRepo) Get(ctx context.Context, f dto.UserRecommendationFilter) (*entities.UserRecommendation, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserRecommendation), args.Error(1)
}
func (m *mockRecRepo) List(ctx context.Context, f dto.UserRecommendationFilter) ([]*entities.UserRecommendation, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRecommendation), args.Error(1)
}
func (m *mockRecRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockRecRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

type mockParamsRepo struct{ mock.Mock }

func (m *mockParamsRepo) Get(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserParams), args.Error(1)
}

type mockWeightRepo struct{ mock.Mock }

func (m *mockWeightRepo) List(ctx context.Context, f dto.UserWeightFilter, withBlock bool) ([]*entities.UserWeight, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserWeight), args.Error(1)
}

type mockAI struct{ mock.Mock }

func (m *mockAI) GenerateRecommendations(ctx context.Context, profile ai.UserProfile) ([]ai.RecommendationItem, error) {
	args := m.Called(ctx, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ai.RecommendationItem), args.Error(1)
}

// ── helpers ────────────────────────────────────────────────────

func newRec(userID uuid.UUID) *entities.UserRecommendation {
	return entities.NewUserRecommendation(entities.WithUserRecommendationInitSpec(entities.UserRecommendationInitSpec{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        entities.RecommendationTypeGeneral,
		Description: "Drink more water",
		Priority:    2,
		GeneratedAt: time.Now(),
	}))
}

var aiItems = []ai.RecommendationItem{
	{Type: "workout", Description: "Do 30 min cardio", Priority: 1},
	{Type: "nutrition", Description: "Eat more protein", Priority: 2},
}

// ── List ───────────────────────────────────────────────────────

func TestService_List(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	limit10 := 10
	offset0 := 0

	tests := []struct {
		name    string
		page    int
		limit   int
		setup   func(r *mockRecRepo)
		wantLen int
		wantErr bool
	}{
		{
			name:  "default pagination",
			page:  1,
			limit: 10,
			setup: func(r *mockRecRepo) {
				r.On("List", mock.Anything, dto.UserRecommendationFilter{UserID: &userID, Limit: &limit10, Offset: &offset0}).
					Return([]*entities.UserRecommendation{newRec(userID), newRec(userID)}, nil)
			},
			wantLen: 2,
		},
		{
			name:  "empty result",
			page:  1,
			limit: 10,
			setup: func(r *mockRecRepo) {
				r.On("List", mock.Anything, mock.Anything).Return([]*entities.UserRecommendation{}, nil)
			},
			wantLen: 0,
		},
		{
			name:  "invalid page/limit defaults to 1/10",
			page:  0,
			limit: 0,
			setup: func(r *mockRecRepo) {
				r.On("List", mock.Anything, dto.UserRecommendationFilter{UserID: &userID, Limit: &limit10, Offset: &offset0}).
					Return([]*entities.UserRecommendation{newRec(userID)}, nil)
			},
			wantLen: 1,
		},
		{
			name:  "repo error",
			page:  1,
			limit: 10,
			setup: func(r *mockRecRepo) {
				r.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recRepo := &mockRecRepo{}
			tt.setup(recRepo)
			s := NewService(&Config{
				RecommendationRepository: recRepo,
				UserParamsRepository:     &mockParamsRepo{},
				UserWeightRepository:     &mockWeightRepo{},
				AIClient:                 &mockAI{},
			})
			recs, err := s.List(ctx, userID, tt.page, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, recs, tt.wantLen)
			}
		})
	}
}

// ── Refresh ────────────────────────────────────────────────────

func TestService_Refresh(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		setup   func(recRepo *mockRecRepo, paramsRepo *mockParamsRepo, aiMock *mockAI)
		wantLen int
		wantErr bool
	}{
		{
			name: "success — old recs deleted, new ones saved",
			setup: func(recRepo *mockRecRepo, paramsRepo *mockParamsRepo, aiMock *mockAI) {
				paramsRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("no params"))
				aiMock.On("GenerateRecommendations", mock.Anything, mock.Anything).Return(aiItems, nil)
				recRepo.On("DeleteByUser", mock.Anything, userID).Return(nil)
				recRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserRecommendation")).Return(nil).Times(2)
			},
			wantLen: 2,
		},
		{
			name: "ai error propagates",
			setup: func(recRepo *mockRecRepo, paramsRepo *mockParamsRepo, aiMock *mockAI) {
				paramsRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("no params"))
				aiMock.On("GenerateRecommendations", mock.Anything, mock.Anything).Return(nil, errors.New("openai down"))
			},
			wantErr: true,
		},
		{
			name: "create fails — error propagates",
			setup: func(recRepo *mockRecRepo, paramsRepo *mockParamsRepo, aiMock *mockAI) {
				paramsRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("no params"))
				aiMock.On("GenerateRecommendations", mock.Anything, mock.Anything).Return(aiItems, nil)
				recRepo.On("DeleteByUser", mock.Anything, userID).Return(nil)
				recRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recRepo := &mockRecRepo{}
			paramsRepo := &mockParamsRepo{}
			aiMock := &mockAI{}
			tt.setup(recRepo, paramsRepo, aiMock)
			weightRepo := &mockWeightRepo{}
			weightRepo.On("List", mock.Anything, mock.Anything, false).Return([]*entities.UserWeight{}, nil).Maybe()
			s := NewService(&Config{
				RecommendationRepository: recRepo,
				UserParamsRepository:     paramsRepo,
				UserWeightRepository:     weightRepo,
				AIClient:                 aiMock,
			})
			recs, err := s.Refresh(ctx, userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, recs, tt.wantLen)
			}
		})
	}
}

// ── MarkRead ───────────────────────────────────────────────────

func TestService_MarkRead(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	recID := uuid.New()

	tests := []struct {
		name    string
		setup   func(r *mockRecRepo)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(r *mockRecRepo) {
				r.On("MarkRead", mock.Anything, recID, userID).Return(nil)
			},
		},
		{
			name: "repo error",
			setup: func(r *mockRecRepo) {
				r.On("MarkRead", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recRepo := &mockRecRepo{}
			tt.setup(recRepo)
			s := NewService(&Config{
				RecommendationRepository: recRepo,
				UserParamsRepository:     &mockParamsRepo{},
				UserWeightRepository:     &mockWeightRepo{},
				AIClient:                 &mockAI{},
			})
			err := s.MarkRead(ctx, recID, userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
