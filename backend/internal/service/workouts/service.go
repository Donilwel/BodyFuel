package workouts

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"context"
	"fmt"
	"math"
	"math/rand"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const (
	moduleFieldName    = "module"
	workoutsModuleName = "workouts"

	locationName = "Europe/Moscow"

	defaultWorkoutsLimit      = 10
	minExercisesPerWorkout    = 4
	maxExercisesPerWorkout    = 12
	restBetweenWorkouts       = 8 * time.Hour
	daysInWeek                = 7
	batchSize                 = 100
	maxConcurrentUsers        = 10
	maxConcurrentDBOperations = 5

	// Коэффициенты для расчета
	preferredExercisesPercent = 0.6 // 60% предпочтительных упражнений
	restBetweenExercises      = 60  // 60 секунд отдыха между упражнениями

	// Таймауты
	dbOperationTimeout = 30 * time.Second
	generateTimeout    = 2 * time.Minute

	// Progressive overload parameters
	progressRepsIncreasePercent = 0.10 // +10% reps per confirmed progression
	progressMaxRepsMultiplier   = 2.0  // never more than 2× base reps
	progressMinRelaxTime        = 30   // seconds minimum rest
	progressRelaxDecreaseRatio  = 0.85 // rest reduction at high reps
	progressLookbackDays        = 30   // days to look back for history
	progressCompletionsRequired = 2    // min completions to trigger overload
)

type (
	UserParamsRepository interface {
		List(ctx context.Context, f dto.UserParamsFilter, withBlock bool) ([]*entities.UserParams, error)
	}

	UserInfoRepository interface {
		Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
		GetBatch(ctx context.Context, userIDs []uuid.UUID) ([]*entities.UserInfo, error)
	}

	UserWeightRepository interface {
		List(ctx context.Context, f dto.UserWeightFilter, withBlock bool) ([]*entities.UserWeight, error)
	}

	WorkoutsRepository interface {
		TopListWithLimit(ctx context.Context, f dto.WorkoutsFilter, limit int, withBlock bool) ([]*entities.Workout, error)
		Create(ctx context.Context, workout *entities.Workout) error
		Get(ctx context.Context, f dto.WorkoutsFilter, withBlock bool) (*entities.Workout, error)
		Update(ctx context.Context, workout *entities.Workout) error
	}

	TasksRepository interface {
		Create(ctx context.Context, task *entities.Task) error
	}

	TransactionManager interface {
		Do(ctx context.Context, f func(ctx context.Context) error) error
	}

	ExerciseRepository interface {
		List(ctx context.Context, f dto.ExerciseFilter, withBlock bool) ([]*entities.Exercise, error)
		Get(ctx context.Context, f dto.ExerciseFilter, withBlock bool) (*entities.Exercise, error)
	}

	WorkoutExerciseRepository interface {
		CreateBulk(ctx context.Context, workoutExercises []entities.WorkoutsExercise) error
		ListSkippedExercises(ctx context.Context, userID uuid.UUID, since time.Time) ([]dto.SkippedExerciseInfo, error)
		ListExerciseProgress(ctx context.Context, userID uuid.UUID, since time.Time) ([]dto.ExerciseProgressInfo, error)
	}

	UserDevicesRepository interface {
		List(ctx context.Context, f dto.UserDeviceFilter) ([]*entities.UserDevice, error)
	}

	UserFoodRepository interface {
		List(ctx context.Context, f dto.UserFoodFilter) ([]*entities.UserFood, error)
	}
)

type Config struct {
	TransactionManager        TransactionManager
	TasksRepository           TasksRepository
	UserParamsRepository      UserParamsRepository
	UserInfoRepository        UserInfoRepository
	UserWeightRepository      UserWeightRepository
	WorkoutsRepository        WorkoutsRepository
	ExerciseRepository        ExerciseRepository
	WorkoutExerciseRepository WorkoutExerciseRepository
	UserDevicesRepository     UserDevicesRepository
	UserFoodRepository        UserFoodRepository

	WorkoutPullUserInterval  time.Duration
	MaxRetrySendNotification int
	LimitGenerateWorkouts    int
	MinExercisesPerWorkout   int
	MaxExercisesPerWorkout   int
	EnableNotifications      bool
	BatchSize                int
	MaxConcurrentUsers       int
}

type Service struct {
	transactionManager        TransactionManager
	userParamsRepository      UserParamsRepository
	userInfoRepository        UserInfoRepository
	userWeightRepository      UserWeightRepository
	workoutsRepository        WorkoutsRepository
	exerciseRepository        ExerciseRepository
	workoutExerciseRepository WorkoutExerciseRepository
	tasksRepository           TasksRepository
	userDevicesRepository     UserDevicesRepository
	userFoodRepository        UserFoodRepository

	workoutPullUserInterval  time.Duration
	limitGenerateWorkouts    int
	minExercisesPerWorkout   int
	maxExercisesPerWorkout   int
	maxRetrySendNotification int
	enableNotifications      bool
	batchSize                int
	maxConcurrentUsers       int

	log    logging.Entry
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	location *time.Location
	rng      *rand.Rand

	metrics *Metrics
}

type Metrics struct {
	mu sync.RWMutex

	GeneratedWorkouts   int64
	FailedGenerations   int64
	SkippedGenerations  int64
	ProcessedUsers      int64
	AverageGenerateTime time.Duration
}

