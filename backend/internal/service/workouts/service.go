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
)

const (
	moduleFieldName    = "module"
	workoutsModuleName = "workouts"

	locationName = "Europe/Moscow"

	// Константы для генерации тренировок
	defaultWorkoutsLimit   = 10
	minExercisesPerWorkout = 4
	maxExercisesPerWorkout = 8
	restBetweenWorkouts    = 8 * time.Hour
	daysInWeek             = 7
	batchSize              = 100
	maxConcurrentUsers     = 10

	// Коэффициенты для расчета
	preferredExercisesPercent = 0.6 // 60% предпочтительных упражнений
	restBetweenExercises      = 60  // 60 секунд отдыха между упражнениями
	caloriesTolerancePercent  = 0.2 // 20% допустимое отклонение от целевых калорий
)

type (
	UserParamsRepository interface {
		List(ctx context.Context, f dto.UserParamsFilter, withBlock bool) ([]*entities.UserParams, error)
	}

	UserInfoRepository interface {
		Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
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
		CreateBulk(ctx context.Context, workoutExercises []*entities.WorkoutsExercise) error
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
	WorkoutPullUserInterval   time.Duration
	MaxRetrySendNotification  int
	LimitGenerateWorkouts     int
	MinExercisesPerWorkout    int
	MaxExercisesPerWorkout    int
	EnableNotifications       bool
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
	workoutPullUserInterval   time.Duration
	limitGenerateWorkouts     int
	minExercisesPerWorkout    int
	maxExercisesPerWorkout    int
	maxRetrySendNotification  int
	enableNotifications       bool

	log logging.Entry

	cancelFn context.CancelFunc
	wg       sync.WaitGroup

	location *time.Location
	rng      *rand.Rand
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
	minExercises := cfg.MinExercisesPerWorkout
	if minExercises == 0 {
		minExercises = minExercisesPerWorkout
	}

	maxExercises := cfg.MaxExercisesPerWorkout
	if maxExercises == 0 {
		maxExercises = maxExercisesPerWorkout
	}

	limitGenerate := cfg.LimitGenerateWorkouts
	if limitGenerate == 0 {
		limitGenerate = 3
	}

	return &Service{
		transactionManager:        cfg.TransactionManager,
		workoutsRepository:        cfg.WorkoutsRepository,
		exerciseRepository:        cfg.ExerciseRepository,
		workoutExerciseRepository: cfg.WorkoutExerciseRepository,
		tasksRepository:           cfg.TasksRepository,
		userParamsRepository:      cfg.UserParamsRepository,
		userInfoRepository:        cfg.UserInfoRepository,
		userWeightRepository:      cfg.UserWeightRepository,
		workoutPullUserInterval:   cfg.WorkoutPullUserInterval,
		maxRetrySendNotification:  cfg.MaxRetrySendNotification,
		limitGenerateWorkouts:     limitGenerate,
		minExercisesPerWorkout:    minExercises,
		maxExercisesPerWorkout:    maxExercises,
		enableNotifications:       cfg.EnableNotifications,

		location: loc,
		rng:      rng,
		log:      logging.WithFields(logging.Fields{"module": "workouts"}),
	}
}

// ==================== ПУБЛИЧНЫЕ МЕТОДЫ ====================

func (s *Service) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel

	s.log = logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
		moduleFieldName: workoutsModuleName,
	})

	s.wg.Add(1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.log.Errorf("Recovered in %s service: %v; stack trace: %s",
					workoutsModuleName, r, debug.Stack())
			}
			s.wg.Done()
		}()

		s.run(ctx)
	}()

	s.log.Infof("Started %s service with interval: %v", workoutsModuleName, s.workoutPullUserInterval)

	return nil
}

func (s *Service) Close() error {
	s.cancelFn()
	s.wg.Wait()
	s.log.Infof("Stopped %s service", workoutsModuleName)
	return nil
}

