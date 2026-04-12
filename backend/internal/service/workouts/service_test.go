package workouts

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── mocks ──────────────────────────────────────────────────────────────────

type mockExerciseRepo struct{ mock.Mock }

func (m *mockExerciseRepo) List(ctx context.Context, f dto.ExerciseFilter, withBlock bool) ([]*entities.Exercise, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Exercise), args.Error(1)
}

func (m *mockExerciseRepo) Get(ctx context.Context, f dto.ExerciseFilter, withBlock bool) (*entities.Exercise, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Exercise), args.Error(1)
}

type mockWorkoutExerciseRepo struct{ mock.Mock }

func (m *mockWorkoutExerciseRepo) CreateBulk(ctx context.Context, exercises []entities.WorkoutsExercise) error {
	return m.Called(ctx, exercises).Error(0)
}

func (m *mockWorkoutExerciseRepo) ListSkippedExercises(ctx context.Context, userID uuid.UUID, since time.Time) ([]dto.SkippedExerciseInfo, error) {
	args := m.Called(ctx, userID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.SkippedExerciseInfo), args.Error(1)
}

type mockWorkoutsRepo struct{ mock.Mock }

func (m *mockWorkoutsRepo) TopListWithLimit(ctx context.Context, f dto.WorkoutsFilter, limit int, withBlock bool) ([]*entities.Workout, error) {
	args := m.Called(ctx, f, limit, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Workout), args.Error(1)
}

func (m *mockWorkoutsRepo) Create(ctx context.Context, w *entities.Workout) error {
	return m.Called(ctx, w).Error(0)
}

func (m *mockWorkoutsRepo) Get(ctx context.Context, f dto.WorkoutsFilter, withBlock bool) (*entities.Workout, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Workout), args.Error(1)
}

func (m *mockWorkoutsRepo) Update(ctx context.Context, w *entities.Workout) error {
	return m.Called(ctx, w).Error(0)
}

type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

type mockUserFoodRepo struct{ mock.Mock }

func (m *mockUserFoodRepo) List(ctx context.Context, f dto.UserFoodFilter) ([]*entities.UserFood, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserFood), args.Error(1)
}

type mockUserWeightRepo struct{ mock.Mock }

func (m *mockUserWeightRepo) List(ctx context.Context, f dto.UserWeightFilter, withBlock bool) ([]*entities.UserWeight, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserWeight), args.Error(1)
}

// ── helpers ────────────────────────────────────────────────────────────────

func newExercise(t entities.ExerciseType) *entities.Exercise {
	return entities.NewExercise(entities.WithExerciseInitSpec(entities.ExerciseInitSpec{
		ID:               uuid.New(),
		TypeExercise:     t,
		LevelPreparation: entities.Medium,
		PlaceExercise:    entities.Gym,
		BaseCountReps:    10,
		BaseRelaxTime:    60,
		AvgCaloriesPer:   5.0,
	}))
}

func newRNG() *rand.Rand {
	return rand.New(rand.NewSource(42))
}

func newService() *Service {
	return &Service{
		minExercisesPerWorkout: 4,
		maxExercisesPerWorkout: 12,
		rng:                    newRNG(),
		location:               time.UTC,
	}
}

// ── exercisePhase ──────────────────────────────────────────────────────────

func TestExercisePhase(t *testing.T) {
	assert.Equal(t, 0, exercisePhase(entities.Flexibility))
	assert.Equal(t, 1, exercisePhase(entities.UpperBody))
	assert.Equal(t, 1, exercisePhase(entities.LowerBody))
	assert.Equal(t, 1, exercisePhase(entities.FullBody))
	assert.Equal(t, 2, exercisePhase(entities.Cardio))
}

// ── sortExercisesByPhase ───────────────────────────────────────────────────

