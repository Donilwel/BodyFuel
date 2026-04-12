package workouts

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"context"
	"errors"
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

func (m *mockWorkoutExerciseRepo) ListExerciseProgress(ctx context.Context, userID uuid.UUID, since time.Time) ([]dto.ExerciseProgressInfo, error) {
	args := m.Called(ctx, userID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.ExerciseProgressInfo), args.Error(1)
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

type mockUserParamsRepo struct{ mock.Mock }

func (m *mockUserParamsRepo) List(ctx context.Context, f dto.UserParamsFilter, withBlock bool) ([]*entities.UserParams, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserParams), args.Error(1)
}

type mockUserInfoRepo struct{ mock.Mock }

func (m *mockUserInfoRepo) Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserInfo), args.Error(1)
}

func (m *mockUserInfoRepo) GetBatch(ctx context.Context, ids []uuid.UUID) ([]*entities.UserInfo, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserInfo), args.Error(1)
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
		metrics:                &Metrics{},
	}
}

func newUserParams(userID uuid.UUID) *entities.UserParams {
	return entities.NewUserParams(entities.WithUserParamsInitSpec(entities.UserParamsInitSpec{
		ID:                  uuid.New(),
		UserID:              userID,
		Height:              175,
		Wants:               entities.LoseWeight,
		Lifestyle:           entities.Active,
		TargetCaloriesDaily: 2000,
		TargetWorkoutsWeeks: 3,
		TargetWeight:        70.0,
	}))
}

func newUserInfo(userID uuid.UUID) *entities.UserInfo {
	return entities.NewUserInfo(entities.WithUserInfoRestoreSpec(entities.UserInfoRestoreSpec{
		ID:    userID,
		Email: "user@example.com",
		Phone: "+79991234567",
	}))
}

func newWorkout(userID uuid.UUID, status entities.WorkoutsStatus, level entities.WorkoutsLevel, createdAt time.Time) *entities.Workout {
	return entities.NewWorkout(entities.WithWorkoutRestoreSpec(entities.WorkoutRestoreSpec{
		ID:        uuid.New(),
		UserID:    userID,
		Level:     level,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}))
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

	assert.Equal(t, entities.Flexibility, sorted[0].TypeExercise())
	for _, ex := range sorted[len(sorted)-2:] {
		assert.Equal(t, entities.Cardio, ex.TypeExercise())
	}
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

	assert.Equal(t, entities.Cardio, sorted[len(sorted)-1].TypeExercise())
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
	assert.Len(t, result, 1)
}

func TestFilterSkippedExercises_SkippedTwiceRecently_Excluded(t *testing.T) {
	svc := &Service{}
	ex := newExercise(entities.UpperBody)
	skipMap := map[uuid.UUID]dto.SkippedExerciseInfo{
		ex.ID(): {ExerciseID: ex.ID(), SkipCount: 2, LastSkippedAt: time.Now().Add(-1 * time.Hour)},
	}
	result := svc.filterSkippedExercises([]*entities.Exercise{ex}, skipMap)
	assert.Len(t, result, 0)
}

func TestFilterSkippedExercises_SkippedTwiceOld_Included(t *testing.T) {
	svc := &Service{}
	ex := newExercise(entities.UpperBody)
	skipMap := map[uuid.UUID]dto.SkippedExerciseInfo{
		ex.ID(): {ExerciseID: ex.ID(), SkipCount: 2, LastSkippedAt: time.Now().Add(-8 * 24 * time.Hour)},
	}
	result := svc.filterSkippedExercises([]*entities.Exercise{ex}, skipMap)
	assert.Len(t, result, 1)
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
		delta    float64
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

// ── shouldSkipGeneration ───────────────────────────────────────────────────

func TestShouldSkipGeneration_ActiveWorkout(t *testing.T) {
	svc := &Service{limitGenerateWorkouts: 3}
	now := time.Now()
	active := newWorkout(uuid.New(), entities.WorkoutStatusInActive, entities.WorkoutLight, now.Add(-1*time.Hour))
	reason := svc.shouldSkipGeneration([]*entities.Workout{active}, active, 3, now)
	assert.NotEmpty(t, reason)
}

func TestShouldSkipGeneration_TooManyUnused(t *testing.T) {
	svc := &Service{limitGenerateWorkouts: 2}
	now := time.Now()
	w1 := newWorkout(uuid.New(), entities.WorkoutStatusCreated, entities.WorkoutLight, now.Add(-25*time.Hour))
	w2 := newWorkout(uuid.New(), entities.WorkoutStatusCreated, entities.WorkoutLight, now.Add(-26*time.Hour))
	reason := svc.shouldSkipGeneration([]*entities.Workout{w1, w2}, w1, 3, now)
	assert.NotEmpty(t, reason)
}

func TestShouldSkipGeneration_RestPeriod(t *testing.T) {
	svc := &Service{limitGenerateWorkouts: 5}
	now := time.Now()
	recent := newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now.Add(-1*time.Hour))
	reason := svc.shouldSkipGeneration([]*entities.Workout{recent}, recent, 3, now)
	assert.NotEmpty(t, reason)
}

func TestShouldSkipGeneration_WeeklyTargetMet(t *testing.T) {
	svc := &Service{limitGenerateWorkouts: 5}
	now := time.Now()
	w1 := newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now.Add(-25*time.Hour))
	w2 := newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now.Add(-26*time.Hour))
	w3 := newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now.Add(-27*time.Hour))
	last := w1
	reason := svc.shouldSkipGeneration([]*entities.Workout{w1, w2, w3}, last, 3, now)
	assert.NotEmpty(t, reason)
}