// GenerateWorkoutForUser генерирует тренировку для конкретного пользователя (публичный API)
func (s *Service) GenerateWorkoutForUser(ctx context.Context, userID uuid.UUID) (*entities.Workout, error) {

	// Получаем параметры пользователя
	userParamsFilter := dto.UserParamsFilter{
		UserID: &userID,
	}

	userParamsList, err := s.userParamsRepository.List(ctx, userParamsFilter, false)
	if err != nil {
		return nil, fmt.Errorf("get user params: %w", err)
	}

	if len(userParamsList) == 0 {
		return nil, fmt.Errorf("user params not found for user %s", userID)
	}

	userParams := userParamsList[0]

	// Получаем информацию о пользователе
	userInfoFilter := dto.UserInfoFilter{
		ID: &userID,
	}

	userInfo, err := s.userInfoRepository.Get(ctx, userInfoFilter, false)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	// Анализируем статистику
	now := time.Now().In(s.location)
	stats, err := s.analyzeWorkoutStats(ctx, userInfo, userParams, now)
	if err != nil {
		return nil, fmt.Errorf("analyze stats: %w", err)
	}

	// Если генерация пропускается, но пользователь запросил принудительно - все равно генерируем
	if stats.SkipGeneration {
		s.log.Infof("Forcing workout generation despite: %s", stats.SkipReason)
		stats.SkipGeneration = false
	}

	// Генерируем тренировку
	workout, err := s.generateWorkout(ctx, userParams, stats)
	if err != nil {
		return nil, fmt.Errorf("generate workout: %w", err)
	}

	return workout, nil
}

func (s *Service) GetUserWorkoutStats(ctx context.Context, userID uuid.UUID) (*dto.AnalyzeWorkoutStats, error) {
	userInfoFilter := dto.UserInfoFilter{
		ID: &userID,
	}

	userInfo, err := s.userInfoRepository.Get(ctx, userInfoFilter, false)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	userParamsFilter := dto.UserParamsFilter{
		UserID: &userID,
	}

	userParamsList, err := s.userParamsRepository.List(ctx, userParamsFilter, false)
	if err != nil {
		return nil, fmt.Errorf("get user params: %w", err)
	}

	var userParams *entities.UserParams
	if len(userParamsList) > 0 {
		userParams = userParamsList[0]
	}

	now := time.Now().In(s.location)

	return s.analyzeWorkoutStats(ctx, userInfo, userParams, now)
}

func (s *Service) run(ctx context.Context) {
	ticker := time.NewTicker(s.workoutPullUserInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Workout service context cancelled, stopping")
			return
		case <-ticker.C:
			if err := s.processAllUsers(ctx); err != nil {
				s.log.Errorf("Failed to process users: %v", err)
			}
		}
	}
}

func (s *Service) processAllUsers(ctx context.Context) error {
	offset := 0

	for {
		// Фильтр для получения всех пользователей с параметрами
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

		if err := s.processUserBatch(ctx, users); err != nil {
			s.log.Errorf("Error processing user batch: %v", err)
		}

		if len(users) < batchSize {
			break
		}
		offset += batchSize
	}

	return nil
}

func (s *Service) processUserBatch(ctx context.Context, users []*entities.UserParams) error {
	semaphore := make(chan struct{}, maxConcurrentUsers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, up := range users {
		up := up

		select {
		case <-ctx.Done():
			return ctx.Err()
		case semaphore <- struct{}{}:
			wg.Add(1)

			go func() {
				defer func() {
					<-semaphore
					wg.Done()
				}()

				if err := s.processGenerateWorkout(ctx, up); err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("user %s: %w", up.UserID(), err))
					mu.Unlock()
				}
			}()
		}
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors while processing users", len(errors))
	}

	return nil
}

