package outbound

import "context"

type LogField struct {
	Key   string
	Value any
}

type Logger interface {
	Info(ctx context.Context, namespace string, message string, event string, fields ...LogField)
	Warn(ctx context.Context, namespace string, message string, event string, fields ...LogField)
	Error(ctx context.Context, namespace string, message string, event string, err error, fields ...LogField)
}
