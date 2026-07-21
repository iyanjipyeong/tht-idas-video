package usecase

import (
	"context"

	"idas-video/internal/usecase/outbound"
)

type noopLogger struct{}

func (noopLogger) Info(context.Context, string, string, string, ...outbound.LogField) {}

func (noopLogger) Warn(context.Context, string, string, string, ...outbound.LogField) {}

func (noopLogger) Error(context.Context, string, string, string, error, ...outbound.LogField) {}

func fallbackLogger(log outbound.Logger) outbound.Logger {
	if log == nil {
		return noopLogger{}
	}
	return log
}