func TestShouldSkipGeneration_ShouldNotSkip(t *testing.T) {
	svc := &Service{limitGenerateWorkouts: 5}
	now := time.Now()
	old := newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now.Add(-48*time.Hour))
	reason := svc.shouldSkipGeneration([]*entities.Workout{old}, old, 5, now)
	assert.Empty(t, reason)
}

// ── getTargetWorkoutsPerWeek ───────────────────────────────────────────────

func TestGetTargetWorkoutsPerWeek_WithParams(t *testing.T) {
	svc := &Service{}
	userID := uuid.New()
	params := newUserParams(userID)
	assert.Equal(t, 3, svc.getTargetWorkoutsPerWeek(params))
}

func TestGetTargetWorkoutsPerWeek_NilParams(t *testing.T) {
	svc := &Service{}
	assert.Equal(t, 3, svc.getTargetWorkoutsPerWeek(nil))
}

// ── countWorkoutsByStatus ──────────────────────────────────────────────────

func TestCountWorkoutsByStatus(t *testing.T) {
	svc := &Service{}
	now := time.Now()
	workouts := []*entities.Workout{
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now),
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now),
		newWorkout(uuid.New(), entities.WorkoutStatusCreated, entities.WorkoutLight, now),
	}
	assert.Equal(t, 2, svc.countWorkoutsByStatus(workouts, entities.WorkoutStatusDone))
	assert.Equal(t, 1, svc.countWorkoutsByStatus(workouts, entities.WorkoutStatusCreated))
	assert.Equal(t, 0, svc.countWorkoutsByStatus(workouts, entities.WorkoutStatusFailed))
}

// ── countFinishedWorkoutsForWeek ───────────────────────────────────────────

func TestCountFinishedWorkoutsForWeek(t *testing.T) {
	svc := &Service{}
	now := time.Now()
	workouts := []*entities.Workout{
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now.Add(-1*24*time.Hour)),
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now.Add(-2*24*time.Hour)),
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now.Add(-10*24*time.Hour)), // old
		newWorkout(uuid.New(), entities.WorkoutStatusCreated, entities.WorkoutLight, now.Add(-1*24*time.Hour)),
	}
	assert.Equal(t, 2, svc.countFinishedWorkoutsForWeek(workouts, now))
}

// ── calculateWorkoutParams ─────────────────────────────────────────────────

func TestCalculateWorkoutParams(t *testing.T) {
	svc := newService()
	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.LowerBody),
		newExercise(entities.Cardio),
	}
	calories, duration := svc.calculateWorkoutParams(exercises, 1.0)
	assert.Greater(t, calories, 0)
	assert.Greater(t, duration, 0)
	// 3 exercises → 2 rest periods of 60s each
	_, durationSingle := svc.calculateWorkoutParams(exercises[:1], 1.0)
	assert.Equal(t, duration-durationSingle*3-2*restBetweenExercises, 0)
}

func TestCalculateWorkoutParams_CoefScales(t *testing.T) {
	svc := newService()
	exercises := []*entities.Exercise{newExercise(entities.UpperBody)}
	cal1, _ := svc.calculateWorkoutParams(exercises, 1.0)
	cal2, _ := svc.calculateWorkoutParams(exercises, 2.0)
	assert.Greater(t, cal2, cal1)
}

// ── applyLevelMultiplier ───────────────────────────────────────────────────

func TestApplyLevelMultiplier(t *testing.T) {
	svc := &Service{}
	light := entities.WorkoutLight
	middle := entities.WorkoutMiddle
	hard := entities.WorkoutHard

	assert.Equal(t, 1.0, svc.applyLevelMultiplier(&light))
	assert.Equal(t, 1.5, svc.applyLevelMultiplier(&middle))
	assert.Equal(t, 2.0, svc.applyLevelMultiplier(&hard))
	assert.Equal(t, 1.0, svc.applyLevelMultiplier(nil))
}

// ── determineWorkoutDisplayLevel ──────────────────────────────────────────

func TestDetermineWorkoutDisplayLevel(t *testing.T) {
	svc := &Service{}
	middle := entities.WorkoutMiddle
	assert.Equal(t, entities.WorkoutMiddle, svc.determineWorkoutDisplayLevel(nil, &middle))
	assert.Equal(t, entities.WorkoutLight, svc.determineWorkoutDisplayLevel(nil, nil))
}

// ── determineExerciseLevel ─────────────────────────────────────────────────

func TestDetermineExerciseLevel_WithParams(t *testing.T) {
	svc := &Service{}
	userID := uuid.New()
	params := newUserParams(userID)
	level := svc.determineExerciseLevel(&dto.GenerateWorkoutParams{UserParams: params})
	assert.Equal(t, entities.Medium, level)
}

func TestDetermineExerciseLevel_NilParams(t *testing.T) {
	svc := &Service{}
	level := svc.determineExerciseLevel(&dto.GenerateWorkoutParams{})
	assert.Equal(t, entities.Medium, level)
}

// ── getInitialLevel ────────────────────────────────────────────────────────

