package logging

import (
	"github.com/rs/zerolog"
)

type Entry interface {
	WithFields(fields Fields) Entry

	Trace(args ...interface{})
	Tracef(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Alert(alertName string, args ...interface{})
	Alertf(alertName string, format string, args ...interface{})

	AlertFatal(alertName string, args ...interface{})
	AlertFatalf(alertName string, format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
}

type entryImpl struct {
	logger zerolog.Logger
}

func (i *entryImpl) WithFields(fields Fields) Entry {
	zFields := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		zFields[k] = v
	}
	return &entryImpl{logger: i.logger.With().Fields(zFields).Logger()}
}

func (i *entryImpl) Trace(args ...interface{}) {
	i.logger.Trace().Msgf("%v", args)
}

func (i *entryImpl) Tracef(format string, args ...interface{}) {
	i.logger.Trace().Msgf(format, args...)
}

func (i *entryImpl) Debug(args ...interface{}) {
	i.logger.Debug().Msgf("%v", args)
}

func (i *entryImpl) Debugf(format string, args ...interface{}) {
	i.logger.Debug().Msgf(format, args...)
}

func (i *entryImpl) Info(args ...interface{}) {
	i.logger.Info().Msgf("%v", args)
}

func (i *entryImpl) Infof(format string, args ...interface{}) {
	i.logger.Info().Msgf(format, args...)
}

func (i *entryImpl) Warn(args ...interface{}) {
	i.logger.Warn().Msgf("%v", args)
}

func (i *entryImpl) Warnf(format string, args ...interface{}) {
	i.logger.Warn().Msgf(format, args...)
}

func (i *entryImpl) Error(args ...interface{}) {
	i.logger.Error().Msgf("%v", args)
}

func (i *entryImpl) Errorf(format string, args ...interface{}) {
	i.logger.Error().Msgf(format, args...)
}

func (i *entryImpl) Alert(alertName string, args ...interface{}) {
	i.logger.Error().Str("alert", alertName).Msgf("%v", args)
}

func (i *entryImpl) Alertf(alertName string, format string, args ...interface{}) {
	i.logger.Error().Str("alert", alertName).Msgf(format, args...)
}

func (i *entryImpl) AlertFatal(alertName string, args ...interface{}) {
	i.Alert(alertName, args...)
}

func (i *entryImpl) AlertFatalf(alertName string, format string, args ...interface{}) {
	i.Alertf(alertName, format, args...)
}

func (i *entryImpl) Fatal(args ...interface{}) {
	i.logger.Fatal().Msgf("%v", args)
}

func (i *entryImpl) Fatalf(format string, args ...interface{}) {
	i.logger.Fatal().Msgf(format, args...)
}

func (i *entryImpl) Panic(args ...interface{}) {
	i.logger.Panic().Msgf("%v", args)
}

func (i *entryImpl) Panicf(format string, args ...interface{}) {
	i.logger.Panic().Msgf(format, args...)
}

func NewEntry(logger zerolog.Logger) Entry {
	return &entryImpl{logger: logger}
}
