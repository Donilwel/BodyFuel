package logging

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

const alertNameField = "alert_name"

type Fields map[string]interface{}

var logger zerolog.Logger

func init() {
	zerolog.MessageFieldName = "msg"
	zerolog.TimeFieldFormat = time.RFC3339
	logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func Configure(config *Config) (err error) {
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		return err
	}
	logger = logger.Level(level)

	fields := map[string]any{
		"deployment_environment": config.DeploymentEnvironment,
		"system":                 config.SageSystem,
		"env":                    config.StandType,
		"inst":                   config.PodName,
	}

	logger = logger.With().Fields(fields).Logger()

	logger = logger.Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		e.Str("@timestamp", time.Now().Format(time.RFC3339))
	}))

	return nil
}

func Trace(args ...interface{}) {
	logger.Trace().Msgf("%v", args)
}

func Tracef(format string, args ...interface{}) {
	logger.Trace().Msgf(format, args...)
}

func Debug(args ...interface{}) {
	logger.Debug().Msgf("%v", args)
}

func Debugf(format string, args ...interface{}) {
	logger.Debug().Msgf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info().Msgf("%v", args)
}

func Infof(format string, args ...interface{}) {
	logger.Info().Msgf(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn().Msgf("%v", args)
}

func Warnf(format string, args ...interface{}) {
	logger.Warn().Msgf(format, args...)
}

func Error(args ...interface{}) {
	logger.Error().Msgf("%v", args)
}

func Errorf(format string, args ...interface{}) {
	logger.Error().Msgf(format, args...)
}

func Alert(alertName string, args ...interface{}) {
	logger.Error().Str(alertNameField, alertName).Msgf("%v", args)
}

func Alertf(alertName string, format string, args ...interface{}) {
	logger.Error().Str(alertNameField, alertName).Msgf(format, args...)
}

func AlertFatal(alertName string, args ...interface{}) {
	Alert(alertName, args...)
	os.Exit(1)
}

func AlertFatalf(alertName string, format string, args ...interface{}) {
	Alertf(alertName, format, args...)
	os.Exit(1)
}

func Fatal(args ...interface{}) {
	logger.Fatal().Msgf("%v", args)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatal().Msgf(format, args...)
}

func Panic(args ...interface{}) {
	logger.Panic().Msgf("%v", args)
}

func Panicf(format string, args ...interface{}) {
	logger.Panic().Msgf(format, args...)
}

func WithFields(fields Fields) Entry {
	zlog := logger.With().Fields(fields).Logger()
	return NewEntry(zlog)
}