func TestGetInitialLevel_Active(t *testing.T) {
	svc := &Service{}
	userID := uuid.New()
	params := newUserParams(userID)
	level := svc.getInitialLevel(params)
	assert.Equal(t, entities.WorkoutsLevel(entities.Medium), level)
}

func TestGetInitialLevel_Nil(t *testing.T) {
	svc := &Service{}
	assert.Equal(t, entities.WorkoutLight, svc.getInitialLevel(nil))
}

// ── analyzeLevelSuccess / findBestLevel ────────────────────────────────────

func TestAnalyzeLevelSuccess(t *testing.T) {
	svc := &Service{}
	now := time.Now()
	workouts := []*entities.Workout{
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutMiddle, now),
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutMiddle, now),
		newWorkout(uuid.New(), entities.WorkoutStatusFailed, entities.WorkoutLight, now),
	}
	stats := svc.analyzeLevelSuccess(workouts)
	assert.Equal(t, 2, stats[entities.WorkoutMiddle].total)
	assert.Equal(t, 2, stats[entities.WorkoutMiddle].success)
	assert.Equal(t, 1, stats[entities.WorkoutLight].total)
	assert.Equal(t, 0, stats[entities.WorkoutLight].success)
}

func TestFindBestLevel(t *testing.T) {
	svc := &Service{}
	stats := map[entities.WorkoutsLevel]struct{ total, success int }{
		entities.WorkoutLight:  {total: 3, success: 1},
		entities.WorkoutMiddle: {total: 4, success: 4},
	}
	best, ratio := svc.findBestLevel(stats)
	assert.Equal(t, entities.WorkoutMiddle, best)
	assert.Equal(t, 1.0, ratio)
}

func TestFindBestLevel_Empty(t *testing.T) {
	svc := &Service{}
	best, _ := svc.findBestLevel(map[entities.WorkoutsLevel]struct{ total, success int }{})
	assert.Equal(t, entities.WorkoutsLevel(""), best)
}

// ── determinePreferredLevel ────────────────────────────────────────────────

func TestDeterminePreferredLevel_NoWorkouts(t *testing.T) {
	svc := &Service{}
	userID := uuid.New()
	params := newUserParams(userID)
	level := svc.determinePreferredLevel(nil, params, nil)
	assert.Equal(t, entities.WorkoutsLevel(entities.Medium), level)
}

func TestDeterminePreferredLevel_WithWorkouts(t *testing.T) {
	svc := &Service{}
	now := time.Now()
	workouts := []*entities.Workout{
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutHard, now),
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutHard, now),
	}
	level := svc.determinePreferredLevel(nil, nil, workouts)
	assert.Equal(t, entities.WorkoutHard, level)
}

// ── selectBalancedExercises / selectRandomExercises ─────────────────────────

func TestSelectBalancedExercises(t *testing.T) {
	svc := newService()
	preferred := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.UpperBody),
		newExercise(entities.UpperBody),
	}
	other := []*entities.Exercise{
		newExercise(entities.Cardio),
		newExercise(entities.Cardio),
	}
	selected := svc.selectBalancedExercises(preferred, other, 4)
	assert.Len(t, selected, 4)
}

func TestSelectRandomExercises_LessThanCount(t *testing.T) {
	svc := newService()
	exercises := []*entities.Exercise{newExercise(entities.UpperBody)}
	result := svc.selectRandomExercises(exercises, 5)
	assert.Len(t, result, 1)
}

func TestSelectRandomExercises_MoreThanCount(t *testing.T) {
	svc := newService()
	exercises := make([]*entities.Exercise, 10)
	for i := range exercises {
		exercises[i] = newExercise(entities.UpperBody)
	}
	result := svc.selectRandomExercises(exercises, 4)
	assert.Len(t, result, 4)
}

// ── shuffleExercises ──────────────────────────────────────────────────────

func TestShuffleExercises_PreservesLength(t *testing.T) {
	svc := newService()
	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
		newExercise(entities.Flexibility),
	}
	result := svc.shuffleExercises(exercises)
	assert.Len(t, result, 3)
}

// ── selectExercisesForWorkout ──────────────────────────────────────────────

func TestSelectExercisesForWorkout_FewExercises(t *testing.T) {
	svc := newService()
	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
	}
	stats := &dto.AnalyzeWorkoutStats{PopularExerciseType: entities.UpperBody}
	result := svc.selectExercisesForWorkout(exercises, stats)
	assert.Len(t, result, 2)
}

func TestSelectExercisesForWorkout_ManyExercises(t *testing.T) {
	svc := newService()
	exercises := make([]*entities.Exercise, 20)
	for i := range exercises {
		if i%2 == 0 {
			exercises[i] = newExercise(entities.UpperBody)
		} else {
			exercises[i] = newExercise(entities.Cardio)
		}
	}
	stats := &dto.AnalyzeWorkoutStats{PopularExerciseType: entities.UpperBody}
	result := svc.selectExercisesForWorkout(exercises, stats)
	assert.GreaterOrEqual(t, len(result), svc.minExercisesPerWorkout)
	assert.LessOrEqual(t, len(result), svc.maxExercisesPerWorkout)
}

// ── selectCustomExercises ──────────────────────────────────────────────────

func TestSelectCustomExercises_ExactCount(t *testing.T) {
	svc := newService()
	exercises := make([]*entities.Exercise, 20)
	for i := range exercises {
		exercises[i] = newExercise(entities.UpperBody)
	}
	count := 6
	params := &dto.GenerateWorkoutParams{ExercisesCount: &count}
	result := svc.selectCustomExercises(exercises, params)
	assert.Len(t, result, 6)
}