func TestSortExercisesByPhase(t *testing.T) {
	exercises := []*entities.Exercise{
		newExercise(entities.Cardio),
		newExercise(entities.UpperBody),
		newExercise(entities.Flexibility),
		newExercise(entities.LowerBody),
		newExercise(entities.Cardio),
		newExercise(entities.FullBody),
	}

	sorted := sortExercisesByPhase(exercises)

	// First block must be Flexibility.
	assert.Equal(t, entities.Flexibility, sorted[0].TypeExercise())
	// Last two must be Cardio.
	for _, ex := range sorted[len(sorted)-2:] {
		assert.Equal(t, entities.Cardio, ex.TypeExercise())
	}
	// Strength in the middle (positions 1-3).
	for _, ex := range sorted[1:4] {
		assert.True(t, ex.TypeExercise() == entities.UpperBody ||
			ex.TypeExercise() == entities.LowerBody ||
			ex.TypeExercise() == entities.FullBody)
	}
}

func TestSortExercisesByPhase_NoFlexibility(t *testing.T) {
	exercises := []*entities.Exercise{
		newExercise(entities.Cardio),
		newExercise(entities.UpperBody),
		newExercise(entities.LowerBody),
	}

	sorted := sortExercisesByPhase(exercises)

	// Cardio must be last.
	assert.Equal(t, entities.Cardio, sorted[len(sorted)-1].TypeExercise())
	// Strength first.
	assert.True(t, sorted[0].TypeExercise() == entities.UpperBody ||
		sorted[0].TypeExercise() == entities.LowerBody)
}

// ── filterSkippedExercises ─────────────────────────────────────────────────

func TestFilterSkippedExercises_NilMap(t *testing.T) {
	svc := &Service{}
	exercises := []*entities.Exercise{newExercise(entities.UpperBody), newExercise(entities.Cardio)}
	result := svc.filterSkippedExercises(exercises, nil)
	assert.Len(t, result, 2)
}

func TestFilterSkippedExercises_SkippedOnce_Included(t *testing.T) {
	svc := &Service{}
	ex := newExercise(entities.UpperBody)
	skipMap := map[uuid.UUID]dto.SkippedExerciseInfo{
		ex.ID(): {ExerciseID: ex.ID(), SkipCount: 1, LastSkippedAt: time.Now().Add(-1 * time.Hour)},
	}
	result := svc.filterSkippedExercises([]*entities.Exercise{ex}, skipMap)
	assert.Len(t, result, 1, "skipped once should still be included")
}

func TestFilterSkippedExercises_SkippedTwiceRecently_Excluded(t *testing.T) {
	svc := &Service{}
	ex := newExercise(entities.UpperBody)
	skipMap := map[uuid.UUID]dto.SkippedExerciseInfo{
		ex.ID(): {ExerciseID: ex.ID(), SkipCount: 2, LastSkippedAt: time.Now().Add(-1 * time.Hour)},
	}
	result := svc.filterSkippedExercises([]*entities.Exercise{ex}, skipMap)
	assert.Len(t, result, 0, "skipped twice recently should be blocked")
}

func TestFilterSkippedExercises_SkippedTwiceOld_Included(t *testing.T) {
	svc := &Service{}
	ex := newExercise(entities.UpperBody)
	// Last skip more than 7 days ago → block lifted.
	skipMap := map[uuid.UUID]dto.SkippedExerciseInfo{
		ex.ID(): {ExerciseID: ex.ID(), SkipCount: 2, LastSkippedAt: time.Now().Add(-8 * 24 * time.Hour)},
	}
	result := svc.filterSkippedExercises([]*entities.Exercise{ex}, skipMap)
	assert.Len(t, result, 1, "skip block expires after 7 days")
}

// ── nutritionCoefAdjustment ────────────────────────────────────────────────

func TestNutritionCoefAdjustment(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name     string
		consumed int
		target   int
		goal     entities.Want
		wantCoef float64
	}{
		{"zero target → no change", 2000, 0, entities.LoseWeight, 1.0},
		{"surplus >300 → harder", 2500, 2000, entities.LoseWeight, 1.2},
		{"deficit >300 → easier", 1500, 2000, entities.LoseWeight, 0.8},
		{"deficit >300 build_muscle → no reduction", 1500, 2000, entities.BuildMuscle, 1.0},
		{"near target → no change", 2100, 2000, entities.LoseWeight, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.nutritionCoefAdjustment(tt.consumed, tt.target, tt.goal)
			assert.Equal(t, tt.wantCoef, got)
		})
	}
}