func NewService(cfg *Config) *Service {
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		loc = time.UTC
	}

	// Создаем отдельный источник случайных чисел
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Устанавливаем значения по умолчанию
	if cfg.MinExercisesPerWorkout == 0 {
		cfg.MinExercisesPerWorkout = minExercisesPerWorkout
	}
	if cfg.MaxExercisesPerWorkout == 0 {
		cfg.MaxExercisesPerWorkout = maxExercisesPerWorkout
	}
	if cfg.LimitGenerateWorkouts == 0 {
		cfg.LimitGenerateWorkouts = 3
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = batchSize
	}
	if cfg.MaxConcurrentUsers == 0 {
		cfg.MaxConcurrentUsers = maxConcurrentUsers
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		transactionManager:        cfg.TransactionManager,
		workoutsRepository:        cfg.WorkoutsRepository,
		exerciseRepository:        cfg.ExerciseRepository,
		workoutExerciseRepository: cfg.WorkoutExerciseRepository,
		tasksRepository:           cfg.TasksRepository,
		userParamsRepository:      cfg.UserParamsRepository,
		userInfoRepository:        cfg.UserInfoRepository,
		userWeightRepository:      cfg.UserWeightRepository,
		userDevicesRepository:     cfg.UserDevicesRepository,
		userFoodRepository:        cfg.UserFoodRepository,

		workoutPullUserInterval:  cfg.WorkoutPullUserInterval,
		maxRetrySendNotification: cfg.MaxRetrySendNotification,
		limitGenerateWorkouts:    cfg.LimitGenerateWorkouts,
		minExercisesPerWorkout:   cfg.MinExercisesPerWorkout,
		maxExercisesPerWorkout:   cfg.MaxExercisesPerWorkout,
		enableNotifications:      cfg.EnableNotifications,
		batchSize:                cfg.BatchSize,
		maxConcurrentUsers:       cfg.MaxConcurrentUsers,

		location: loc,
		rng:      rng,
		log:      logging.WithFields(logging.Fields{"module": "workouts"}),
		ctx:      ctx,
		cancel:   cancel,
		metrics:  &Metrics{},
	}
}

func (s *Service) Run() error {
	s.log = logging.GetLoggerFromContext(s.ctx).WithFields(logging.Fields{
		moduleFieldName: workoutsModuleName,
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.handlePanic()
		s.runWorkoutGenerationLoop()
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.handlePanic()
		s.metricsCollectorLoop()
	}()

	s.log.Infof("Started %s service with interval: %v", workoutsModuleName, s.workoutPullUserInterval)
	return nil
}

func (s *Service) Close() error {
	s.cancel()
	s.wg.Wait()
	s.log.Infof("Stopped %s service", workoutsModuleName)
	return nil
}

func (s *Service) handlePanic() {
	if r := recover(); r != nil {
		s.log.Errorf("Recovered in %s service: %v; stack trace: %s",
			workoutsModuleName, r, debug.Stack())

		time.Sleep(5 * time.Second)
		go s.Run()
	}
}

func (s *Service) metricsCollectorLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.logMetrics()
		}
	}
}

func (s *Service) logMetrics() {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	s.log.Infof("Metrics - Generated: %d, Failed: %d, Skipped: %d, Processed: %d, AvgTime: %v",
		s.metrics.GeneratedWorkouts,
		s.metrics.FailedGenerations,
		s.metrics.SkippedGenerations,
		s.metrics.ProcessedUsers,
		s.metrics.AverageGenerateTime,
	)
}

func (s *Service) GenerateWorkoutForUser(ctx context.Context, userID uuid.UUID) (*entities.Workout, error) {
	startTime := time.Now()
	defer func() {
		s.updateMetrics(time.Since(startTime), true)
	}()

	userParams, err := s.getUserParams(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user params: %w", err)
	}

	userInfo, err := s.getUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	stats, err := s.analyzeWorkoutStats(ctx, userInfo, userParams, time.Now().In(s.location))
	if err != nil {
		return nil, fmt.Errorf("analyze workout stats: %w", err)
	}

	workout, err := s.generateWorkoutWithRetry(ctx, userParams, stats)
	if err != nil {
		s.metrics.mu.Lock()
		s.metrics.FailedGenerations++
		s.metrics.mu.Unlock()
		return nil, fmt.Errorf("generate workout: %w", err)
	}

	return workout, nil
}

func (s *Service) GetUserWorkoutStats(ctx context.Context, userID uuid.UUID) (*dto.AnalyzeWorkoutStats, error) {
	ctx, cancel := context.WithTimeout(ctx, dbOperationTimeout)
	defer cancel()

	userInfo, err := s.getUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	userParams, err := s.getUserParams(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user params: %w", err)
	}

	now := time.Now().In(s.location)

	return s.analyzeWorkoutStats(ctx, userInfo, userParams, now)
}

func (s *Service) runWorkoutGenerationLoop() {
	ticker := time.NewTicker(s.workoutPullUserInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.log.Info("Workout service context cancelled, stopping")
			return
		case <-ticker.C:
			if err := s.processAllUsers(); err != nil {
				s.log.Errorf("Failed to process users: %v", err)
			}
		}
	}
}

func (s *Service) processAllUsers() error {
	ctx, cancel := context.WithTimeout(s.ctx, generateTimeout)
	defer cancel()

	offset := 0
	totalProcessed := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		filter := dto.UserParamsFilter{}
		users, err := s.userParamsRepository.List(ctx, filter, false)
		if err != nil {
			return fmt.Errorf("getting user params at offset %d: %w", offset, err)
		}

		if len(users) == 0 {
			s.log.Debugf("No more users to process at offset %d", offset)
			break
		}

		s.log.Infof("Processing batch of %d users (offset: %d)", len(users), offset)

		processed, err := s.processUserBatch(ctx, users)
		if err != nil {
			s.log.Errorf("Error processing user batch: %v", err)
		}
		totalProcessed += processed

		if len(users) < s.batchSize {
			break
		}
		offset += s.batchSize
	}

	s.metrics.mu.Lock()
	s.metrics.ProcessedUsers += int64(totalProcessed)
	s.metrics.mu.Unlock()

	s.log.Infof("Finished processing users, total processed: %d", totalProcessed)
	return nil
}