func TestSelectCustomExercises_ClampedToMax(t *testing.T) {
	svc := newService()
	exercises := make([]*entities.Exercise, 30)
	for i := range exercises {
		exercises[i] = newExercise(entities.UpperBody)
	}
	count := 100
	params := &dto.GenerateWorkoutParams{ExercisesCount: &count}
	result := svc.selectCustomExercises(exercises, params)
	assert.LessOrEqual(t, len(result), svc.maxExercisesPerWorkout)
}

// ── prepareWorkoutExercises ────────────────────────────────────────────────

func TestPrepareWorkoutExercises(t *testing.T) {
	svc := newService()
	workoutID := uuid.New()
	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
	}
	result := svc.prepareWorkoutExercises(workoutID, exercises, nil)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].OrderIndex())
	assert.Equal(t, 2, result[1].OrderIndex())
}

// ── getUserParams ──────────────────────────────────────────────────────────

func TestGetUserParams_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newUserParams(userID)

	repo := &mockUserParamsRepo{}
	repo.On("List", mock.Anything, dto.UserParamsFilter{UserID: &userID}, false).
		Return([]*entities.UserParams{params}, nil)

	svc := &Service{userParamsRepository: repo}
	got, err := svc.getUserParams(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, params, got)
}

func TestGetUserParams_Empty(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	repo := &mockUserParamsRepo{}
	repo.On("List", mock.Anything, dto.UserParamsFilter{UserID: &userID}, false).
		Return([]*entities.UserParams{}, nil)

	svc := &Service{userParamsRepository: repo}
	_, err := svc.getUserParams(ctx, userID)
	assert.Error(t, err)
}

func TestGetUserParams_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	repo := &mockUserParamsRepo{}
	repo.On("List", mock.Anything, mock.Anything, false).Return(nil, errors.New("db error"))

	svc := &Service{userParamsRepository: repo}
	_, err := svc.getUserParams(ctx, userID)
	assert.Error(t, err)
}

// ── getUserInfo ────────────────────────────────────────────────────────────

func TestGetUserInfo_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	info := newUserInfo(userID)

	repo := &mockUserInfoRepo{}
	repo.On("Get", mock.Anything, dto.UserInfoFilter{ID: &userID}, false).Return(info, nil)

	svc := &Service{userInfoRepository: repo}
	got, err := svc.getUserInfo(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, info, got)
}

func TestGetUserInfo_Error(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	repo := &mockUserInfoRepo{}
	repo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("not found"))

	svc := &Service{userInfoRepository: repo}
	_, err := svc.getUserInfo(ctx, userID)
	assert.Error(t, err)
}

// ── getExercisesForWorkout ─────────────────────────────────────────────────

func TestGetExercisesForWorkout_Found(t *testing.T) {
	ctx := context.Background()
	place := entities.Gym
	lvl := "medium"
	exercises := []*entities.Exercise{newExercise(entities.UpperBody)}

	repo := &mockExerciseRepo{}
	repo.On("List", mock.Anything, mock.Anything, false).Return(exercises, nil)

	svc := &Service{exerciseRepository: repo}
	got, err := svc.getExercisesForWorkout(ctx, lvl, place)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestGetExercisesForWorkout_FallbackNoPlace(t *testing.T) {
	ctx := context.Background()
	place := entities.Gym
	lvl := "medium"
	exercises := []*entities.Exercise{newExercise(entities.UpperBody)}

	repo := &mockExerciseRepo{}
	// First call (with place) → empty; second call (without place) → exercises.
	repo.On("List", mock.Anything, mock.MatchedBy(func(f dto.ExerciseFilter) bool {
		return f.PlaceExercise != nil
	}), false).Return([]*entities.Exercise{}, nil)
	repo.On("List", mock.Anything, mock.MatchedBy(func(f dto.ExerciseFilter) bool {
		return f.PlaceExercise == nil
	}), false).Return(exercises, nil)

	svc := &Service{exerciseRepository: repo}
	got, err := svc.getExercisesForWorkout(ctx, lvl, place)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
}

// ── saveWorkout ────────────────────────────────────────────────────────────

func TestSaveWorkout_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
	}

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Workout")).Return(nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListExerciseProgress", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	weRepo.On("CreateBulk", mock.Anything, mock.Anything).Return(nil)

	svc := &Service{
		transactionManager:        &mockTxManager{},
		workoutsRepository:        workoutsRepo,
		workoutExerciseRepository: weRepo,
	}

	workout, err := svc.saveWorkout(ctx, userID, exercises, 300, 3600, "medium")
	assert.NoError(t, err)
	assert.NotNil(t, workout)
	workoutsRepo.AssertExpectations(t)
	weRepo.AssertExpectations(t)
}

func TestSaveWorkout_CreateError(t *testing.T) {
	ctx := context.Background()

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListExerciseProgress", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	svc := &Service{
		transactionManager:        &mockTxManager{},
		workoutsRepository:        workoutsRepo,
		workoutExerciseRepository: weRepo,
	}

	_, err := svc.saveWorkout(ctx, uuid.New(), []*entities.Exercise{newExercise(entities.UpperBody)}, 0, 0, "medium")
	assert.Error(t, err)
}

