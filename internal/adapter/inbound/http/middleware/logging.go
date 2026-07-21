package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	logContext "github.com/digitalrealmforgestudios/d-logger/context"
	logOption "github.com/digitalrealmforgestudios/d-logger/option"

	"idas-video/internal/adapter/inbound/http/observability"
)

func RequestLogging(next http.Handler) http.Handler {
	log := observability.Child("http.middleware")
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		requestID := request.Header.Get("X-Request-ID")
		if strings.TrimSpace(requestID) == "" {
			requestID = fmt.Sprintf("req-%d", startedAt.UnixNano())
		}

		ctx := logContext.SetRequestId(request.Context(), requestID)
		request = request.WithContext(ctx)
		writer.Header().Set("X-Request-ID", requestID)

		log.Info(
			"request started",
			logOption.Context(ctx),
			logOption.EventName("http.request.started"),
			logOption.Attributes(map[string]interface{}{
				"http.method": request.Method,
				"http.path":   request.URL.Path,
				"http.query":  request.URL.RawQuery,
			}),
		)

		next.ServeHTTP(writer, request)

		log.Info(
			"request completed",
			logOption.Context(ctx),
			logOption.EventName("http.request.completed"),
			logOption.Duration(time.Since(startedAt)),
			logOption.Attributes(map[string]interface{}{
				"http.method": request.Method,
				"http.path":   request.URL.Path,
			}),
		)
	})
}
