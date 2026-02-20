package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

var log *zap.SugaredLogger

func InitLogger() {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = timestampKey
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var core zapcore.Core
	if os.Getenv(appEnvKey) == appEnvProduction {
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)
	} else {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
	}

	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}

// WithTraceID stores a trace ID in the context for use with *WithContext functions.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func Fatalf(format string, args ...any) {
	log.Fatalf(format, args...)
}

func Infof(format string, args ...any) {
	log.Infof(format, args...)
}

func InfofWithContext(ctx context.Context, format string, args ...any) {
	log.With(traceFields(ctx)...).Infof(format, args...)
}

func Errorf(format string, args ...any) {
	log.Errorf(format, args...)
}

func ErrorfWithContext(ctx context.Context, format string, args ...any) {
	log.With(traceFields(ctx)...).Errorf(format, args...)
}

func Debugf(format string, args ...any) {
	log.Debugf(format, args...)
}

func DebugfWithContext(ctx context.Context, format string, args ...any) {
	log.With(traceFields(ctx)...).Debugf(format, args...)
}

func Warnf(format string, args ...any) {
	log.Warnf(format, args...)
}

func WarnfWithContext(ctx context.Context, format string, args ...any) {
	log.With(traceFields(ctx)...).Warnf(format, args...)
}

func traceFields(ctx context.Context) []any {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		return []any{string(traceIDKey), traceID}
	}
	return nil
}