// ── analyzeWorkoutStats ────────────────────────────────────────────────────

func TestAnalyzeWorkoutStats_NoWorkouts(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	info := newUserInfo(userID)
	params := newUserParams(userID)
	now := time.Now()

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("TopListWithLimit", mock.Anything, mock.Anything, defaultWorkoutsLimit, false).
		Return([]*entities.Workout{}, nil)

	svc := &Service{
		workoutsRepository:    workoutsRepo,
		limitGenerateWorkouts: 3,
		location:              time.UTC,
	}

	stats, err := svc.analyzeWorkoutStats(ctx, info, params, now)
	assert.NoError(t, err)
	assert.False(t, stats.SkipGeneration)
	assert.Equal(t, 0, stats.TotalWorkouts)
	assert.Equal(t, 2000, stats.TargetCalories)
}

func TestAnalyzeWorkoutStats_SkipsIfActiveWorkout(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	info := newUserInfo(userID)
	params := newUserParams(userID)
	now := time.Now()

	active := newWorkout(userID, entities.WorkoutStatusInActive, entities.WorkoutLight, now.Add(-1*time.Hour))

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("TopListWithLimit", mock.Anything, mock.Anything, defaultWorkoutsLimit, false).
		Return([]*entities.Workout{active}, nil)

	svc := &Service{
		workoutsRepository:    workoutsRepo,
		limitGenerateWorkouts: 3,
		location:              time.UTC,
	}

	stats, err := svc.analyzeWorkoutStats(ctx, info, params, now)
	assert.NoError(t, err)
	assert.True(t, stats.SkipGeneration)
}

func TestAnalyzeWorkoutStats_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	info := newUserInfo(userID)
	params := newUserParams(userID)

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("TopListWithLimit", mock.Anything, mock.Anything, defaultWorkoutsLimit, false).
		Return(nil, errors.New("db error"))

	svc := &Service{
		workoutsRepository: workoutsRepo,
		location:           time.UTC,
	}

	_, err := svc.analyzeWorkoutStats(ctx, info, params, time.Now())
	assert.Error(t, err)
}

// ── getExercisesByParams ───────────────────────────────────────────────────

func TestGetExercisesByParams_Found(t *testing.T) {
	ctx := context.Background()
	exercises := []*entities.Exercise{newExercise(entities.UpperBody)}

	repo := &mockExerciseRepo{}
	repo.On("List", mock.Anything, mock.Anything, false).Return(exercises, nil)

	svc := &Service{exerciseRepository: repo}
	got, err := svc.getExercisesByParams(ctx, entities.Medium, &dto.GenerateWorkoutParams{})
	assert.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestGetExercisesByParams_FallbackWithoutPlace(t *testing.T) {
	ctx := context.Background()
	place := entities.Gym
	exercises := []*entities.Exercise{newExercise(entities.UpperBody)}

	repo := &mockExerciseRepo{}
	repo.On("List", mock.Anything, mock.MatchedBy(func(f dto.ExerciseFilter) bool {
		return f.PlaceExercise != nil
	}), false).Return([]*entities.Exercise{}, nil)
	repo.On("List", mock.Anything, mock.MatchedBy(func(f dto.ExerciseFilter) bool {
		return f.PlaceExercise == nil
	}), false).Return(exercises, nil)

	svc := &Service{exerciseRepository: repo}
	got, err := svc.getExercisesByParams(ctx, entities.Medium, &dto.GenerateWorkoutParams{PlaceExercise: &place})
	assert.NoError(t, err)
	assert.Len(t, got, 1)
}

// ── calculateWorkoutParamsWithCoef ─────────────────────────────────────────

func TestCalculateWorkoutParamsWithCoef(t *testing.T) {
	svc := newService()
	exercises := []*entities.Exercise{newExercise(entities.UpperBody), newExercise(entities.Cardio)}
	cal, dur := svc.calculateWorkoutParamsWithCoef(exercises, 1.5)
	assert.Greater(t, cal, 0)
	assert.Greater(t, dur, 0)
}

// ── additional mocks ───────────────────────────────────────────────────────

type mockTasksRepo struct{ mock.Mock }

func (m *mockTasksRepo) Create(ctx context.Context, task *entities.Task) error {
	return m.Called(ctx, task).Error(0)
}

type mockUserDevicesRepo struct{ mock.Mock }

func (m *mockUserDevicesRepo) List(ctx context.Context, f dto.UserDeviceFilter) ([]*entities.UserDevice, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserDevice), args.Error(1)
}

// ── selectBalancedExercisesByType ──────────────────────────────────────────

func TestSelectBalancedExercisesByType_Mixed(t *testing.T) {
	svc := newService()
	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
		newExercise(entities.Cardio),
		newExercise(entities.Flexibility),
	}
	result := svc.selectBalancedExercisesByType(exercises, 3)
	assert.Len(t, result, 3)
}

func TestSelectBalancedExercisesByType_Empty(t *testing.T) {
	svc := newService()
	result := svc.selectBalancedExercisesByType([]*entities.Exercise{}, 4)
	assert.Empty(t, result)
}

