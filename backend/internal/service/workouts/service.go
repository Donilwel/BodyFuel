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
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const (
	moduleFieldName    = "module"
	workoutsModuleName = "workouts"

	locationName = "Europe/Moscow"

	// Константы для генерации тренировок
	defaultWorkoutsLimit      = 10
	minExercisesPerWorkout    = 4
	maxExercisesPerWorkout    = 8
	restBetweenWorkouts       = 8 * time.Hour
	daysInWeek                = 7
	batchSize                 = 100
	maxConcurrentUsers        = 10
	maxConcurrentDBOperations = 5

	// Коэффициенты для расчета
	preferredExercisesPercent  = 0.6 // 60% предпочтительных упражнений
	restBetweenExercises       = 60  // 60 секунд отдыха между упражнениями
	defaultCaloriesPerExercise = 150 // Калорий по умолчанию для упражнения

	// Таймауты
	dbOperationTimeout = 30 * time.Second
	generateTimeout    = 2 * time.Minute
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

// ==================== ПУБЛИЧНЫЕ МЕТОДЫ ====================

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

		// Перезапускаем сервис через некоторое время
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

// GenerateWorkoutForUser генерирует тренировку для конкретного пользователя
func (s *Service) GenerateWorkoutForUser(ctx context.Context, userID uuid.UUID) (*entities.Workout, error) {
	startTime := time.Now()
	defer func() {
		s.updateMetrics(time.Since(startTime), true)
	}()

	// Получаем параметры пользователя
	userParams, err := s.getUserParams(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user params: %w", err)
	}

	// Получаем информацию о пользователе
	userInfo, err := s.getUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	// Анализируем статистику
	stats, err := s.analyzeWorkoutStats(ctx, userInfo, userParams, time.Now().In(s.location))
	if err != nil {
		return nil, fmt.Errorf("analyze workout stats: %w", err)
	}

	// Генерируем тренировку
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
		// Проверяем не отменён ли контекст
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Получаем пачку пользователей
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

		// Обрабатываем пачку пользователей
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
		up := up // захват переменной для замыкания

		g.Go(func() error {
			// Создаём дочерний контекст с таймаутом для каждого пользователя
			userCtx, cancel := context.WithTimeout(ctx, dbOperationTimeout)
			defer cancel()

			err := s.processGenerateWorkout(userCtx, up)
			if err != nil {
				s.log.Errorf("Failed to process user %s: %v", up.UserID(), err)
				return nil // Не возвращаем ошибку, чтобы продолжить обработку других пользователей
			}

			processedMu.Lock()
			processedCount++
			processedMu.Unlock()

			return nil
		})
	}

	// Ждём завершения всех горутин
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

	// Получаем информацию о пользователе
	userInfo, err := s.getUserInfo(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user info: %w", err)
	}

	// Анализируем статистику тренировок
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

	// Генерируем тренировку
	workout, err := s.generateWorkoutWithRetry(ctx, up, stats)
	if err != nil {
		return fmt.Errorf("generate workout: %w", err)
	}

	s.log.Infof("Successfully generated workout %s for user %s with %d calories, %d minutes",
		workout.ID(), userID, workout.PredictionCalories(), workout.Duration())

	s.metrics.mu.Lock()
	s.metrics.GeneratedWorkouts++
	s.metrics.mu.Unlock()

	// Создаём задачу на уведомление асинхронно
	if s.enableNotifications {
		go s.createNotificationTaskAsync(s.ctx, workout.ID(), userID)
	}

	return nil
}

func (s *Service) updateMetrics(duration time.Duration, isPublicAPI bool) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	// Обновляем среднее время генерации
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

	// Получаем последние тренировки пользователя
	workoutsFilter := dto.WorkoutsFilter{UserID: &userID}
	workouts, err := s.workoutsRepository.TopListWithLimit(ctx, workoutsFilter, defaultWorkoutsLimit, false)
	if err != nil {
		return nil, fmt.Errorf("get workouts: %w", err)
	}

	// Получаем целевую недельную нагрузку
	targetWorkoutsPerWeek := s.getTargetWorkoutsPerWeek(userParams)

	// Анализируем предпочтения пользователя
	popularExerciseType, popularPlaceExercise := s.analyzeUserPreferences(ctx, userID, workouts)

	// Определяем предпочтительный уровень
	preferredLevel := s.determinePreferredLevel(userInfo, userParams, workouts)

	stats := &dto.AnalyzeWorkoutStats{
		IDUser:                userID,
		PopularExerciseType:   popularExerciseType,
		PopularPlaceExercise:  popularPlaceExercise,
		AWGLevel:              preferredLevel,
		TargetWorkoutsPerWeek: targetWorkoutsPerWeek,
		TotalWorkouts:         len(workouts),
	}

	// Если тренировок нет - это новый пользователь
	if len(workouts) == 0 {
		return stats, nil
	}

	lastWorkout := workouts[0]
	stats.LastTimeGenerateWorkout = lastWorkout.CreatedAt()

	// Проверяем возможность генерации новой тренировки
	skipReason := s.shouldSkipGeneration(workouts, lastWorkout, targetWorkoutsPerWeek, now)
	if skipReason != "" {
		stats.SkipGeneration = true
		stats.SkipReason = skipReason
		return stats, nil
	}

	// Заполняем полную статистику
	stats.TotalCancelled = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusFailed)
	stats.TotalNew = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusCreated)
	stats.TotalFinished = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusDone)
	stats.TotalFinishedWorkoutsForWeek = s.countFinishedWorkoutsForWeek(workouts, now)
	stats.SkipGeneration = false

	return stats, nil
}

