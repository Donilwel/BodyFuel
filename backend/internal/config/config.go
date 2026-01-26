package config

import (
	"backend/internal/infrastructure/repositories/minio"
	"backend/internal/infrastructure/repositories/postgres"
	"backend/pkg/logging"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type HTTPServerConfig struct {
	Host       string `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
	Port       int    `yaml:"port" env:"PORT" envDefault:"8080"`
	ApiHost    string `yaml:"api_host" env:"API_HOST" envDefault:"0.0.0.0"`
	MetricPort int    `yaml:"metric_port" env:"METRIC_PORT" envDefault:"8081"`
	TLS        bool   `yaml:"tls" env:"TLS" envDefault:"false"`
	CertPath   string `yaml:"cert_path" env:"CERT_PATH"`
	KeyPath    string `yaml:"key_path" env:"KEY_PATH"`
}

type AppConfig struct {
	HTTPServerConfig      HTTPServerConfig `yaml:"http_server"`
	GracefulTimeout       time.Duration    `yaml:"graceful_timeout" env:"GRACEFUL_TIMEOUT" envDefault:"5s"`
	TasksTrackingDuration time.Duration    `yaml:"tasks_tracking_duration" env:"TASKS_TRACKING_DURATION" envDefault:"13s"`
}

type Config struct {
	AppConfig AppConfig       `yaml:"app"`
	Log       logging.Config  `yaml:"sage" env:"SAGE_"`
	Postgres  postgres.Config `yaml:"postgres" env-prefix:"POSTGRES_"`
	Minio     minio.Config    `yaml:"minio" env-prefix:"MINIO_"`
}

func ReadConfig(filePaths ...string) (*Config, error) {
	cfg := new(Config)

	for _, filePath := range filePaths {
		if err := cleanenv.ReadConfig(filePath, cfg); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}
