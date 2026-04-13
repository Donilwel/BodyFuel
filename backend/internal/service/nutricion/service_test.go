package nutricion

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

type mockFoodRepo struct{ mock.Mock }

func (m *mockFoodRepo) Create(ctx context.Context, f *entities.UserFood) error {
	return m.Called(ctx, f).Error(0)
}
func (m *mockFoodRepo) Get(ctx context.Context, f dto.UserFoodFilter) (*entities.UserFood, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserFood), args.Error(1)
}
func (m *mockFoodRepo) List(ctx context.Context, f dto.UserFoodFilter) ([]*entities.UserFood, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserFood), args.Error(1)
}
func (m *mockFoodRepo) Update(ctx context.Context, f *entities.UserFood) error {
	return m.Called(ctx, f).Error(0)
}
func (m *mockFoodRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

type mockAIClient struct{ mock.Mock }

func (m *mockAIClient) AnalyzeNutritionPhoto(ctx context.Context, imageURL string) (*ai.NutritionAnalysis, error) {
	args := m.Called(ctx, imageURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.NutritionAnalysis), args.Error(1)
}
func (m *mockAIClient) GenerateRecipes(ctx context.Context, intake ai.DailyIntake) ([]ai.RecipeItem, error) {
	args := m.Called(ctx, intake)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ai.RecipeItem), args.Error(1)
}

// ── helpers ────────────────────────────────────────────────────

func newFoodEntry(userID uuid.UUID, cal int, protein, carbs, fat float64) *entities.UserFood {
	return entities.NewUserFood(entities.WithUserFoodInitSpec(entities.UserFoodInitSpec{
		ID:          uuid.New(),
		UserID:      userID,
		Description: "test food",
		Calories:    cal,
		Protein:     protein,
		Carbs:       carbs,
		Fat:         fat,
		MealType:    entities.MealTypeLunch,
		Date:        time.Now(),
	}))
}

func makeEntry(id, userID uuid.UUID) *entities.UserFood {
	return entities.NewUserFood(entities.WithUserFoodInitSpec(entities.UserFoodInitSpec{
		ID:       id,
		UserID:   userID,
		MealType: entities.MealTypeLunch,
		Date:     time.Now(),
	}))
}

// ── CreateFoodEntry ────────────────────────────────────────────

func TestService_CreateFoodEntry(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		setup   func(r *mockFoodRepo)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(r *mockFoodRepo) {
				r.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserFood")).Return(nil)
			},
		},
		{
			name: "repo error",
			setup: func(r *mockFoodRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFoodRepo{}
			tt.setup(repo)
			s := NewService(&Config{UserFoodRepository: repo, AIClient: &mockAIClient{}})
			err := s.CreateFoodEntry(ctx, entities.UserFoodInitSpec{
				ID: uuid.New(), UserID: userID, Description: "egg", Calories: 80, MealType: entities.MealTypeBreakfast, Date: time.Now(),
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── GetDiary ───────────────────────────────────────────────────

func TestService_GetDiary(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	today := time.Now()

	entry1 := newFoodEntry(userID, 400, 30, 50, 10)
	entry2 := newFoodEntry(userID, 300, 20, 40, 8)

	tests := []struct {
		name          string
		setup         func(r *mockFoodRepo)
		wantCalories  int
		wantProtein   float64
		wantErr       bool
	}{
		{
			name: "aggregates totals correctly",
			setup: func(r *mockFoodRepo) {
				r.On("List", mock.Anything, dto.UserFoodFilter{UserID: &userID, Date: &today}).
					Return([]*entities.UserFood{entry1, entry2}, nil)
			},
			wantCalories: 700,
			wantProtein:  50,
		},
		{
			name: "empty diary — zero totals",
			setup: func(r *mockFoodRepo) {
				r.On("List", mock.Anything, mock.Anything).Return([]*entities.UserFood{}, nil)
			},
			wantCalories: 0,
			wantProtein:  0,
		},
		{
			name: "repo error",
			setup: func(r *mockFoodRepo) {
				r.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFoodRepo{}
			tt.setup(repo)
			s := NewService(&Config{UserFoodRepository: repo, AIClient: &mockAIClient{}})
			diary, err := s.GetDiary(ctx, userID, today)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, diary)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCalories, diary.TotalCalories)
				assert.Equal(t, tt.wantProtein, diary.TotalProtein)
			}
		})
	}
}

// ── GetReport ──────────────────────────────────────────────────

func TestService_GetReport(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC) // 7 days

	entries := []*entities.UserFood{
		newFoodEntry(userID, 2000, 150, 200, 60),
		newFoodEntry(userID, 2100, 160, 210, 65),
	}

	tests := []struct {
		name         string
		setup        func(r *mockFoodRepo)
		wantDays     int
		wantAvgCal   float64
		wantErr      bool
	}{
		{
			name: "calculates days and average correctly",
			setup: func(r *mockFoodRepo) {
				r.On("List", mock.Anything, mock.Anything).Return(entries, nil)
			},
			wantDays:   7,
			wantAvgCal: float64(2000+2100) / 7,
		},
		{
			name: "repo error",
			setup: func(r *mockFoodRepo) {
				r.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFoodRepo{}
			tt.setup(repo)
			s := NewService(&Config{UserFoodRepository: repo, AIClient: &mockAIClient{}})
			report, err := s.GetReport(ctx, userID, from, to)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDays, report.Days)
				assert.InDelta(t, tt.wantAvgCal, report.AvgCalories, 0.01)
			}
		})
	}
}