// ── weightProgressTypePreference ──────────────────────────────────────────

func TestWeightProgressTypePreference(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name     string
		delta    float64 // current - target
		goal     entities.Want
		wantType entities.ExerciseType
	}{
		{"at target → no preference", 0, entities.LoseWeight, ""},
		{"small delta → no preference", 0.5, entities.LoseWeight, ""},
		{"need to lose, not BuildMuscle → Cardio", 5.0, entities.LoseWeight, entities.Cardio},
		{"need to lose, BuildMuscle → FullBody", 5.0, entities.BuildMuscle, entities.FullBody},
		{"need to gain → UpperBody", -3.0, entities.BuildMuscle, entities.UpperBody},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.weightProgressTypePreference(tt.delta, tt.goal)
			assert.Equal(t, tt.wantType, got)
		})
	}
}

// ── getTodayCalories ──────────────────────────────────────────────────────

func TestGetTodayCalories_NoFoodRepo(t *testing.T) {
	svc := &Service{userFoodRepository: nil}
	assert.Equal(t, 0, svc.getTodayCalories(context.Background(), uuid.New(), time.Now()))
}

func TestGetTodayCalories_SumsEntries(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	food1 := entities.NewUserFood(entities.WithUserFoodInitSpec(entities.UserFoodInitSpec{
		ID: uuid.New(), UserID: userID, Calories: 500, MealType: entities.MealTypeBreakfast, Date: now,
	}))
	food2 := entities.NewUserFood(entities.WithUserFoodInitSpec(entities.UserFoodInitSpec{
		ID: uuid.New(), UserID: userID, Calories: 700, MealType: entities.MealTypeLunch, Date: now,
	}))

	foodRepo := &mockUserFoodRepo{}
	foodRepo.On("List", mock.Anything, mock.MatchedBy(func(f dto.UserFoodFilter) bool {
		return f.UserID != nil && *f.UserID == userID && f.Date != nil
	})).Return([]*entities.UserFood{food1, food2}, nil)

	svc := &Service{userFoodRepository: foodRepo, location: time.UTC}
	total := svc.getTodayCalories(ctx, userID, now.UTC())
	assert.Equal(t, 1200, total)
	foodRepo.AssertExpectations(t)
}

// ── getLatestWeight ──────────────────────────────────────────────────────

func TestGetLatestWeight_NoWeightRepo(t *testing.T) {
	svc := &Service{userWeightRepository: nil}
	assert.Equal(t, 0.0, svc.getLatestWeight(context.Background(), uuid.New()))
}

func TestGetLatestWeight_ReturnsNewest(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	older := entities.NewUserWeight(entities.WithUserWeightInitSpec(entities.UserWeightInitSpec{
		ID: uuid.New(), UserID: userID, Weight: 80.0,
		Date: time.Now().Add(-48 * time.Hour),
	}))
	newer := entities.NewUserWeight(entities.WithUserWeightInitSpec(entities.UserWeightInitSpec{
		ID: uuid.New(), UserID: userID, Weight: 78.5,
		Date: time.Now().Add(-24 * time.Hour),
	}))

	weightRepo := &mockUserWeightRepo{}
	weightRepo.On("List", mock.Anything, mock.AnythingOfType("dto.UserWeightFilter"), false).
		Return([]*entities.UserWeight{older, newer}, nil)

	svc := &Service{userWeightRepository: weightRepo}
	w := svc.getLatestWeight(ctx, userID)
	assert.Equal(t, 78.5, w)
	weightRepo.AssertExpectations(t)
}

// ── buildSkipMap ───────────────────────────────────────────────────────────

func TestBuildSkipMap_ReturnsMapByExerciseID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	exID := uuid.New()

	info := dto.SkippedExerciseInfo{ExerciseID: exID, SkipCount: 1, LastSkippedAt: time.Now()}

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.AnythingOfType("time.Time")).
		Return([]dto.SkippedExerciseInfo{info}, nil)

	svc := &Service{workoutExerciseRepository: weRepo}
	m, err := svc.buildSkipMap(ctx, userID)
	assert.NoError(t, err)
	assert.Contains(t, m, exID)
	assert.Equal(t, 1, m[exID].SkipCount)
}
