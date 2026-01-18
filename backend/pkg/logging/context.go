package logging

import "context"

const loggerKey = "logger"

func CtxWithLogger(ctx context.Context, logger Entry) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func GetLoggerFromContext(ctx context.Context) Entry {
	v := ctx.Value(loggerKey)
	if l, ok := v.(Entry); ok {
		return l
	}

	return WithFields(make(Fields))
}