func (s *Service) processUserBatch(ctx context.Context, users []*entities.UserParams) (int, error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(s.maxConcurrentUsers)

	var processedMu sync.Mutex
	processedCount := 0

	for _, up := range users {
		up := up

		g.Go(func() error {
			userCtx, cancel := context.WithTimeout(ctx, dbOperationTimeout)
			defer cancel()

			err := s.processGenerateWorkout(userCtx, up)
			if err != nil {
				s.log.Errorf("Failed to process user %s: %v", up.UserID(), err)
				return nil
			}

			processedMu.Lock()
			processedCount++
			processedMu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return processedCount, fmt.Errorf("processing users: %w", err)
	}

	return processedCount, nil
}

func (s *Service) processGenerateWorkout(ctx context.Context, up *entities.UserParams) error {
	startTime := time.Now()
	defer s.updateMetrics(time.Since(startTime), false)

	now := time.Now().In(s.location)
	userID := up.UserID()

	userInfo, err := s.getUserInfo(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user info: %w", err)
	}

	stats, err := s.analyzeWorkoutStats(ctx, userInfo, up, now)
	if err != nil {
		return fmt.Errorf("analyze workout stats: %w", err)
	}

	if stats.SkipGeneration {
		s.log.Infof("Skipping workout generation for user %s: %s", userID, stats.SkipReason)
		s.metrics.mu.Lock()
		s.metrics.SkippedGenerations++
		s.metrics.mu.Unlock()
		return nil
	}

	workout, err := s.generateWorkoutWithRetry(ctx, up, stats)
	if err != nil {
		return fmt.Errorf("generate workout: %w", err)
	}

	s.log.Infof("Successfully generated workout %s for user %s with %d calories, %d minutes",
		workout.ID(), userID, workout.PredictionCalories(), workout.Duration())

	s.metrics.mu.Lock()
	s.metrics.GeneratedWorkouts++
	s.metrics.mu.Unlock()

	go s.createNotificationTaskAsync(s.ctx, workout.ID(), userID)

	return nil
}

func (s *Service) updateMetrics(duration time.Duration, isPublicAPI bool) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	if s.metrics.AverageGenerateTime == 0 {
		s.metrics.AverageGenerateTime = duration
	} else {
		s.metrics.AverageGenerateTime = (s.metrics.AverageGenerateTime + duration) / 2
	}
}

func (s *Service) getUserParams(ctx context.Context, userID uuid.UUID) (*entities.UserParams, error) {
	filter := dto.UserParamsFilter{UserID: &userID}
	userParamsList, err := s.userParamsRepository.List(ctx, filter, false)
	if err != nil {
		return nil, fmt.Errorf("get user params: %w", err)
	}

	if len(userParamsList) == 0 {
		return nil, fmt.Errorf("no user params")
	}

	return userParamsList[0], nil
}

func (s *Service) getUserInfo(ctx context.Context, userID uuid.UUID) (*entities.UserInfo, error) {
	filter := dto.UserInfoFilter{ID: &userID}
	userInfo, err := s.userInfoRepository.Get(ctx, filter, false)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	return userInfo, nil
}

func (s *Service) analyzeWorkoutStats(ctx context.Context, userInfo *entities.UserInfo, userParams *entities.UserParams, now time.Time) (*dto.AnalyzeWorkoutStats, error) {
	userID := userInfo.ID()

	workoutsFilter := dto.WorkoutsFilter{UserID: &userID}
	workouts, err := s.workoutsRepository.TopListWithLimit(ctx, workoutsFilter, defaultWorkoutsLimit, false)
	if err != nil {
		return nil, fmt.Errorf("get workouts: %w", err)
	}

	targetWorkoutsPerWeek := s.getTargetWorkoutsPerWeek(userParams)

	popularExerciseType, popularPlaceExercise := s.analyzeUserPreferences(ctx, userID, workouts)

	preferredLevel := s.determinePreferredLevel(userInfo, userParams, workouts)

	// Nutrition context: today's consumed calories vs target.
	todayCalories := s.getTodayCalories(ctx, userID, now)
	targetCalories := 0
	if userParams != nil {
		targetCalories = userParams.TargetCaloriesDaily()
	}

	// Weight progress.
	currentWeight := 0.0
	targetWeight := 0.0
	if userParams != nil {
		currentWeight = userParams.CurrentWeight()
		targetWeight = userParams.TargetWeight()
	}
	// Try to get the most recent logged weight.
	if recentWeight := s.getLatestWeight(ctx, userID); recentWeight > 0 {
		currentWeight = recentWeight
	}

	stats := &dto.AnalyzeWorkoutStats{
		IDUser:                userID,
		PopularExerciseType:   popularExerciseType,
		PopularPlaceExercise:  popularPlaceExercise,
		AWGLevel:              preferredLevel,
		TargetWorkoutsPerWeek: targetWorkoutsPerWeek,
		TotalWorkouts:         len(workouts),
		TodayCalories:         todayCalories,
		TargetCalories:        targetCalories,
		CalorieBalance:        todayCalories - targetCalories,
		CurrentWeight:         currentWeight,
		TargetWeight:          targetWeight,
		WeightDelta:           currentWeight - targetWeight,
	}

	if len(workouts) == 0 {
		return stats, nil
	}

	lastWorkout := workouts[0]
	stats.LastTimeGenerateWorkout = lastWorkout.CreatedAt()

	skipReason := s.shouldSkipGeneration(workouts, lastWorkout, targetWorkoutsPerWeek, now)
	if skipReason != "" {
		stats.SkipGeneration = true
		stats.SkipReason = skipReason
		return stats, nil
	}

	stats.TotalCancelled = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusFailed)
	stats.TotalNew = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusCreated)
	stats.TotalFinished = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusDone)
	stats.TotalFinishedWorkoutsForWeek = s.countFinishedWorkoutsForWeek(workouts, now)
	stats.SkipGeneration = false

	return stats, nil
}