func TestSelectBalancedExercisesByType_FillsRemainder(t *testing.T) {
	svc := newService()
	// 10 upper, 2 cardio → asking for 8
	exercises := make([]*entities.Exercise, 0, 12)
	for i := 0; i < 10; i++ {
		exercises = append(exercises, newExercise(entities.UpperBody))
	}
	exercises = append(exercises, newExercise(entities.Cardio), newExercise(entities.Cardio))
	result := svc.selectBalancedExercisesByType(exercises, 8)
	assert.LessOrEqual(t, len(result), 8)
	assert.Greater(t, len(result), 0)
}

// ── analyzeUserPreferences ─────────────────────────────────────────────────

func TestAnalyzeUserPreferences_NoWorkouts(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	exType, place := svc.analyzeUserPreferences(ctx, uuid.New(), nil)
	assert.Equal(t, entities.UpperBody, exType)
	assert.Equal(t, entities.Home, place)
}

func TestAnalyzeUserPreferences_WithWorkouts(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()
	now := time.Now()
	workouts := []*entities.Workout{
		newWorkout(uuid.New(), entities.WorkoutStatusDone, entities.WorkoutLight, now),
	}
	exType, place := svc.analyzeUserPreferences(ctx, uuid.New(), workouts)
	assert.Equal(t, entities.UpperBody, exType)
	assert.Equal(t, entities.Home, place)
}

// ── generateWorkout ────────────────────────────────────────────────────────

func newFullService(
	exerciseRepo *mockExerciseRepo,
	workoutsRepo *mockWorkoutsRepo,
	weRepo *mockWorkoutExerciseRepo,
) *Service {
	return &Service{
		transactionManager:        &mockTxManager{},
		exerciseRepository:        exerciseRepo,
		workoutsRepository:        workoutsRepo,
		workoutExerciseRepository: weRepo,
		minExercisesPerWorkout:    2,
		maxExercisesPerWorkout:    6,
		rng:                       newRNG(),
		location:                  time.UTC,
		limitGenerateWorkouts:     3,
		metrics:                   &Metrics{},
		log:                       logging.GetLoggerFromContext(context.Background()),
	}
}

func TestGenerateWorkout_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newUserParams(userID)

	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
		newExercise(entities.Flexibility),
	}
	stats := &dto.AnalyzeWorkoutStats{
		IDUser:              userID,
		PopularExerciseType: entities.UpperBody,
		PopularPlaceExercise: entities.Gym,
	}

	exerciseRepo := &mockExerciseRepo{}
	exerciseRepo.On("List", mock.Anything, mock.Anything, false).Return(exercises, nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.Anything).
		Return([]dto.SkippedExerciseInfo{}, nil)
	weRepo.On("ListExerciseProgress", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	weRepo.On("CreateBulk", mock.Anything, mock.Anything).Return(nil)

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	svc := newFullService(exerciseRepo, workoutsRepo, weRepo)
	workout, err := svc.generateWorkout(ctx, params, stats)

	assert.NoError(t, err)
	assert.NotNil(t, workout)
}

func TestGenerateWorkout_NoExercises(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newUserParams(userID)
	stats := &dto.AnalyzeWorkoutStats{IDUser: userID}

	exerciseRepo := &mockExerciseRepo{}
	exerciseRepo.On("List", mock.Anything, mock.Anything, false).Return([]*entities.Exercise{}, nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.Anything).
		Return([]dto.SkippedExerciseInfo{}, nil)

	workoutsRepo := &mockWorkoutsRepo{}

	svc := newFullService(exerciseRepo, workoutsRepo, weRepo)
	_, err := svc.generateWorkout(ctx, params, stats)
	assert.Error(t, err)
}

// ── GenerateCustomWorkout ──────────────────────────────────────────────────

func TestGenerateCustomWorkout_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newUserParams(userID)
	level := entities.WorkoutMiddle

	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
	}

	exerciseRepo := &mockExerciseRepo{}
	exerciseRepo.On("List", mock.Anything, mock.Anything, false).Return(exercises, nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.Anything).
		Return([]dto.SkippedExerciseInfo{}, nil)
	weRepo.On("ListExerciseProgress", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	weRepo.On("CreateBulk", mock.Anything, mock.Anything).Return(nil)

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	svc := newFullService(exerciseRepo, workoutsRepo, weRepo)

	workout, err := svc.GenerateCustomWorkout(ctx, &dto.GenerateWorkoutParams{
		UserID:     userID,
		UserParams: params,
		Level:      &level,
	})
	assert.NoError(t, err)
	assert.NotNil(t, workout)
}

func TestGenerateCustomWorkout_NoExercises(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	exerciseRepo := &mockExerciseRepo{}
	exerciseRepo.On("List", mock.Anything, mock.Anything, false).Return([]*entities.Exercise{}, nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.Anything).
		Return([]dto.SkippedExerciseInfo{}, nil)

	workoutsRepo := &mockWorkoutsRepo{}

	svc := newFullService(exerciseRepo, workoutsRepo, weRepo)

	_, err := svc.GenerateCustomWorkout(ctx, &dto.GenerateWorkoutParams{UserID: userID})
	assert.Error(t, err)
}

// ── GenerateWorkoutForUser ─────────────────────────────────────────────────