func (s *Service) shouldSkipGeneration(workouts []*entities.Workout, lastWorkout *entities.Workout, targetPerWeek int, now time.Time) string {
	// Проверяем наличие активной тренировки
	if lastWorkout.Status() == entities.WorkoutStatusInActive {
		return "found active workout, need to finish it first"
	}

	// Считаем количество неиспользованных тренировок
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

	// Проверяем, достигнута ли цель по тренировкам за неделю
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
	return 3 // Значение по умолчанию
}

func (s *Service) analyzeUserPreferences(ctx context.Context, userID uuid.UUID, workouts []*entities.Workout) (entities.ExerciseType, entities.PlaceExercise) {
	// По умолчанию
	defaultExerciseType := entities.UpperBody
	defaultPlace := entities.Home

	if len(workouts) == 0 {
		return defaultExerciseType, defaultPlace
	}

	// TODO: Реализовать анализ реальных предпочтений из workout_exercises

	return defaultExerciseType, defaultPlace
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

		// Экспоненциальная задержка между попытками
		time.Sleep(time.Duration(100*(1<<i)) * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to generate workout: %w", lastErr)
}

func (s *Service) generateWorkout(ctx context.Context, userParams *entities.UserParams, stats *dto.AnalyzeWorkoutStats) (*entities.Workout, error) {
	// Получаем коэффициент нагрузки из образа жизни
	coef, err := userParams.Lifestyle().ToCoef()
	if err != nil {
		return nil, fmt.Errorf("failed to generate coef: %w", err)
	}

	userLevel, err := userParams.Lifestyle().ToLevelPreparation()
	if err != nil {
		return nil, fmt.Errorf("parsing lifestyle to levelpreparation: %w", err)
	}

	// Получаем упражнения для данного уровня
	exercises, err := s.getExercisesForWorkout(ctx, userLevel.String(), stats.PopularPlaceExercise)
	if err != nil {
		return nil, fmt.Errorf("getting exercises for workout: %w", err)
	}

	if len(exercises) == 0 {
		return nil, fmt.Errorf("no exercises found")
	}

	// Выбираем упражнения для тренировки
	selectedExercises := s.selectExercisesForWorkout(exercises, stats)

	// Рассчитываем параметры тренировки
	totalCalories, totalDuration := s.calculateWorkoutParams(selectedExercises, coef)

	// Создаём и сохраняем тренировку
	workout, err := s.saveWorkout(ctx, stats.IDUser, selectedExercises, totalCalories, totalDuration, userLevel.String())
	if err != nil {
		return nil, fmt.Errorf("save workout: %w", err)
	}

	return workout, nil
}

func (s *Service) getExercisesForWorkout(ctx context.Context, userLevel string, place entities.PlaceExercise) ([]*entities.Exercise, error) {
	// Пытаемся получить упражнения для предпочитаемого места
	filter := dto.ExerciseFilter{
		LevelPreparation: (*entities.LevelPreparation)(&userLevel),
		PlaceExercise:    &place,
	}

	exercises, err := s.exerciseRepository.List(ctx, filter, false)
	if err != nil {
		return nil, fmt.Errorf("get exercises: %w", err)
	}

	if len(exercises) == 0 {
		// Если нет упражнений для предпочитаемого места, ищем все для данного уровня
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

	var workout *entities.Workout

	err := s.transactionManager.Do(ctx, func(txCtx context.Context) error {
		// Создаём тренировку
		workout = entities.NewWorkout(entities.WithWorkoutInitSpec(entities.WorkoutInitSpec{
			ID:                 uuid.New(),
			UserID:             userID,
			Level:              s.determineWorkoutLevel(exercises, entities.WorkoutsLevel(userLevel)),
			Status:             entities.WorkoutStatusCreated,
			PredictionCalories: totalCalories,
			Duration:           int64(totalDuration),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}))

		if err := s.workoutsRepository.Create(txCtx, workout); err != nil {
			return fmt.Errorf("create workout: %w", err)
		}

		// Создаём упражнения для тренировки
		workoutExercises := s.prepareWorkoutExercises(workout.ID(), exercises)

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

	// Добавляем упражнения в объект тренировки
	workout.SetExercises(exercises)
	return workout, nil
}

func (s *Service) prepareWorkoutExercises(workoutID uuid.UUID, exercises []*entities.Exercise) []entities.WorkoutsExercise {
	result := make([]entities.WorkoutsExercise, 0, len(exercises))

	for i, ex := range exercises {
		workoutExercise := *entities.NewWorkoutsExercise(entities.WithWorkoutsExerciseInitSpec(
			entities.WorkoutsExerciseInitSpec{
				WorkoutID:       workoutID,
				ExerciseID:      ex.ID(),
				ModifyReps:      ex.BaseCountReps(),
				ModifyRelaxTime: ex.BaseRelaxTime(),
				Calories:        defaultCaloriesPerExercise, // TODO: Рассчитывать реальные калории
				Status:          entities.ExerciseStatusPending,
				OrderIndex:      i + 1,
				UpdatedAt:       time.Now(),
				CreatedAt:       time.Now(),
			}))
		result = append(result, workoutExercise)
	}

	return result
}

func (s *Service) selectExercisesForWorkout(exercises []*entities.Exercise, stats *dto.AnalyzeWorkoutStats) []*entities.Exercise {
	// Определяем количество упражнений в тренировке
	exerciseCount := s.rng.Intn(s.maxExercisesPerWorkout-s.minExercisesPerWorkout+1) + s.minExercisesPerWorkout

	if len(exercises) <= exerciseCount {
		return s.shuffleExercises(exercises)
	}

	// Группируем упражнения по типу
	var preferredExercises, otherExercises []*entities.Exercise
	for _, ex := range exercises {
		if ex.TypeExercise() == stats.PopularExerciseType {
			preferredExercises = append(preferredExercises, ex)
		} else {
			otherExercises = append(otherExercises, ex)
		}
	}

	// Выбираем упражнения с учётом предпочтений
	selected := s.selectBalancedExercises(preferredExercises, otherExercises, exerciseCount)

	// Финальное перемешивание
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
		// Корректируем количество предпочтительных, если не хватает других
		preferredCount = targetCount - otherCount
	}

	selected := make([]*entities.Exercise, 0, targetCount)

	// Выбираем случайные упражнения из каждой группы
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

	// Перемешиваем и берём первые count
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

	// Добавляем время на отдых между упражнениями
	if len(exercises) > 1 {
		restTime := (len(exercises) - 1) * restBetweenExercises
		totalDuration += restTime
	}

	return int(math.Round(totalCalories)), totalDuration
}

func (s *Service) determineWorkoutLevel(exercises []*entities.Exercise, baseLevel entities.WorkoutsLevel) entities.WorkoutsLevel {
	if len(exercises) == 0 {
		return baseLevel
	}

	// Карта сложности уровней
	difficultyMap := map[entities.LevelPreparation]int{
		entities.Beginner:  1,
		entities.Medium:    2,
		entities.Sportsman: 3,
	}

	// Рассчитываем среднюю сложность
	totalDifficulty := 0
	for _, ex := range exercises {
		if val, ok := difficultyMap[ex.LevelPreparation()]; ok {
			totalDifficulty += val
		}
	}

	avgDifficulty := float64(totalDifficulty) / float64(len(exercises))

	// Определяем общий уровень тренировки
	switch {
	case avgDifficulty < 1.5:
		return entities.WorkoutLight
	case avgDifficulty < 2.5:
		return entities.WorkoutMiddle
	default:
		return entities.WorkoutHard
	}
}

func (s *Service) determinePreferredLevel(userInfo *entities.UserInfo, userParams *entities.UserParams, workouts []*entities.Workout) entities.WorkoutsLevel {
	if len(workouts) == 0 {
		return s.getInitialLevel(userParams)
	}

	// Анализируем успешность тренировок разного уровня
	levelStats := s.analyzeLevelSuccess(workouts)

	// Находим уровень с наилучшим соотношением успеха
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
	task := entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
		TypeNm:      entities.TaskTypeSendNotificationPhone,
		Message:     entities.TaskMessageSendAuthomaticGeneratedWorkout,
		MaxAttempts: s.maxRetrySendNotification,
		Attribute: map[string]interface{}{
			"workout_id": workoutID.String(),
			"user_id":    userID.String(),
		},
	}))

	return s.tasksRepository.Create(ctx, task)
}