// getTodayCalories sums calories from all food entries for the user today.
func (s *Service) getTodayCalories(ctx context.Context, userID uuid.UUID, now time.Time) int {
	if s.userFoodRepository == nil {
		return 0
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	foods, err := s.userFoodRepository.List(ctx, dto.UserFoodFilter{UserID: &userID, Date: &today})
	if err != nil {
		s.log.Warnf("getTodayCalories: %v", err)
		return 0
	}
	total := 0
	for _, f := range foods {
		total += f.Calories()
	}
	return total
}

// getLatestWeight returns the most recent logged weight for the user, or 0 if unavailable.
func (s *Service) getLatestWeight(ctx context.Context, userID uuid.UUID) float64 {
	if s.userWeightRepository == nil {
		return 0
	}
	weights, err := s.userWeightRepository.List(ctx, dto.UserWeightFilter{UserID: &userID}, false)
	if err != nil || len(weights) == 0 {
		return 0
	}
	// Find the most recent entry.
	latest := weights[0]
	for _, w := range weights[1:] {
		if w.Date().After(latest.Date()) {
			latest = w
		}
	}
	return latest.Weight()
}

func (s *Service) shouldSkipGeneration(workouts []*entities.Workout, lastWorkout *entities.Workout, targetPerWeek int, now time.Time) string {
	if lastWorkout.Status() == entities.WorkoutStatusInActive {
		return "found active workout, need to finish it first"
	}

	unusedCount := 0
	for _, w := range workouts {
		if w.Status() == entities.WorkoutStatusCreated {
			unusedCount++
		}
	}

	if unusedCount >= s.limitGenerateWorkouts {
		return fmt.Sprintf("already have %d unused workouts (max: %d)", unusedCount, s.limitGenerateWorkouts)
	}

	if lastWorkout.UpdatedAt().Add(restBetweenWorkouts).After(now) {
		timeLeft := time.Until(lastWorkout.UpdatedAt().Add(restBetweenWorkouts))
		return fmt.Sprintf("need to rest %.1f more hours", timeLeft.Hours())
	}

	weeklyWorkouts := s.countFinishedWorkoutsForWeek(workouts, now)
	if weeklyWorkouts >= targetPerWeek {
		return fmt.Sprintf("already completed %d workouts this week (target: %d)", weeklyWorkouts, targetPerWeek)
	}

	return ""
}

func (s *Service) getTargetWorkoutsPerWeek(userParams *entities.UserParams) int {
	if userParams != nil {
		return userParams.TargetWorkoutsWeeks()
	}
	return 3
}

func (s *Service) analyzeUserPreferences(ctx context.Context, userID uuid.UUID, workouts []*entities.Workout) (entities.ExerciseType, entities.PlaceExercise) {
	defaultExerciseType := entities.UpperBody
	defaultPlace := entities.Home

	if s.workoutExerciseRepository == nil {
		return defaultExerciseType, defaultPlace
	}
	since := time.Now().AddDate(0, 0, -progressLookbackDays)
	progress, err := s.workoutExerciseRepository.ListExerciseProgress(ctx, userID, since)
	if err != nil || len(progress) == 0 {
		return defaultExerciseType, defaultPlace
	}

	// Count completions per exercise type and place to find the most popular.
	typeCounts := make(map[entities.ExerciseType]int)
	placeCounts := make(map[entities.PlaceExercise]int)

	for _, p := range progress {
		if p.CompletedCount > 0 && p.TypeExercise != "" {
			typeCounts[p.TypeExercise] += p.CompletedCount
		}
		if p.CompletedCount > 0 && p.PlaceExercise != "" {
			placeCounts[p.PlaceExercise] += p.CompletedCount
		}
	}

	bestType := defaultExerciseType
	bestTypeCount := 0
	for t, c := range typeCounts {
		if c > bestTypeCount {
			bestTypeCount = c
			bestType = t
		}
	}

	bestPlace := defaultPlace
	bestPlaceCount := 0
	for p, c := range placeCounts {
		if c > bestPlaceCount {
			bestPlaceCount = c
			bestPlace = p
		}
	}

	return bestType, bestPlace
}

func (s *Service) generateWorkoutWithRetry(ctx context.Context, userParams *entities.UserParams, stats *dto.AnalyzeWorkoutStats) (*entities.Workout, error) {
	var lastErr error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		workout, err := s.generateWorkout(ctx, userParams, stats)
		if err == nil {
			return workout, nil
		}

		lastErr = err
		s.log.Warnf("Retry %d/%d generating workout: %v", i+1, maxRetries, err)

		time.Sleep(time.Duration(100*(1<<i)) * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to generate workout: %w", lastErr)
}

func (s *Service) generateWorkout(ctx context.Context, userParams *entities.UserParams, stats *dto.AnalyzeWorkoutStats) (*entities.Workout, error) {
	coef, err := userParams.Lifestyle().ToCoef()
	if err != nil {
		return nil, fmt.Errorf("failed to generate coef: %w", err)
	}

	// Adjust intensity based on today's nutrition balance.
	coef *= s.nutritionCoefAdjustment(stats.TodayCalories, stats.TargetCalories, userParams.Want())

	userLevel, err := userParams.Lifestyle().ToLevelPreparation()
	if err != nil {
		return nil, fmt.Errorf("parsing lifestyle to levelpreparation: %w", err)
	}

	// Weight-progress-based exercise type preference overrides the historical average.
	preferredType := s.weightProgressTypePreference(stats.WeightDelta, userParams.Want())
	if preferredType != "" {
		stats.PopularExerciseType = preferredType
	}

	exercises, err := s.getExercisesForWorkout(ctx, userLevel.String(), stats.PopularPlaceExercise)
	if err != nil {
		return nil, fmt.Errorf("getting exercises for workout: %w", err)
	}

	if len(exercises) == 0 {
		return nil, fmt.Errorf("no exercises found")
	}

	// Filter out exercises blocked by skip history.
	skipMap, err := s.buildSkipMap(ctx, stats.IDUser)
	if err != nil {
		s.log.Warnf("buildSkipMap: %v (continuing without skip filter)", err)
	}
	exercises = s.filterSkippedExercises(exercises, skipMap)

	if len(exercises) == 0 {
		return nil, fmt.Errorf("no exercises available after skip filtering")
	}

	selectedExercises := s.selectExercisesForWorkout(exercises, stats)

	// Sort by exercise phase: Flexibility → Strength → Cardio.
	selectedExercises = sortExercisesByPhase(selectedExercises)

	totalCalories, totalDuration := s.calculateWorkoutParams(selectedExercises, coef)

	workout, err := s.saveWorkout(ctx, stats.IDUser, selectedExercises, totalCalories, totalDuration, userLevel.String())
	if err != nil {
		return nil, fmt.Errorf("save workout: %w", err)
	}

	return workout, nil
}

func (s *Service) getExercisesForWorkout(ctx context.Context, userLevel string, place entities.PlaceExercise) ([]*entities.Exercise, error) {
	filter := dto.ExerciseFilter{
		LevelPreparation: (*entities.LevelPreparation)(&userLevel),
		PlaceExercise:    &place,
	}

	exercises, err := s.exerciseRepository.List(ctx, filter, false)
	if err != nil {
		return nil, fmt.Errorf("get exercises: %w", err)
	}

	if len(exercises) == 0 {
		filter.PlaceExercise = nil
		exercises, err = s.exerciseRepository.List(ctx, filter, false)
		if err != nil {
			return nil, fmt.Errorf("list exercises: %w", err)
		}
	}

	return exercises, nil
}

func (s *Service) saveWorkout(ctx context.Context, userID uuid.UUID, exercises []*entities.Exercise,
	totalCalories int, totalDuration int, userLevel string) (*entities.Workout, error) {

	progressMap, err := s.buildProgressMap(ctx, userID)
	if err != nil {
		s.log.Warnf("buildProgressMap: %v (continuing without progressive overload)", err)
		progressMap = nil
	}

	var workout *entities.Workout

	err = s.transactionManager.Do(ctx, func(txCtx context.Context) error {
		workout = entities.NewWorkout(entities.WithWorkoutInitSpec(entities.WorkoutInitSpec{
			ID:                 uuid.New(),
			UserID:             userID,
			Level:              entities.WorkoutMiddle,
			Status:             entities.WorkoutStatusCreated,
			PredictionCalories: totalCalories,
			Duration:           int64(totalDuration),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}))

		if err := s.workoutsRepository.Create(txCtx, workout); err != nil {
			return fmt.Errorf("create workout: %w", err)
		}

		workoutExercises := s.prepareWorkoutExercises(workout.ID(), exercises, progressMap)

		if len(workoutExercises) == 0 {
			return fmt.Errorf("no exercises available for workout")
		}

		if err := s.workoutExerciseRepository.CreateBulk(txCtx, workoutExercises); err != nil {
			return fmt.Errorf("create exercises: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	workout.SetExercises(exercises)
	return workout, nil
}

func (s *Service) prepareWorkoutExercises(workoutID uuid.UUID, exercises []*entities.Exercise, progressMap map[uuid.UUID]dto.ExerciseProgressInfo) []entities.WorkoutsExercise {
	result := make([]entities.WorkoutsExercise, 0, len(exercises))

	for i, ex := range exercises {
		reps := ex.BaseCountReps()
		relaxTime := ex.BaseRelaxTime()

		if progressMap != nil {
			if info, ok := progressMap[ex.ID()]; ok {
				reps, relaxTime = s.applyProgressiveOverload(ex, info)
			}
		}

		// Real calorie calculation using entity helper.
		calories := int(math.Round(ex.AvgCaloriesPer() * float64(reps)))
		if calories <= 0 {
			calories = reps * 5 // fallback: ~5 kcal per rep
		}

		workoutExercise := *entities.NewWorkoutsExercise(entities.WithWorkoutsExerciseInitSpec(
			entities.WorkoutsExerciseInitSpec{
				WorkoutID:       workoutID,
				ExerciseID:      ex.ID(),
				ModifyReps:      reps,
				ModifyRelaxTime: relaxTime,
				Calories:        calories,
				Status:          entities.ExerciseStatusPending,
				OrderIndex:      i + 1,
				UpdatedAt:       time.Now(),
				CreatedAt:       time.Now(),
			}))
		result = append(result, workoutExercise)
	}

	return result
}

// applyProgressiveOverload computes the reps and relax time for a workout exercise
// based on the user's recent completion history for that exercise.
//
// Rules:
//   - Need at least progressCompletionsRequired completions to trigger overload.
//   - Reps increase by progressRepsIncreasePercent (10%) per qualified step,
//     capped at progressMaxRepsMultiplier × base reps.
//   - When reps reach ≥75% of cap, reduce rest time by progressRelaxDecreaseRatio,
//     but never below progressMinRelaxTime seconds.
func (s *Service) applyProgressiveOverload(ex *entities.Exercise, info dto.ExerciseProgressInfo) (reps, relaxTime int) {
	baseReps := ex.BaseCountReps()
	baseRelax := ex.BaseRelaxTime()

	if info.CompletedCount < progressCompletionsRequired {
		// Not enough data yet — use base values.
		return baseReps, baseRelax
	}

	// Start from the last used reps (or base if we have no completion data).
	lastReps := info.LastReps
	if lastReps <= 0 {
		lastReps = baseReps
	}

	// Apply +10% increase.
	newReps := int(math.Round(float64(lastReps) * (1 + progressRepsIncreasePercent)))

	// Cap at 2× base.
	maxReps := int(math.Round(float64(baseReps) * progressMaxRepsMultiplier))
	if newReps > maxReps {
		newReps = maxReps
	}

	// Determine relax time.
	lastRelax := info.LastRelaxTime
	if lastRelax <= 0 {
		lastRelax = baseRelax
	}

	newRelax := lastRelax
	// When reps reach ≥75% of the cap, start reducing rest to increase intensity.
	if newReps >= int(math.Round(float64(maxReps)*0.75)) {
		newRelax = int(math.Round(float64(lastRelax) * progressRelaxDecreaseRatio))
		if newRelax < progressMinRelaxTime {
			newRelax = progressMinRelaxTime
		}
	}

	return newReps, newRelax
}

// buildProgressMap fetches 30-day exercise history for the user and returns a lookup map.
func (s *Service) buildProgressMap(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]dto.ExerciseProgressInfo, error) {
	if s.workoutExerciseRepository == nil {
		return nil, nil
	}
	since := time.Now().AddDate(0, 0, -progressLookbackDays)
	progress, err := s.workoutExerciseRepository.ListExerciseProgress(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]dto.ExerciseProgressInfo, len(progress))
	for _, p := range progress {
		m[p.ExerciseID] = p
	}
	return m, nil
}

func (s *Service) selectExercisesForWorkout(exercises []*entities.Exercise, stats *dto.AnalyzeWorkoutStats) []*entities.Exercise {
	exerciseCount := s.rng.Intn(s.maxExercisesPerWorkout-s.minExercisesPerWorkout+1) + s.minExercisesPerWorkout

	if len(exercises) <= exerciseCount {
		return s.shuffleExercises(exercises)
	}

	var preferredExercises, otherExercises []*entities.Exercise
	for _, ex := range exercises {
		if ex.TypeExercise() == stats.PopularExerciseType {
			preferredExercises = append(preferredExercises, ex)
		} else {
			otherExercises = append(otherExercises, ex)
		}
	}

	selected := s.selectBalancedExercises(preferredExercises, otherExercises, exerciseCount)

	return s.shuffleExercises(selected)
}

func (s *Service) selectBalancedExercises(preferred, other []*entities.Exercise, targetCount int) []*entities.Exercise {
	preferredCount := int(float64(targetCount) * preferredExercisesPercent)
	if preferredCount > len(preferred) {
		preferredCount = len(preferred)
	}

	otherCount := targetCount - preferredCount
	if otherCount > len(other) {
		otherCount = len(other)
		preferredCount = targetCount - otherCount
	}

	selected := make([]*entities.Exercise, 0, targetCount)

	if preferredCount > 0 {
		selected = append(selected, s.selectRandomExercises(preferred, preferredCount)...)
	}
	if otherCount > 0 {
		selected = append(selected, s.selectRandomExercises(other, otherCount)...)
	}

	return selected
}

func (s *Service) selectRandomExercises(exercises []*entities.Exercise, count int) []*entities.Exercise {
	if count >= len(exercises) {
		return exercises
	}

	shuffled := s.shuffleExercises(exercises)
	return shuffled[:count]
}

func (s *Service) shuffleExercises(exercises []*entities.Exercise) []*entities.Exercise {
	result := make([]*entities.Exercise, len(exercises))
	copy(result, exercises)

	s.rng.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

func (s *Service) calculateWorkoutParams(exercises []*entities.Exercise, coef float64) (int, int) {
	totalCalories := 0.0
	totalDuration := 0

	for _, ex := range exercises {
		totalCalories += ex.CalculateCalories(coef)
		totalDuration += ex.CalculateDuration(coef)
	}

	if len(exercises) > 1 {
		restTime := (len(exercises) - 1) * restBetweenExercises
		totalDuration += restTime
	}

	return int(math.Round(totalCalories)), totalDuration
}

func (s *Service) determinePreferredLevel(userInfo *entities.UserInfo, userParams *entities.UserParams, workouts []*entities.Workout) entities.WorkoutsLevel {
	if len(workouts) == 0 {
		return s.getInitialLevel(userParams)
	}

	levelStats := s.analyzeLevelSuccess(workouts)

	bestLevel, _ := s.findBestLevel(levelStats)
	if bestLevel == "" {
		return entities.WorkoutLight
	}

	return bestLevel
}

func (s *Service) analyzeLevelSuccess(workouts []*entities.Workout) map[entities.WorkoutsLevel]struct{ total, success int } {
	stats := make(map[entities.WorkoutsLevel]struct{ total, success int })

	for _, w := range workouts {
		levelStats := stats[w.Level()]
		levelStats.total++
		if w.Status() == entities.WorkoutStatusDone {
			levelStats.success++
		}
		stats[w.Level()] = levelStats
	}

	return stats
}

func (s *Service) findBestLevel(stats map[entities.WorkoutsLevel]struct{ total, success int }) (entities.WorkoutsLevel, float64) {
	var bestLevel entities.WorkoutsLevel
	bestRatio := 0.0

	for level, stat := range stats {
		if stat.total > 0 {
			ratio := float64(stat.success) / float64(stat.total)
			if ratio > bestRatio {
				bestRatio = ratio
				bestLevel = level
			}
		}
	}

	return bestLevel, bestRatio
}

func (s *Service) getInitialLevel(userParams *entities.UserParams) entities.WorkoutsLevel {
	if userParams != nil && userParams.Lifestyle() != "" {
		level, err := userParams.Lifestyle().ToLevelPreparation()
		if err == nil && level != "" {
			return entities.WorkoutsLevel(level)
		}
	}
	return entities.WorkoutLight
}

func (s *Service) countWorkoutsByStatus(workouts []*entities.Workout, status entities.WorkoutsStatus) int {
	count := 0
	for _, w := range workouts {
		if w.Status() == status {
			count++
		}
	}
	return count
}

func (s *Service) countFinishedWorkoutsForWeek(workouts []*entities.Workout, now time.Time) int {
	count := 0
	weekAgo := now.AddDate(0, 0, -daysInWeek)

	for _, w := range workouts {
		if w.Status() == entities.WorkoutStatusDone &&
			w.CreatedAt().After(weekAgo) &&
			w.CreatedAt().Before(now) {
			count++
		}
	}
	return count
}

func (s *Service) createNotificationTaskAsync(ctx context.Context, workoutID, userID uuid.UUID) {
	defer s.handlePanic()

	if err := s.createNotificationTask(ctx, workoutID, userID); err != nil {
		s.log.Errorf("Failed to create notification task for workout %s: %v", workoutID, err)
	}
}

func (s *Service) createNotificationTask(ctx context.Context, workoutID, userID uuid.UUID) error {
	msgBody := string(entities.TaskMessageSendAuthomaticGeneratedWorkout)

	userInfo, err := s.userInfoRepository.Get(ctx, dto.UserInfoFilter{ID: &userID}, false)
	if err != nil {
		s.log.Warnf("createNotificationTask: get user info: %v", err)
	}

	if userInfo != nil && userInfo.Email() != "" {
		task := entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
			TypeNm:      entities.TaskTypeSendNotificationEmail,
			Message:     entities.TaskMessageSendAuthomaticGeneratedWorkout,
			MaxAttempts: s.maxRetrySendNotification,
			Attribute: entities.TaskAttribute{
				UserID:  userID,
				Email:   userInfo.Email(),
				Subject: "Новая тренировка готова",
				Body:    msgBody,
			},
		}))
		if err := s.tasksRepository.Create(ctx, task); err != nil {
			s.log.Errorf("createNotificationTask: create email task: %v", err)
		}
	}

	if userInfo != nil && userInfo.Phone() != "" {
		task := entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
			TypeNm:      entities.TaskTypeSendNotificationPhone,
			Message:     entities.TaskMessageSendAuthomaticGeneratedWorkout,
			MaxAttempts: s.maxRetrySendNotification,
			Attribute: entities.TaskAttribute{
				UserID: userID,
				Phone:  userInfo.Phone(),
				Body:   msgBody,
			},
		}))
		if err := s.tasksRepository.Create(ctx, task); err != nil {
			s.log.Errorf("createNotificationTask: create sms task: %v", err)
		}
	}

	if s.userDevicesRepository != nil {
		devices, err := s.userDevicesRepository.List(ctx, dto.UserDeviceFilter{UserID: &userID})
		if err != nil {
			s.log.Warnf("createNotificationTask: get user devices: %v", err)
		}
		for _, device := range devices {
			task := entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
				TypeNm:      entities.TaskTypeSendPushNotification,
				Message:     entities.TaskMessageSendAuthomaticGeneratedWorkout,
				MaxAttempts: s.maxRetrySendNotification,
				Attribute: entities.TaskAttribute{
					UserID:      userID,
					DeviceToken: device.DeviceToken(),
					Title:       "Новая тренировка готова",
					Body:        msgBody,
				},
			}))
			if err := s.tasksRepository.Create(ctx, task); err != nil {
				s.log.Errorf("createNotificationTask: create push task: %v", err)
			}
		}
	}

	return nil
}

