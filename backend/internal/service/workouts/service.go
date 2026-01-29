package workouts

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"
)

const (
	moduleFieldName             = "module"
	workoutsModuleName          = "workouts"
	processIdleQueriesErrorType = "process_idle_queries"
	getClustersErrorType        = "get_clusters"

	locationName = "Europe/Moscow"

	limit = 10
)

type (
	UserParamsRepository interface {
		List(ctx context.Context, f dto.UserParamsFilter, withBlock bool) ([]*entities.UserParams, error)
	}

	UserInfoRepository interface {
		Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
	}

	WorkoutsRepository interface {
		TopListWithLimit(ctx context.Context, f dto.WorkoutsFilter, limit int, withBlock bool) ([]*entities.Workout, error)
		Create(ctx context.Context, workout *entities.Workout) error
	}

	TasksRepository interface {
		Create(ctx context.Context, task *entities.Task) error
	}

	TransactionManager interface {
		Do(ctx context.Context, f func(ctx context.Context) error) error
	}

	ExerciseRepository interface {
		List(ctx context.Context, f dto.ExerciseFilter, withBlock bool) ([]*entities.Exercise, error)
	}
)

type Config struct {
	TransactionManager       TransactionManager
	TasksRepository          TasksRepository
	UserParamsRepository     UserParamsRepository
	UserInfoRepository       UserInfoRepository
	WorkoutsRepository       WorkoutsRepository
	ExerciseRepository       ExerciseRepository
	WorkoutPullUserInterval  time.Duration
	MaxRetrySendNotification int
}

type Service struct {
	transactionManager      TransactionManager
	userParamsRepository    UserParamsRepository
	userInfoRepository      UserInfoRepository
	workoutsRepository      WorkoutsRepository
	exerciseRepository      ExerciseRepository
	tasksRepository         TasksRepository
	workoutPullUserInterval time.Duration

	maxRetrySendNotification int

	log logging.Entry

	cancelFn context.CancelFunc
	wg       sync.WaitGroup

	location *time.Location
}

func NewService(cfg *Config) *Service {
	loc, _ := time.LoadLocation(locationName)

	return &Service{
		transactionManager:       cfg.TransactionManager,
		workoutsRepository:       cfg.WorkoutsRepository,
		exerciseRepository:       cfg.ExerciseRepository,
		tasksRepository:          cfg.TasksRepository,
		userParamsRepository:     cfg.UserParamsRepository,
		userInfoRepository:       cfg.UserInfoRepository,
		workoutPullUserInterval:  cfg.WorkoutPullUserInterval,
		maxRetrySendNotification: cfg.MaxRetrySendNotification,

		location: loc,
		log:      logging.WithFields(logging.Fields{"module": "idles"}),
	}
}

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
				s.log.Errorf("Recovered in %s service: %v; stack trace: %s", workoutsModuleName, r, debug.Stack())
			}
			s.wg.Done()
		}()

		s.run(ctx)
	}()

	s.log.Infof("Started %s service", workoutsModuleName)

	return nil
}

func (s *Service) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.workoutPullUserInterval):
			ups, err := s.userParamsRepository.List(ctx, dto.UserParamsFilter{}, false)
			if err != nil {
				s.log.Errorf("Getting user params: %v", err)
				continue
			}

			for _, up := range ups {
				if err = s.processGenerateWorkout(ctx, up); err != nil {
					s.log.Errorf("Processing generate workout: %v", err)
				}
			}
		}
	}
}

func (s *Service) processGenerateWorkout(ctx context.Context, up *entities.UserParams) error {
	now := time.Now().In(s.location)
	userID := up.UserID()

	ui, err := s.userInfoRepository.Get(ctx, dto.UserInfoFilter{ID: &userID}, false)
	if err != nil {
		return fmt.Errorf("get user info: %w", err)
	}

	stats, err := s.analyzeWorkoutStats(ctx, ui, now)
	if err != nil {
		return fmt.Errorf("analyze workout stats: %w", err)
	}

	if stats.SkipGeneration {
		s.log.Infof("Skipping workout generation for user %s: %s", userID, stats.SkipReason)
		return nil
	}

	_, err = s.generateWorkout(ctx, up, stats)
	if err != nil {
		return fmt.Errorf("generate workout: %w", err)
	}

	//if err := s.workoutsRepository.Create(ctx, &workout); err != nil {
	//	return fmt.Errorf("create workout: %w", err)
	//}

	//if err = s.tasksRepository.Create(ctx, entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
	//	TypeNm:      entities.TaskTypeSendNotificationPhone,
	//	Message:     "new workout in your profile",
	//	MaxAttempts: s.maxRetrySendNotification,
	//	Attribute:   workout,
	//}))); err != nil {
	//	return fmt.Errorf("failed to create task: %w", err)
	//}

	return nil
}

