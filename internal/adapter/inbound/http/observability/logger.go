package observability

import (
	"context"

	dlogger "github.com/digitalrealmforgestudios/d-logger"
	logOption "github.com/digitalrealmforgestudios/d-logger/option"
)

func Child(namespace string) dlogger.Logger {
	return dlogger.Get().NewChild(logOption.WithNamespace(namespace))
}

func WithContext(ctx context.Context) logOption.SetterFunc {
	return logOption.Context(ctx)
}