func (s *Service) GenerateCustomWorkout(ctx context.Context, params *dto.GenerateWorkoutParams) (*entities.Workout, error) {
	startTime := time.Now()
	defer s.updateMetrics(time.Since(startTime), true)

	finalCoef := s.applyLevelMultiplier(params.Level)

	// Adjust intensity based on today's nutrition if user params available.
	if params.UserParams != nil {
		todayCalories := s.getTodayCalories(ctx, params.UserID, time.Now().In(s.location))
		targetCalories := params.UserParams.TargetCaloriesDaily()
		finalCoef *= s.nutritionCoefAdjustment(todayCalories, targetCalories, params.UserParams.Want())
	}

	levelPreparation := s.determineExerciseLevel(params)

	exercises, err := s.getExercisesByParams(ctx, levelPreparation, params)
	if err != nil {
		return nil, fmt.Errorf("get exercises by params: %w", err)
	}

	if len(exercises) == 0 {
		return nil, fmt.Errorf("no exercises found for given parameters")
	}

	// Apply skip filtering.
	skipMap, err := s.buildSkipMap(ctx, params.UserID)
	if err != nil {
		s.log.Warnf("GenerateCustomWorkout buildSkipMap: %v (continuing without skip filter)", err)
	}
	exercises = s.filterSkippedExercises(exercises, skipMap)

	if len(exercises) == 0 {
		return nil, fmt.Errorf("no exercises available after skip filtering")
	}

	selectedExercises := s.selectCustomExercises(exercises, params)

	// Sort exercises by phase: Flexibility → Strength → Cardio.
	selectedExercises = sortExercisesByPhase(selectedExercises)

	totalCalories, totalDuration := s.calculateWorkoutParamsWithCoef(selectedExercises, finalCoef)

	workoutLevel := s.determineWorkoutDisplayLevel(selectedExercises, params.Level)

	workout, err := s.saveWorkout(ctx, params.UserID, selectedExercises, totalCalories, totalDuration, workoutLevel.String())
	if err != nil {
		return nil, fmt.Errorf("save workout: %w", err)
	}

	s.log.Infof("Generated custom workout %s for user %s with %d exercises, %d calories, %d minutes (coef: %.2f, level: %s)",
		workout.ID(), params.UserID, len(selectedExercises), totalCalories, totalDuration/60, finalCoef, workoutLevel)

	return workout, nil
}