func (s *Service) processGenerateWorkout(ctx context.Context, up *entities.UserParams) error {
	now := time.Now().In(s.location)
	userID := up.UserID()

	// Получаем информацию о пользователе
	userInfoFilter := dto.UserInfoFilter{
		ID: &userID,
	}

	ui, err := s.userInfoRepository.Get(ctx, userInfoFilter, false)
	if err != nil {
		return fmt.Errorf("get user info: %w", err)
	}

	// Анализируем статистику тренировок
	stats, err := s.analyzeWorkoutStats(ctx, ui, up, now)
	if err != nil {
		return fmt.Errorf("analyze workout stats: %w", err)
	}

	if stats.SkipGeneration {
		s.log.Infof("Skipping workout generation: %s", stats.SkipReason)
		return nil
	}

	// Генерируем тренировку
	workout, err := s.generateWorkout(ctx, up, stats)
	if err != nil {
		return fmt.Errorf("generate workout: %w", err)
	}

	s.log.Infof("Successfully generated workout %s with %d calories, %d minutes",
		workout.ID(), workout.PredictionCalories(), workout.Duration())

	// Создаем задачу на уведомление, если включено
	if s.enableNotifications {
		if err := s.createNotificationTask(ctx, workout.ID(), up.UserID()); err != nil {
			s.log.Errorf("Failed to create notification task: %v", err)
		}
	}

	return nil
}

func (s *Service) analyzeWorkoutStats(ctx context.Context, userInfo *entities.UserInfo, userParams *entities.UserParams, now time.Time) (*dto.AnalyzeWorkoutStats, error) {
	userID := userInfo.ID()

	// Получаем последние тренировки пользователя
	workoutsFilter := dto.WorkoutsFilter{
		UserID: &userID,
	}

	workouts, err := s.workoutsRepository.TopListWithLimit(ctx, workoutsFilter, defaultWorkoutsLimit, false)
	if err != nil {
		return nil, fmt.Errorf("failed to find list workout: %w", err)
	}
	targetWorkoutsWeek := userParams.TargetWorkoutsWeeks()
	// Значения по умолчанию
	targetWorkoutsPerWeek := 3
	if userParams != nil && &targetWorkoutsWeek != nil {
		targetWorkoutsPerWeek = targetWorkoutsWeek
	}
	// Анализируем популярные типы упражнений и места
	popularExerciseType, popularPlaceExercise := s.analyzeUserPreferences(ctx, userID, workouts)

	// Определяем предпочтительный уровень
	preferredLevel := s.determinePreferredLevel(userInfo, userParams, workouts)

	stats := &dto.AnalyzeWorkoutStats{
		IDUser:                userID,
		PopularExerciseType:   popularExerciseType,
		PopularPlaceExercise:  popularPlaceExercise,
		AWGLevel:              preferredLevel,
		TargetWorkoutsPerWeek: targetWorkoutsPerWeek,
	}

	// Если тренировок нет - это новый пользователь
	if len(workouts) == 0 {
		stats.TotalWorkouts = 0
		stats.SkipGeneration = false
		return stats, nil
	}

	lastWorkout := workouts[0]
	stats.LastTimeGenerateWorkout = lastWorkout.CreatedAt()

	// Проверяем наличие активной тренировки
	if lastWorkout.Status() == entities.WorkoutStatusInActive {
		stats.SkipGeneration = true
		stats.SkipReason = "found active workout, need to finish it first"
		return stats, nil
	}

	// Считаем количество сгенерированных, но не использованных тренировок
	unusedCount := 0
	for _, w := range workouts {
		if w.Status() == entities.WorkoutStatusCreated {
			unusedCount++
		}
	}

	if unusedCount >= s.limitGenerateWorkouts {
		stats.SkipGeneration = true
		stats.SkipReason = fmt.Sprintf("already have %d unused workouts (max: %d)",
			unusedCount, s.limitGenerateWorkouts)
		return stats, nil
	}

	// Проверяем, прошло ли достаточно времени с последней тренировки
	if lastWorkout.UpdatedAt().Add(restBetweenWorkouts).After(now) {
		timeLeft := time.Until(lastWorkout.UpdatedAt().Add(restBetweenWorkouts))
		stats.SkipGeneration = true
		stats.SkipReason = fmt.Sprintf("need to rest %.1f more hours", timeLeft.Hours())
		return stats, nil
	}

	// Проверяем, достигнута ли цель по тренировкам за неделю
	weeklyWorkouts := s.countFinishedWorkoutsForWeek(workouts, now)
	if weeklyWorkouts >= targetWorkoutsPerWeek {
		stats.SkipGeneration = true
		stats.SkipReason = fmt.Sprintf("already completed %d workouts this week (target: %d)",
			weeklyWorkouts, targetWorkoutsPerWeek)
		return stats, nil
	}

	// Заполняем полную статистику
	stats.TotalWorkouts = len(workouts)
	stats.TotalCancelled = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusFailed)
	stats.TotalNew = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusCreated)
	stats.TotalFinished = s.countWorkoutsByStatus(workouts, entities.WorkoutStatusDone)
	stats.TotalFinishedWorkoutsForWeek = weeklyWorkouts
	stats.SkipGeneration = false

	return stats, nil
}

