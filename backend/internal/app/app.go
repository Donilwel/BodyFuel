package app

import (
	"backend/internal/config"
	"backend/internal/handlers"
	v1 "backend/internal/handlers/v1"
	"backend/internal/infrastructure/repositories/postgres"
	"backend/internal/service/auth"
	"backend/internal/service/avatar"
	"backend/internal/service/crud"
	"backend/internal/service/executor"
	"backend/internal/service/nutricion"
	"backend/internal/service/recomendation"
	"backend/internal/service/workouts"
	"backend/pkg/ai"
	"backend/pkg/cache"
	"backend/pkg/logging"
	notifapns "backend/pkg/notifications/apns"
	notifsg "backend/pkg/notifications/sendgrid"
	notiftwilio "backend/pkg/notifications/twilio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BackgroundWorker interface {
	Run() error
	io.Closer
}

type App struct {
	cfg        *config.Config
	httpServer *http.Server
	workers    []BackgroundWorker
	closers    []io.Closer
}

func NewApp(configPaths ...string) *App {
	cfg, err := config.ReadConfig(configPaths...)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	if err := logging.Configure(&cfg.Log); err != nil {
		log.Fatalf("Failed to configure logger: %v", err)
	}
	ctx := context.Background()

	logger := logging.GetLoggerFromContext(ctx)

	db, err := initDB(cfg.Postgres)
	if err != nil {
		logger.Fatalf("Failed to init db: %v", err)
	}

	s3, err := initS3(cfg.Minio)
	if err != nil {
		logger.Fatalf("Failed to init minio: %v", err)
	}

	workers := make([]BackgroundWorker, 0, 8)
	closers := make([]io.Closer, 0, 8)

	// Redis is optional — if addr is empty or unreachable we run without cache.
	var redisClient *cache.Client
	if cfg.Redis.Addr != "" {
		rc, rerr := cache.NewClient(cfg.Redis)
		if rerr != nil {
			logger.Warnf("Redis unavailable (%v) — running without AI cache", rerr)
		} else {
			redisClient = rc
			closers = append(closers, redisClient)
		}
	}

	transactionManager := postgres.NewTransactionManager(db)
	userInfoRepository := postgres.NewUserInfoRepository(db)
	userParamsRepository := postgres.NewUserParamsRepository(db)
	userWeightRepository := postgres.NewUserWeightRepository(db)
	exercisesRepository := postgres.NewExerciseRepository(db)
	tasksRepository := postgres.NewTasksRepository(db)
	workoutsRepository := postgres.NewWorkoutRepository(db)
	workoutsExerciseRepository := postgres.NewWorkoutsExerciseRepository(db)
	userDevicesRepository := postgres.NewUserDevicesRepository(db)
	userCaloriesRepository := postgres.NewUserCaloriesRepository(db)
	userRefreshTokensRepository := postgres.NewUserRefreshTokensRepository(db)
	userVerificationCodesRepository := postgres.NewUserVerificationCodesRepository(db)
	userFoodRepository := postgres.NewUserFoodRepository(db)
	userRecommendationsRepository := postgres.NewUserRecommendationsRepository(db)

	authService := auth.NewService(&auth.Config{
		TransactionManager:          transactionManager,
		UserInfoRepository:          userInfoRepository,
		UserRefreshTokensRepository: userRefreshTokensRepository,
		VerificationCodesRepository: userVerificationCodesRepository,
		TasksRepository:             tasksRepository,
	})

	crudService := crud.NewService(&crud.Config{
		TransactionManager:         transactionManager,
		UserInfoRepository:         userInfoRepository,
		UserParamsRepository:       userParamsRepository,
		UserWeightRepository:       userWeightRepository,
		TasksRepository:            tasksRepository,
		ExercisesRepository:        exercisesRepository,
		WorkoutsRepository:         workoutsRepository,
		WorkoutsExerciseRepository: workoutsExerciseRepository,
		UserDevicesRepository:      userDevicesRepository,
		UserCaloriesRepository:     userCaloriesRepository,
		Log:                        logger,
	})

	avatarService := avatar.NewService(avatar.Config{
		S3:         s3,
		Bucket:     cfg.Minio.Bucket,
		PresignTTL: cfg.Minio.PresignTTL,
		PublicURL:  cfg.Minio.PublicURL,
	})

	workoutService := workouts.NewService(&workouts.Config{
		TransactionManager:        transactionManager,
		TasksRepository:           tasksRepository,
		ExerciseRepository:        exercisesRepository,
		UserInfoRepository:        userInfoRepository,
		UserParamsRepository:      userParamsRepository,
		UserWeightRepository:      userWeightRepository,
		WorkoutExerciseRepository: workoutsExerciseRepository,
		WorkoutsRepository:        workoutsRepository,
		UserDevicesRepository:     userDevicesRepository,
		UserFoodRepository:        userFoodRepository,
		WorkoutPullUserInterval:   cfg.AppConfig.WorkoutsConfig.WorkoutPullUserInterval,
		LimitGenerateWorkouts:     cfg.AppConfig.WorkoutsConfig.LimitGenerateWorkouts,
	})
	workers = append(workers, workoutService)

	emailClient := notifsg.NewClient(notifsg.Config{
		APIKey:    cfg.SendGrid.APIKey,
		FromEmail: cfg.SendGrid.FromEmail,
		FromName:  cfg.SendGrid.FromName,
	})

	smsClient := notiftwilio.NewClient(notiftwilio.Config{
		AccountSID: cfg.Twilio.AccountSID,
		AuthToken:  cfg.Twilio.AuthToken,
		FromPhone:  cfg.Twilio.FromPhone,
	})

	var pushClient executor.PushClient
	if cfg.APNs.KeyPath != "" {
		apnsClient, err := notifapns.NewClient(notifapns.Config{
			KeyPath:  cfg.APNs.KeyPath,
			KeyID:    cfg.APNs.KeyID,
			TeamID:   cfg.APNs.TeamID,
			BundleID: cfg.APNs.BundleID,
			Sandbox:  cfg.APNs.Sandbox,
		})
		if err != nil {
			logger.Fatalf("Failed to init APNs client: %v", err)
		}
		pushClient = apnsClient
	}

	executorService := executor.NewService(&executor.Config{
		TransactionManager: transactionManager,
		TasksRepository:    tasksRepository,
		UserInfoRepository: userInfoRepository,
		EmailClient:        emailClient,
		SMSClient:          smsClient,
		PushClient:         pushClient,
		QueryDelay:         cfg.AppConfig.TasksTrackingDuration,
	})
	workers = append(workers, executorService)

	aiClient := ai.NewClient(cfg.OpenAI.APIKey)

	nutritionService := nutricion.NewService(&nutricion.Config{
		UserFoodRepository: userFoodRepository,
		AIClient:           aiClient,
		StorageService:     avatarService,
		RecipeCache:        redisClient,
	})

	recommendationService := recomendation.NewService(&recomendation.Config{
		RecommendationRepository: userRecommendationsRepository,
		UserParamsRepository:     userParamsRepository,
		UserWeightRepository:     userWeightRepository,
		AIClient:                 aiClient,
		RecommendationCache:      redisClient,
	})

	validator := validator.New()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	handlers.Register(
		router.Group(""),
		cfg.AppConfig.HTTPServerConfig.ApiHost,
		v1.NewHandlers(v1.Config{
			AuthService:           authService,
			CRUDService:           crudService,
			WorkoutService:        workoutService,
			AvatarService:         avatarService,
			NutritionService:      nutritionService,
			RecommendationService: recommendationService,
			Validator:             *validator,
			Log:                   logger,
		}),
	)

	return &App{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.AppConfig.HTTPServerConfig.Host, cfg.AppConfig.HTTPServerConfig.Port),
			Handler:      router,
			ReadTimeout:  cfg.AppConfig.HTTPServerConfig.ReadTimeout,
			WriteTimeout: cfg.AppConfig.HTTPServerConfig.WriteTimeout,
			IdleTimeout:  cfg.AppConfig.HTTPServerConfig.IdleTimeout,
		},
		workers: workers,
		closers: closers,
	}
}