func (s *Service) applyLevelMultiplier(level *entities.WorkoutsLevel) float64 {
	baseCoef := 1.0
	if level == nil {
		return baseCoef
	}

	switch *level {
	case entities.WorkoutLight:
		return baseCoef * 1
	case entities.WorkoutMiddle:
		return baseCoef * 1.5
	case entities.WorkoutHard:
		return baseCoef * 2.0
	default:
		return baseCoef
	}
}

func (s *Service) calculateWorkoutParamsWithCoef(exercises []*entities.Exercise, coef float64) (int, int) {
	totalCalories := 0.0
	totalDuration := 0

	for _, ex := range exercises {
		totalCalories += ex.CalculateCalories(coef)
		totalDuration += ex.CalculateDuration(coef)
	}

	if len(exercises) > 1 {
		restTime := (len(exercises) - 1) * restBetweenExercises
		totalDuration += restTime
	}

	return int(math.Round(totalCalories)), totalDuration
}

func (s *Service) determineExerciseLevel(params *dto.GenerateWorkoutParams) entities.LevelPreparation {
	if params.UserParams != nil && params.UserParams.Lifestyle() != "" {
		level, err := params.UserParams.Lifestyle().ToLevelPreparation()
		if err == nil {
			return level
		}
	}
	return entities.Medium
}