func (s *Service) analyzeWorkoutStats(ctx context.Context, userInfo *entities.UserInfo, now time.Time) (*dto.AnalyzeWorkoutStats, error) {
	// TODO: создать анализатор последних 10 тренировок.
	userID := userInfo.ID()

	w, err := s.workoutsRepository.TopListWithLimit(ctx, dto.WorkoutsFilter{UserID: &userID}, limit, false)
	if err != nil {
		return &dto.AnalyzeWorkoutStats{}, fmt.Errorf("failed to find list workout: %w", err)
	}
	if len(w) == 0 {
		return &dto.AnalyzeWorkoutStats{
			IDUser:               userID,
			PopularExerciseType:  s.findPopularExerciseType(w),
			PopularPlaceExercise: s.findPopularPlaceExercise(w),
			AWGLevel:             s.findAWGLevelWorkout(w),
		}, nil
	}

	lastWorkout := w[0]
	if lastWorkout.Status() == entities.WorkoutStatusInActive {
		return &dto.AnalyzeWorkoutStats{
			IDUser:                  userID,
			SkipGeneration:          true,
			SkipReason:              fmt.Sprintf("find workout with status is active, need to finish"),
			LastTimeGenerateWorkout: lastWorkout.CreatedAt(),
		}, nil
	}

	if lastWorkout.UpdatedAt().Add(8 * time.Hour).After(now) {
		return &dto.AnalyzeWorkoutStats{
			IDUser:                  userID,
			SkipGeneration:          true,
			SkipReason:              fmt.Sprintf("last workout was at %s, need 8h rest", lastWorkout.UpdatedAt().Format(time.RFC3339)),
			LastTimeGenerateWorkout: lastWorkout.CreatedAt(),
		}, nil
	}

	return &dto.AnalyzeWorkoutStats{
		IDUser:                       userID,
		TotalWorkouts:                len(w),
		TotalCancelled:               s.countCancelledWorkouts(w),
		TotalNew:                     s.countNewWorkouts(w),
		TotalFinishedWorkoutsForWeek: s.countFinishedWorkoutsForWeek(w, now),
		TotalFinished:                s.countFinishedWorkouts(w),
		SkipGeneration:               false,
		SkipReason:                   "",
		PopularExerciseType:          s.findPopularExerciseType(w),
		PopularPlaceExercise:         s.findPopularPlaceExercise(w),
		AWGLevel:                     s.findAWGLevelWorkout(w),
		LastTimeGenerateWorkout:      w[0].CreatedAt(),
	}, nil
}

func (s *Service) generateWorkout(ctx context.Context, userParams *entities.UserParams, aws *dto.AnalyzeWorkoutStats) (entities.Workout, error) {
	coef, err := userParams.Lifestyle().ToCoef()
	if err != nil {
		return entities.Workout{}, fmt.Errorf("parsing lifestyle to coef: %w", err)
	}

	userLevel, err := userParams.Lifestyle().ToLevelPreparation()
	if err != nil {
		fmt.Errorf("list exercise: to level preparation: %w", err)
	}

	e, err := s.exerciseRepository.List(ctx, dto.ExerciseFilter{LevelPreparation: &userLevel, PlaceExercise: &aws.PopularPlaceExercise}, false)
	if err != nil {
		return entities.Workout{}, fmt.Errorf("list exercise: %w", err)
	}
	exercises := s.selectRandomExercisesWeighted(e, aws.PopularExerciseType)
	for _, e := range exercises {
		for i := 0; i < e.Steps(); i++ {
			fmt.Println(e.Name())
			fmt.Println(e.BaseCountReps() * int(coef))
		}
	}
	fmt.Println()
	return entities.Workout{}, nil
}

