package app

import (
	"backend/internal/config"
	"backend/internal/infrastructure/repositories/postgres"
	"backend/pkg/logging"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type BackgroundWorker interface {
	Run() error
	io.Closer
}

type App struct {
	cfg              *config.Config
	httpServer       *http.Server
	metricHttpServer *http.Server
	workers          []BackgroundWorker
	closers          []io.Closer
	healthMetric     prometheus.Gauge
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

	appMetrics := initializeMetrics()

	workers := make([]BackgroundWorker, 0, 8)

	transactionManager := postgres.NewTransactionManager(db)
	userInfoRepository := postgres.NewUserInfoRepository(db)
	userParamsRepository := postgres.NewUserParamsRepository(db)

	return &App{
		cfg: cfg,
		metricHttpServer: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.AppConfig.HTTPServerConfig.Host, cfg.AppConfig.HTTPServerConfig.MetricPort),
			Handler: metricsRouter,
		},
		workers:      workers,
		closers:      closers,
		healthMetric: appMetric.HealthMetric,
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

	go func() {
		log.Printf("Metrics server is listening on %d", a.cfg.AppConfig.HTTPServerConfig.MetricPort)

		if err := a.metricHttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed listen and serve metrics http server: %v", err)
		}
		log.Printf("Metrics server closed")
	}()

	for _, s := range a.workers {
		if err := s.Run(); err != nil {
			log.Fatalf("Failed to start service: %v", err)
		}
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	a.healthMetric.Set(1)
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