func TestGenerateWorkoutForUser_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newUserParams(userID)
	info := newUserInfo(userID)

	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
	}

	paramsRepo := &mockUserParamsRepo{}
	paramsRepo.On("List", mock.Anything, dto.UserParamsFilter{UserID: &userID}, false).
		Return([]*entities.UserParams{params}, nil)

	infoRepo := &mockUserInfoRepo{}
	infoRepo.On("Get", mock.Anything, dto.UserInfoFilter{ID: &userID}, false).Return(info, nil)

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("TopListWithLimit", mock.Anything, mock.Anything, defaultWorkoutsLimit, false).
		Return([]*entities.Workout{}, nil)
	workoutsRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	exerciseRepo := &mockExerciseRepo{}
	exerciseRepo.On("List", mock.Anything, mock.Anything, false).Return(exercises, nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.Anything).
		Return([]dto.SkippedExerciseInfo{}, nil)
	weRepo.On("ListExerciseProgress", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	weRepo.On("CreateBulk", mock.Anything, mock.Anything).Return(nil)

	svc := &Service{
		transactionManager:        &mockTxManager{},
		userParamsRepository:      paramsRepo,
		userInfoRepository:        infoRepo,
		workoutsRepository:        workoutsRepo,
		exerciseRepository:        exerciseRepo,
		workoutExerciseRepository: weRepo,
		minExercisesPerWorkout:    2,
		maxExercisesPerWorkout:    6,
		rng:                       newRNG(),
		location:                  time.UTC,
		limitGenerateWorkouts:     3,
		metrics:                   &Metrics{},
		log:                       logging.GetLoggerFromContext(ctx),
	}

	workout, err := svc.GenerateWorkoutForUser(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, workout)
}

func TestGenerateWorkoutForUser_UserParamsError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	paramsRepo := &mockUserParamsRepo{}
	paramsRepo.On("List", mock.Anything, mock.Anything, false).Return(nil, errors.New("db error"))

	svc := &Service{
		userParamsRepository: paramsRepo,
		location:             time.UTC,
		metrics:              &Metrics{},
	}

	_, err := svc.GenerateWorkoutForUser(ctx, userID)
	assert.Error(t, err)
}

func TestGenerateWorkoutForUser_UserInfoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newUserParams(userID)

	paramsRepo := &mockUserParamsRepo{}
	paramsRepo.On("List", mock.Anything, mock.Anything, false).Return([]*entities.UserParams{params}, nil)

	infoRepo := &mockUserInfoRepo{}
	infoRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("not found"))

	svc := &Service{
		userParamsRepository: paramsRepo,
		userInfoRepository:   infoRepo,
		location:             time.UTC,
		metrics:              &Metrics{},
	}

	_, err := svc.GenerateWorkoutForUser(ctx, userID)
	assert.Error(t, err)
}

// ── createNotificationTask ─────────────────────────────────────────────────

func TestCreateNotificationTask_WithEmailAndPhone(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	workoutID := uuid.New()
	info := newUserInfo(userID)

	infoRepo := &mockUserInfoRepo{}
	infoRepo.On("Get", mock.Anything, mock.Anything, false).Return(info, nil)

	tasksRepo := &mockTasksRepo{}
	tasksRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Task")).Return(nil)

	svc := &Service{
		userInfoRepository:       infoRepo,
		tasksRepository:          tasksRepo,
		maxRetrySendNotification: 3,
		log:                      logging.GetLoggerFromContext(ctx),
	}

	err := svc.createNotificationTask(ctx, workoutID, userID)
	assert.NoError(t, err)
	// Email + Phone = 2 calls
	tasksRepo.AssertNumberOfCalls(t, "Create", 2)
}

func TestCreateNotificationTask_WithPushDevice(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	workoutID := uuid.New()

	infoRepo := &mockUserInfoRepo{}
	infoRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("no info"))

	device := entities.RestoreUserDevice(entities.UserDeviceRestoreSpec{
		ID:          uuid.New(),
		UserID:      userID,
		DeviceToken: "push-token",
		Platform:    "ios",
	})

	devicesRepo := &mockUserDevicesRepo{}
	devicesRepo.On("List", mock.Anything, dto.UserDeviceFilter{UserID: &userID}).
		Return([]*entities.UserDevice{device}, nil)

	tasksRepo := &mockTasksRepo{}
	tasksRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Task")).Return(nil)

	svc := &Service{
		userInfoRepository:       infoRepo,
		tasksRepository:          tasksRepo,
		userDevicesRepository:    devicesRepo,
		maxRetrySendNotification: 3,
		log:                      logging.GetLoggerFromContext(ctx),
	}

	err := svc.createNotificationTask(ctx, workoutID, userID)
	assert.NoError(t, err)
	tasksRepo.AssertNumberOfCalls(t, "Create", 1) // push only
}

// ── generateWorkoutWithRetry ───────────────────────────────────────────────

