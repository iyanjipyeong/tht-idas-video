package logger

import (
	"context"
	"os"

	dlogger "github.com/digitalrealmforgestudios/d-logger"
	"github.com/digitalrealmforgestudios/d-logger/level"
	logOption "github.com/digitalrealmforgestudios/d-logger/option"

	"idas-video/internal/usecase/outbound"
)

const defaultNamespace = "idas-video"
const defaultLevel = "info"

func Init() dlogger.Logger {
	_ = os.Setenv(dlogger.EnvServiceName, defaultNamespace)
	log := dlogger.NewStdLogger(
		dlogger.NewPrinter(os.Stdout),
		logOption.Level(level.Parse(envOrDefault(dlogger.EnvLogLevel, defaultLevel))),
		logOption.WithNamespace(envOrDefault(dlogger.EnvLogNamespace, defaultNamespace)),
	)
	dlogger.Register(log)
	return dlogger.Get()
}

func Child(namespace string) dlogger.Logger {
	return dlogger.Get().NewChild(logOption.WithNamespace(namespace))
}

func WithContext(ctx context.Context) logOption.SetterFunc {
	return logOption.Context(ctx)
}

type UsecaseLogger struct{}

func NewUsecaseLogger() UsecaseLogger {
	return UsecaseLogger{}
}

func (UsecaseLogger) Info(ctx context.Context, namespace string, message string, event string, fields ...outbound.LogField) {
	Child(namespace).Info(message, logOptions(ctx, event, nil, fields)...)
}

func (UsecaseLogger) Warn(ctx context.Context, namespace string, message string, event string, fields ...outbound.LogField) {
	Child(namespace).Warn(message, logOptions(ctx, event, nil, fields)...)
}

func (UsecaseLogger) Error(ctx context.Context, namespace string, message string, event string, err error, fields ...outbound.LogField) {
	Child(namespace).Error(message, logOptions(ctx, event, err, fields)...)
}

func logOptions(ctx context.Context, event string, err error, fields []outbound.LogField) []logOption.SetterFunc {
	options := []logOption.SetterFunc{WithContext(ctx)}
	if event != "" {
		options = append(options, logOption.EventName(event))
	}
	if err != nil {
		options = append(options, logOption.Error(err))
	}
	for _, field := range fields {
		options = append(options, logOption.Attribute(field.Key, field.Value))
	}
	return options
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