func (s *Service) analyzeUserPreferences(ctx context.Context, userID uuid.UUID, workouts []*entities.Workout) (entities.ExerciseType, entities.PlaceExercise) {
	// По умолчанию
	defaultExerciseType := entities.UpperBody
	defaultPlace := entities.Home

	if len(workouts) == 0 {
		return defaultExerciseType, defaultPlace
	}

	// TODO: Получить реальные предпочтения из workout_exercises
	// Сейчас заглушка

	return defaultExerciseType, defaultPlace
}

func (s *Service) generateWorkout(ctx context.Context, userParams *entities.UserParams, stats *dto.AnalyzeWorkoutStats) (*entities.Workout, error) {
	// Получаем коэффициент нагрузки из образа жизни
	coef, err := userParams.Lifestyle().ToCoef()
	if err != nil {
		return nil, fmt.Errorf("parsing lifestyle to coef: %w", err)
	}

	userLevel, err := userParams.Lifestyle().ToLevelPreparation()
	if err != nil {
		return nil, fmt.Errorf("failed to determine user level: %w", err)
	}

	// Получаем упражнения для данного уровня
	exerciseFilter := dto.ExerciseFilter{
		LevelPreparation: (*entities.LevelPreparation)(&userLevel),
		PlaceExercise:    &stats.PopularPlaceExercise,
	}

	exercises, err := s.exerciseRepository.List(ctx, exerciseFilter, false)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}

	if len(exercises) == 0 {
		// Если нет упражнений для предпочитаемого места, ищем все для данного уровня
		exerciseFilter.PlaceExercise = nil
		exercises, err = s.exerciseRepository.List(ctx, exerciseFilter, false)
		if err != nil {
			return nil, fmt.Errorf("list exercises without place filter: %w", err)
		}

		if len(exercises) == 0 {
			return nil, fmt.Errorf("no exercises found for level %s", userLevel)
		}
	}

	// Выбираем упражнения для тренировки
	selectedExercises := s.selectExercisesForWorkout(exercises, stats)

	// Рассчитываем параметры тренировки
	totalCalories, totalDuration := s.calculateWorkoutParams(selectedExercises, coef)

	// Создаем тренировку
	workout := entities.NewWorkout(entities.WithWorkoutInitSpec(entities.WorkoutInitSpec{
		ID:                 uuid.New(),
		UserID:             stats.IDUser,
		Level:              s.determineWorkoutLevel(selectedExercises, entities.WorkoutsLevel(userLevel)),
		Status:             entities.WorkoutStatusCreated,
		PredictionCalories: totalCalories,
		Duration:           int64(totalDuration),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}))

	// Сохраняем тренировку в транзакции
	err = s.transactionManager.Do(ctx, func(ctx context.Context) error {
		if err := s.workoutsRepository.Create(ctx, workout); err != nil {
			return fmt.Errorf("create workout: %w", err)
		}

		// Создаем связи тренировки с упражнениями
		workoutExercises := make([]*entities.WorkoutsExercise, 0, len(selectedExercises))
		for i, ex := range selectedExercises {
			entities.NewWorkoutsExercise(entities.WithWorkoutsExerciseInitSpec(entities.WorkoutsExerciseInitSpec{
				WorkoutID:  workout.ID(),
				ExerciseID: ex.ID(),
				Calories:   323,
				Status:     entities.ExerciseStatusPending,
				OrderIndex: i + 1,
				CreatedAt:  time.Time{},
			}))
		}

		if err := s.workoutExerciseRepository.CreateBulk(ctx, workoutExercises); err != nil {
			return fmt.Errorf("create workout exercises: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	// Добавляем упражнения в объект тренировки для использования вне транзакции
	workout.SetExercises(selectedExercises)

	return workout, nil
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

	// Определяем, сколько предпочтительных упражнений включить
	preferredCount := int(float64(exerciseCount) * preferredExercisesPercent)
	if preferredCount > len(preferredExercises) {
		preferredCount = len(preferredExercises)
	}

	otherCount := exerciseCount - preferredCount
	if otherCount > len(otherExercises) {
		otherCount = len(otherExercises)
		// Если не хватает других упражнений, добираем предпочтительными
		preferredCount = exerciseCount - otherCount
	}

	// Выбираем случайные упражнения из каждой группы
	selected := make([]*entities.Exercise, 0, exerciseCount)

	// Перемешиваем и выбираем предпочтительные
	s.shuffleExercises(preferredExercises)
	selected = append(selected, preferredExercises[:preferredCount]...)

	// Перемешиваем и выбираем остальные
	s.shuffleExercises(otherExercises)
	selected = append(selected, otherExercises[:otherCount]...)

	// Финальное перемешивание порядка упражнений
	return s.shuffleExercises(selected)
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

	// Анализируем сложность выбранных упражнений
	totalDifficulty := 0
	difficultyMap := map[entities.LevelPreparation]int{
		entities.Beginner:  1,
		entities.Medium:    2,
		entities.Sportsman: 3,
	}

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
	case avgDifficulty < 3.5:
		return entities.WorkoutHard
	}

	return entities.WorkoutLight
}

func (s *Service) determinePreferredLevel(userInfo *entities.UserInfo, userParams *entities.UserParams, workouts []*entities.Workout) entities.WorkoutsLevel {
	if len(workouts) == 0 {
		// Для нового пользователя определяем начальный уровень
		return s.getInitialLevel(userParams)
	}

	// Анализируем успешность тренировок разного уровня
	levelStats := make(map[entities.WorkoutsLevel]struct {
		total   int
		success int
	})

	for _, w := range workouts {
		stats := levelStats[w.Level()]
		stats.total++
		if w.Status() == entities.WorkoutStatusDone {
			stats.success++
		}
		levelStats[w.Level()] = stats
	}

	// Находим уровень с наилучшим соотношением успеха
	var bestLevel entities.WorkoutsLevel
	bestRatio := 0.0

	for level, stats := range levelStats {
		if stats.total > 0 {
			ratio := float64(stats.success) / float64(stats.total)
			if ratio > bestRatio {
				bestRatio = ratio
				bestLevel = level
			}
		}
	}

	if bestLevel == "" {
		return entities.WorkoutLight
	}

	return bestLevel
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

func (s *Service) createNotificationTask(ctx context.Context, workoutID, userID uuid.UUID) error {
	task := entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
		TypeNm:      entities.TaskTypeSendNotificationPhone,
		Message:     entities.TaskMessageSendAuthomaticGeneratedWorkout,
		MaxAttempts: s.maxRetrySendNotification,
		Attribute: map[string]interface{}{
			"workout_id": workoutID,
			"user_id":    userID,
		},
	}))

	return s.tasksRepository.Create(ctx, task)
}