func TestGenerateWorkoutWithRetry_SuccessOnFirstTry(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newUserParams(userID)
	stats := &dto.AnalyzeWorkoutStats{
		IDUser:              userID,
		PopularExerciseType: entities.UpperBody,
		PopularPlaceExercise: entities.Gym,
	}

	exercises := []*entities.Exercise{
		newExercise(entities.UpperBody),
		newExercise(entities.Cardio),
	}

	exerciseRepo := &mockExerciseRepo{}
	exerciseRepo.On("List", mock.Anything, mock.Anything, false).Return(exercises, nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.Anything).
		Return([]dto.SkippedExerciseInfo{}, nil)
	weRepo.On("ListExerciseProgress", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	weRepo.On("CreateBulk", mock.Anything, mock.Anything).Return(nil)

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	svc := newFullService(exerciseRepo, workoutsRepo, weRepo)

	workout, err := svc.generateWorkoutWithRetry(ctx, params, stats)
	assert.NoError(t, err)
	assert.NotNil(t, workout)
}

func TestGenerateWorkoutWithRetry_AllRetriesFail(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	params := newUserParams(userID)
	stats := &dto.AnalyzeWorkoutStats{IDUser: userID}

	// Return empty exercises → generateWorkout will always fail
	exerciseRepo := &mockExerciseRepo{}
	exerciseRepo.On("List", mock.Anything, mock.Anything, false).Return([]*entities.Exercise{}, nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.Anything).
		Return([]dto.SkippedExerciseInfo{}, nil).Maybe()

	workoutsRepo := &mockWorkoutsRepo{}

	svc := newFullService(exerciseRepo, workoutsRepo, weRepo)

	_, err := svc.generateWorkoutWithRetry(ctx, params, stats)
	assert.Error(t, err)
}

// ── getExercisesByParams fallback without type ─────────────────────────────

func TestGetExercisesByParams_FallbackWithoutType(t *testing.T) {
	ctx := context.Background()
	exType := entities.UpperBody
	exercises := []*entities.Exercise{newExercise(entities.UpperBody)}

	repo := &mockExerciseRepo{}
	// With type → empty
	repo.On("List", mock.Anything, mock.MatchedBy(func(f dto.ExerciseFilter) bool {
		return f.TypeExercise != nil && f.PlaceExercise == nil
	}), false).Return([]*entities.Exercise{}, nil)
	// Without type → found
	repo.On("List", mock.Anything, mock.MatchedBy(func(f dto.ExerciseFilter) bool {
		return f.TypeExercise == nil && f.PlaceExercise == nil
	}), false).Return(exercises, nil)

	svc := &Service{exerciseRepository: repo}
	got, err := svc.getExercisesByParams(ctx, entities.Medium, &dto.GenerateWorkoutParams{TypeExercise: &exType})
	assert.NoError(t, err)
	assert.Len(t, got, 1)
}

// ── saveWorkout empty exercises ────────────────────────────────────────────

func TestSaveWorkout_EmptyExercises(t *testing.T) {
	ctx := context.Background()

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListExerciseProgress", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	svc := &Service{
		transactionManager:        &mockTxManager{},
		workoutsRepository:        workoutsRepo,
		workoutExerciseRepository: weRepo,
	}

	_, err := svc.saveWorkout(ctx, uuid.New(), []*entities.Exercise{}, 0, 0, "medium")
	assert.Error(t, err)
}

// ── applyLevelMultiplier unknown level ─────────────────────────────────────

func TestApplyLevelMultiplier_UnknownLevel(t *testing.T) {
	svc := &Service{}
	unknown := entities.WorkoutsLevel("unknown")
	assert.Equal(t, 1.0, svc.applyLevelMultiplier(&unknown))
}

// ── GenerateCustomWorkout ExercisesCount clamped to min ──────────────────

func TestSelectCustomExercises_ClampedToMin(t *testing.T) {
	svc := newService()
	exercises := make([]*entities.Exercise, 20)
	for i := range exercises {
		exercises[i] = newExercise(entities.UpperBody)
	}
	count := 0 // below min → should be clamped to minExercisesPerWorkout
	params := &dto.GenerateWorkoutParams{ExercisesCount: &count}
	result := svc.selectCustomExercises(exercises, params)
	assert.GreaterOrEqual(t, len(result), svc.minExercisesPerWorkout)
}

// ── buildSkipMap error ─────────────────────────────────────────────────────

func TestBuildSkipMap_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	weRepo := &mockWorkoutExerciseRepo{}
	weRepo.On("ListSkippedExercises", mock.Anything, userID, mock.Anything).
		Return(nil, errors.New("db error"))

	svc := &Service{workoutExerciseRepository: weRepo}
	_, err := svc.buildSkipMap(ctx, userID)
	assert.Error(t, err)
}

// ── GetUserWorkoutStats ────────────────────────────────────────────────────

func TestGetUserWorkoutStats_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	info := newUserInfo(userID)
	params := newUserParams(userID)

	infoRepo := &mockUserInfoRepo{}
	infoRepo.On("Get", mock.Anything, dto.UserInfoFilter{ID: &userID}, false).Return(info, nil)

	paramsRepo := &mockUserParamsRepo{}
	paramsRepo.On("List", mock.Anything, dto.UserParamsFilter{UserID: &userID}, false).
		Return([]*entities.UserParams{params}, nil)

	workoutsRepo := &mockWorkoutsRepo{}
	workoutsRepo.On("TopListWithLimit", mock.Anything, mock.Anything, defaultWorkoutsLimit, false).
		Return([]*entities.Workout{}, nil)

	svc := &Service{
		userInfoRepository:    infoRepo,
		userParamsRepository:  paramsRepo,
		workoutsRepository:    workoutsRepo,
		limitGenerateWorkouts: 3,
		location:              time.UTC,
	}

	stats, err := svc.GetUserWorkoutStats(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestGetUserWorkoutStats_InfoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	infoRepo := &mockUserInfoRepo{}
	infoRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, errors.New("not found"))

	svc := &Service{
		userInfoRepository: infoRepo,
		location:           time.UTC,
	}

	_, err := svc.GetUserWorkoutStats(ctx, userID)
	assert.Error(t, err)
}