func (s *Service) determineWorkoutDisplayLevel(exercises []*entities.Exercise, requestedLevel *entities.WorkoutsLevel) entities.WorkoutsLevel {
	if requestedLevel != nil {
		return *requestedLevel
	}
	return entities.WorkoutLight
}

func (s *Service) getExercisesByParams(ctx context.Context, level entities.LevelPreparation, params *dto.GenerateWorkoutParams) ([]*entities.Exercise, error) {
	filter := dto.ExerciseFilter{
		LevelPreparation: &level,
	}
	if params.PlaceExercise != nil {
		filter.PlaceExercise = params.PlaceExercise
	}

	if params.TypeExercise != nil {
		filter.TypeExercise = params.TypeExercise
	}

	exercises, err := s.exerciseRepository.List(ctx, filter, false)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}

	if len(exercises) == 0 && params.PlaceExercise != nil {
		filter.PlaceExercise = nil
		exercises, err = s.exerciseRepository.List(ctx, filter, false)
		if err != nil {
			return nil, fmt.Errorf("list exercises without place: %w", err)
		}
	}

	if len(exercises) == 0 && params.TypeExercise != nil {
		filter.TypeExercise = nil
		exercises, err = s.exerciseRepository.List(ctx, filter, false)
		if err != nil {
			return nil, fmt.Errorf("list exercises without type: %w", err)
		}
	}

	return exercises, nil
}