// ── UpdateFoodEntry ────────────────────────────────────────────

func TestService_UpdateFoodEntry(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	entryID := uuid.New()
	entry := newFoodEntry(userID, 400, 30, 50, 10)

	newCal := 500

	tests := []struct {
		name    string
		setup   func(r *mockFoodRepo)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(r *mockFoodRepo) {
				r.On("Get", mock.Anything, dto.UserFoodFilter{ID: &entryID, UserID: &userID}).Return(entry, nil)
				r.On("Update", mock.Anything, mock.MatchedBy(func(f *entities.UserFood) bool {
					return f.Calories() == newCal
				})).Return(nil)
			},
		},
		{
			name: "entry not found",
			setup: func(r *mockFoodRepo) {
				r.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFoodRepo{}
			tt.setup(repo)
			s := NewService(&Config{UserFoodRepository: repo, AIClient: &mockAIClient{}})
			err := s.UpdateFoodEntry(ctx, entryID, userID, entities.UserFoodUpdateParams{Calories: &newCal})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── DeleteFoodEntry ────────────────────────────────────────────

func TestService_DeleteFoodEntry(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	entryID := uuid.New()

	tests := []struct {
		name    string
		setup   func(r *mockFoodRepo)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(r *mockFoodRepo) {
				// Get is called first to obtain the date for cache invalidation.
				r.On("Get", mock.Anything, mock.Anything).Return(makeEntry(entryID, userID), nil)
				r.On("Delete", mock.Anything, entryID, userID).Return(nil)
			},
		},
		{
			name: "repo error",
			setup: func(r *mockFoodRepo) {
				r.On("Get", mock.Anything, mock.Anything).Return((*entities.UserFood)(nil), errors.New("not found"))
				r.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFoodRepo{}
			tt.setup(repo)
			s := NewService(&Config{UserFoodRepository: repo, AIClient: &mockAIClient{}})
			err := s.DeleteFoodEntry(ctx, entryID, userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── AnalyzePhoto ───────────────────────────────────────────────

func TestService_AnalyzePhoto(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(a *mockAIClient)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(a *mockAIClient) {
				a.On("AnalyzeNutritionPhoto", mock.Anything, "http://img.example.com/food.jpg").
					Return(&ai.NutritionAnalysis{Description: "salad", Calories: 250, Protein: 10, Carbs: 30, Fat: 5}, nil)
			},
		},
		{
			name: "ai error",
			setup: func(a *mockAIClient) {
				a.On("AnalyzeNutritionPhoto", mock.Anything, mock.Anything).Return(nil, errors.New("openai error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aiMock := &mockAIClient{}
			tt.setup(aiMock)
			s := NewService(&Config{UserFoodRepository: &mockFoodRepo{}, AIClient: aiMock})
			result, err := s.AnalyzePhoto(ctx, "http://img.example.com/food.jpg")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "salad", result.Description)
				assert.Equal(t, 250, result.Calories)
			}
		})
	}
}

// ── RecommendRecipes ───────────────────────────────────────────

func TestService_RecommendRecipes(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	today := time.Now()

	entry := newFoodEntry(userID, 800, 60, 90, 25)

	aiRecipes := []ai.RecipeItem{
		{Name: "Тунец с авокадо", Description: "Лёгкий ужин", Macros: ai.MacroNutrients{Protein: 32, Fat: 14, Carbs: 4}, PreparationTime: 5},
		{Name: "Гречка с курицей", Description: "Сытный обед", Macros: ai.MacroNutrients{Protein: 40, Fat: 8, Carbs: 50}, PreparationTime: 20},
	}

	tests := []struct {
		name    string
		setup   func(r *mockFoodRepo, a *mockAIClient)
		wantLen int
		wantErr bool
	}{
		{
			name: "success — passes diary totals to AI",
			setup: func(r *mockFoodRepo, a *mockAIClient) {
				r.On("List", mock.Anything, dto.UserFoodFilter{UserID: &userID, Date: &today}).
					Return([]*entities.UserFood{entry}, nil)
				a.On("GenerateRecipes", mock.Anything, ai.DailyIntake{
					ConsumedCalories: 800,
					ConsumedProtein:  60,
					ConsumedCarbs:    90,
					ConsumedFat:      25,
				}).Return(aiRecipes, nil)
			},
			wantLen: 2,
		},
		{
			name: "diary error propagates",
			setup: func(r *mockFoodRepo, a *mockAIClient) {
				r.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "ai error propagates",
			setup: func(r *mockFoodRepo, a *mockAIClient) {
				r.On("List", mock.Anything, mock.Anything).Return([]*entities.UserFood{}, nil)
				a.On("GenerateRecipes", mock.Anything, mock.Anything).Return(nil, errors.New("openai error"))
			},
			wantErr: true,
		},
		{
			name: "empty diary — still calls AI with zeros",
			setup: func(r *mockFoodRepo, a *mockAIClient) {
				r.On("List", mock.Anything, mock.Anything).Return([]*entities.UserFood{}, nil)
				a.On("GenerateRecipes", mock.Anything, ai.DailyIntake{}).Return(aiRecipes, nil)
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFoodRepo{}
			aiMock := &mockAIClient{}
			tt.setup(repo, aiMock)
			s := NewService(&Config{UserFoodRepository: repo, AIClient: aiMock})
			recipes, err := s.RecommendRecipes(ctx, userID, today)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, recipes)
			} else {
				assert.NoError(t, err)
				assert.Len(t, recipes, tt.wantLen)
			}
		})
	}
}