func (s *Service) selectRandomExercisesWeighted(exercises []*entities.Exercise, et entities.ExerciseType) []*entities.Exercise {
	count := 5
	if len(exercises) <= count {
		return exercises
	}

	weightedExercises := make([]*entities.Exercise, 0, len(exercises))

	for _, exercise := range exercises {
		weight := 1

		if exercise.TypeExercise() == et {
			weight = 3
		}

		for i := 0; i < weight; i++ {
			weightedExercises = append(weightedExercises, exercise)
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(weightedExercises), func(i, j int) {
		weightedExercises[i], weightedExercises[j] = weightedExercises[j], weightedExercises[i]
	})

	selected := make([]*entities.Exercise, 0, count)
	seen := make(map[uuid.UUID]bool)

	for _, exercise := range weightedExercises {
		if len(selected) >= count {
			break
		}
		if !seen[exercise.ID()] {
			seen[exercise.ID()] = true
			selected = append(selected, exercise)
		}
	}

	return selected
}

func (s *Service) Close() error {
	s.cancelFn()

	s.wg.Wait()

	s.log.Infof("Stopped %s", workoutsModuleName)

	return nil
}

func (s *Service) countNewWorkouts(workouts []*entities.Workout) int {
	count := 0
	for _, w := range workouts {
		if w.Status() == entities.WorkoutStatusCreated {
			count++
		}
	}
	return count
}

func (s *Service) countFinishedWorkouts(workouts []*entities.Workout) int {
	count := 0
	for _, w := range workouts {
		if w.Status() == entities.WorkoutStatusDone {
			count++
		}
	}
	return count
}

func (s *Service) countFinishedWorkoutsForWeek(workouts []*entities.Workout, now time.Time) int {
	count := 0
	weekAgo := now.AddDate(0, 0, -7)

	for _, w := range workouts {
		if w.Status() == entities.WorkoutStatusDone &&
			w.CreatedAt().After(weekAgo) &&
			w.CreatedAt().Before(now) {
			count++
		}
	}
	return count
}

func (s *Service) countCancelledWorkouts(workouts []*entities.Workout) int {
	count := 0
	for _, w := range workouts {
		if w.Status() == entities.WorkoutStatusFailed {
			count++
		}
	}
	return count
}

func (s *Service) findPopularExerciseType(workouts []*entities.Workout) entities.ExerciseType {
	if len(workouts) == 0 {
		return entities.UpperBody
	}
	// TODO: доделать, пока что тут чисто заглушка
	//frequency := make(map[entities.ExerciseType]int)
	//for _, w := range workouts {
	//	if exerciseType := w.; exerciseType != "" {
	//		frequency[exerciseType]++
	//	}
	//}
	//
	//var popularType entities.ExerciseType
	//maxCount := 0
	//for exerciseType, count := range frequency {
	//	if count > maxCount {
	//		maxCount = count
	//		popularType = exerciseType
	//	}
	//}

	return entities.UpperBody
}

func (s *Service) findPopularPlaceExercise(workouts []*entities.Workout) entities.PlaceExercise {
	if len(workouts) == 0 {
		return entities.Home
	}

	// TODO: доделать, пока что тут чисто заглушка
	//frequency := make(map[entities.PlaceExercise]int)
	//for _, w := range workouts {
	//	if place := w.PlaceExercise(); place != "" {
	//		frequency[place]++
	//	}
	//}
	//
	//var popularPlace entities.PlaceExercise
	//maxCount := 0
	//for place, count := range frequency {
	//	if count > maxCount {
	//		maxCount = count
	//		popularPlace = place
	//	}
	//}

	return entities.Home
}

func (s *Service) findAWGLevelWorkout(workouts []*entities.Workout) entities.WorkoutsLevel {
	if len(workouts) == 0 {
		return entities.WorkoutLight
	}

	frequency := make(map[entities.WorkoutsLevel]int)
	for _, w := range workouts {
		frequency[w.Level()]++
	}

	var popularLevel entities.WorkoutsLevel
	maxCount := 0

	for level, count := range frequency {
		if count > maxCount {
			maxCount = count
			popularLevel = level
		}
	}

	return popularLevel
}
