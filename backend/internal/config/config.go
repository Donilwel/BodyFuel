package config

import (
	"backend/internal/infrastructure/repositories/minio"
	"backend/internal/infrastructure/repositories/postgres"
	"backend/pkg/cache"
	"backend/pkg/logging"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServerConfig struct {
	Host         string        `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
	Port         int           `yaml:"port" env:"PORT" envDefault:"8080"`
	ApiHost      string        `yaml:"api_host" env:"API_HOST" envDefault:"0.0.0.0"`
	MetricPort   int           `yaml:"metric_port" env:"METRIC_PORT" envDefault:"8081"`
	TLS          bool          `yaml:"tls" env:"TLS" envDefault:"false"`
	CertPath     string        `yaml:"cert_path" env:"CERT_PATH"`
	KeyPath      string        `yaml:"key_path" env:"KEY_PATH"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" envDefault:"60s"`
}

type AppConfig struct {
	HTTPServerConfig      HTTPServerConfig `yaml:"http_server"`
	GracefulTimeout       time.Duration    `yaml:"graceful_timeout" env:"GRACEFUL_TIMEOUT" envDefault:"5s"`
	TasksTrackingDuration time.Duration    `yaml:"tasks_tracking_duration" env:"TASKS_TRACKING_DURATION" envDefault:"13s"`
	WorkoutsConfig        WorkoutsConfig   `yaml:"workouts_config" env-prefix:"WORKOUTS_CONFIG_"`
}

type WorkoutsConfig struct {
	WorkoutPullUserInterval time.Duration `yaml:"workout_pull_user_interval" env:"WORKOUT_PULL_USER_INTERVAL" envDefault:"60s"`
	LimitGenerateWorkouts   int           `yaml:"limit_generate_workouts,omitempty" env:"LIMIT_GENERATE_WORKS" envDefault:"3"`
}

type SendGridConfig struct {
	APIKey    string `yaml:"api_key" env:"API_KEY"`
	FromEmail string `yaml:"from_email" env:"FROM_EMAIL"`
	FromName  string `yaml:"from_name" env:"FROM_NAME"`
}

type TwilioConfig struct {
	AccountSID string `yaml:"account_sid" env:"ACCOUNT_SID"`
	AuthToken  string `yaml:"auth_token" env:"AUTH_TOKEN"`
	FromPhone  string `yaml:"from_phone" env:"FROM_PHONE"`
}

type APNsConfig struct {
	KeyPath  string `yaml:"key_path" env:"KEY_PATH"`
	KeyID    string `yaml:"key_id" env:"KEY_ID"`
	TeamID   string `yaml:"team_id" env:"TEAM_ID"`
	BundleID string `yaml:"bundle_id" env:"BUNDLE_ID"`
	Sandbox  bool   `yaml:"sandbox" env:"SANDBOX" envDefault:"true"`
}

type OpenAIConfig struct {
	APIKey string `yaml:"api_key" env:"API_KEY"`
}

type Config struct {
	AppConfig AppConfig       `yaml:"app"`
	Log       logging.Config  `yaml:"sage" env:"SAGE_"`
	Postgres  postgres.Config `yaml:"postgres" env-prefix:"POSTGRES_"`
	Minio     minio.Config    `yaml:"minio" env-prefix:"MINIO_"`
	Redis     cache.Config    `yaml:"redis" env-prefix:"REDIS_"`
	SendGrid  SendGridConfig  `yaml:"sendgrid" env-prefix:"SENDGRID_"`
	Twilio    TwilioConfig    `yaml:"twilio" env-prefix:"TWILIO_"`
	APNs      APNsConfig      `yaml:"apns" env-prefix:"APNS_"`
	OpenAI    OpenAIConfig    `yaml:"openai" env-prefix:"OPENAI_"`
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