func (a *App) Run() {
	go func() {
		if a.cfg.AppConfig.HTTPServerConfig.TLS {
			if err := a.httpServer.ListenAndServeTLS(
				a.cfg.AppConfig.HTTPServerConfig.CertPath,
				a.cfg.AppConfig.HTTPServerConfig.KeyPath,
			); err != nil {
				log.Fatalf("Failed to start http server: %v", err)
			}

			return
		}

		log.Printf("Server is listening on %d", a.cfg.AppConfig.HTTPServerConfig.Port)

		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed listen and serve http server: %v", err)
		}
		log.Printf("Server closed")
	}()

	for _, s := range a.workers {
		if err := s.Run(); err != nil {
			log.Fatalf("Failed to start service: %v", err)
		}
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Println("Application is running")
	a.waitGracefulShutdown()
}

func (a *App) waitGracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(
		quit,
		syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		os.Interrupt,
	)

	log.Printf("Caught signal %s. Shutting down...", <-quit)

	done := make(chan struct{})
	go func() {
		if err := a.httpServer.Shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown http server: %v", err)
		}

		for _, s := range a.workers {
			if err := s.Close(); err != nil {
				log.Fatalf("Failed to start service: %v", err)
			}
		}

		for _, c := range a.closers {
			if err := c.Close(); err != nil {
				log.Fatalf("Failed to close: %v", err)
			}
		}

		done <- struct{}{}
	}()

	select {
	case <-time.After(a.cfg.AppConfig.GracefulTimeout):
	case <-done:
		log.Println("Http server stopped")
	}
}

func initMetricServer(metricHandler http.Handler) *gin.Engine {
	router := gin.New()
	router.GET("/metrics", gin.WrapH(metricHandler))
	return router
}