func (s *Service) selectCustomExercises(exercises []*entities.Exercise, params *dto.GenerateWorkoutParams) []*entities.Exercise {
	exerciseCount := s.minExercisesPerWorkout
	if params.ExercisesCount != nil {
		exerciseCount = *params.ExercisesCount
		if exerciseCount > s.maxExercisesPerWorkout {
			exerciseCount = s.maxExercisesPerWorkout
		}
		if exerciseCount < s.minExercisesPerWorkout {
			exerciseCount = s.minExercisesPerWorkout
		}
	} else {
		exerciseCount = s.rng.Intn(s.maxExercisesPerWorkout-s.minExercisesPerWorkout+1) + s.minExercisesPerWorkout
	}

	if len(exercises) <= exerciseCount {
		return s.shuffleExercises(exercises)
	}

	// Если указан конкретный тип упражнения, выбираем только их
	if params.TypeExercise != nil {
		return s.selectRandomExercises(exercises, exerciseCount)
	}

	// Иначе пытаемся сбалансировать по типам
	return s.selectBalancedExercisesByType(exercises, exerciseCount)
}

// ── Skip-tracking helpers ──────────────────────────────────────────────────

// buildSkipMap fetches skip data for the past 7 days and builds a lookup map keyed by exercise ID.
func (s *Service) buildSkipMap(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]dto.SkippedExerciseInfo, error) {
	since := time.Now().Add(-7 * 24 * time.Hour)
	skipped, err := s.workoutExerciseRepository.ListSkippedExercises(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]dto.SkippedExerciseInfo, len(skipped))
	for _, info := range skipped {
		m[info.ExerciseID] = info
	}
	return m, nil
}

// filterSkippedExercises removes exercises that have been skipped twice within the last 7 days.
// Exercises skipped only once remain available (they get one more chance).
func (s *Service) filterSkippedExercises(exercises []*entities.Exercise, skipMap map[uuid.UUID]dto.SkippedExerciseInfo) []*entities.Exercise {
	if len(skipMap) == 0 {
		return exercises
	}
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	result := make([]*entities.Exercise, 0, len(exercises))
	for _, ex := range exercises {
		info, found := skipMap[ex.ID()]
		if !found {
			result = append(result, ex)
			continue
		}
		// Blocked: skipped ≥2 times and last skip is within the past week.
		if info.SkipCount >= 2 && info.LastSkippedAt.After(weekAgo) {
			continue
		}
		result = append(result, ex)
	}
	return result
}

// ── Phase-ordering helpers ─────────────────────────────────────────────────

// exercisePhase assigns an ordering phase to an exercise type:
//
//	0 = Flexibility (warm-up / cooldown – goes first)
//	1 = Strength (upper/lower/full body)
//	2 = Cardio (goes last)
func exercisePhase(t entities.ExerciseType) int {
	switch t {
	case entities.Flexibility:
		return 0
	case entities.UpperBody, entities.LowerBody, entities.FullBody:
		return 1
	case entities.Cardio:
		return 2
	default:
		return 1
	}
}

// sortExercisesByPhase returns a copy of exercises sorted by phase, preserving relative order
// within each phase (stable sort).
func sortExercisesByPhase(exercises []*entities.Exercise) []*entities.Exercise {
	sorted := make([]*entities.Exercise, len(exercises))
	copy(sorted, exercises)
	sort.SliceStable(sorted, func(i, j int) bool {
		return exercisePhase(sorted[i].TypeExercise()) < exercisePhase(sorted[j].TypeExercise())
	})
	return sorted
}

// ── Nutrition-aware intensity helpers ─────────────────────────────────────

// nutritionCoefAdjustment returns a multiplier (0.8–1.2) based on today's calorie balance.
//
//   - Surplus > 300 kcal  → push harder (+20 %)
//   - Deficit > 300 kcal  → ease off (−20 %) – protect muscle when under-fuelled
//   - Otherwise           → no adjustment
func (s *Service) nutritionCoefAdjustment(consumed, target int, goal entities.Want) float64 {
	if target <= 0 {
		return 1.0
	}
	balance := consumed - target
	switch {
	case balance > 300:
		return 1.2
	case balance < -300:
		// For muscle-building goals, don't reduce – keep stimulation even in deficit.
		if goal == entities.BuildMuscle {
			return 1.0
		}
		return 0.8
	default:
		return 1.0
	}
}

// weightProgressTypePreference returns the exercise type that should be emphasised given the
// user's weight delta (current − target) and their stated goal.
//
//   - Need to lose weight (delta > 1 kg) → prefer Cardio or FullBody
//   - Need to gain / build muscle (delta < -1 kg) → prefer strength (UpperBody)
//   - On-target → no preference (empty string)
func (s *Service) weightProgressTypePreference(weightDelta float64, goal entities.Want) entities.ExerciseType {
	const threshold = 1.0
	switch {
	case weightDelta > threshold:
		// More than 1 kg above target: favour cardio/full-body to burn calories.
		if goal == entities.BuildMuscle {
			return entities.FullBody
		}
		return entities.Cardio
	case weightDelta < -threshold:
		return entities.UpperBody
	default:
		return ""
	}
}

func (s *Service) selectBalancedExercisesByType(exercises []*entities.Exercise, targetCount int) []*entities.Exercise {
	byType := make(map[entities.ExerciseType][]*entities.Exercise)
	for _, ex := range exercises {
		byType[ex.TypeExercise()] = append(byType[ex.TypeExercise()], ex)
	}

	availableTypes := make([]entities.ExerciseType, 0, len(byType))
	for t := range byType {
		availableTypes = append(availableTypes, t)
	}

	if len(availableTypes) == 0 {
		return []*entities.Exercise{}
	}

	typesCount := len(availableTypes)
	basePerType := targetCount / typesCount
	remainder := targetCount % typesCount

	selected := make([]*entities.Exercise, 0, targetCount)

	for i, t := range availableTypes {
		count := basePerType
		if i < remainder {
			count++
		}

		typeExercises := byType[t]
		if count > len(typeExercises) {
			count = len(typeExercises)
		}

		if count > 0 {
			selected = append(selected, s.selectRandomExercises(typeExercises, count)...)
		}
	}

	if len(selected) < targetCount {
		remaining := make([]*entities.Exercise, 0)
		for _, ex := range exercises {
			found := false
			for _, s := range selected {
				if s.ID() == ex.ID() {
					found = true
					break
				}
			}
			if !found {
				remaining = append(remaining, ex)
			}
		}

		additional := s.selectRandomExercises(remaining, targetCount-len(selected))
		selected = append(selected, additional...)
	}

	return s.shuffleExercises(selected)
}
