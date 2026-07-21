package middleware

import (
	"context"
	"net/http"
	"strings"

	logOption "github.com/digitalrealmforgestudios/d-logger/option"

	"idas-video/internal/adapter/inbound/http/observability"
	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func AuthMiddleware(authenticator inbound.LoginUsecase) func(http.Handler) http.Handler {
	log := observability.Child("http.middleware.auth")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			authorization := request.Header.Get("Authorization")
			if authorization == entity.EmptyString || !strings.HasPrefix(authorization, entity.AuthorizationBearer) {
				log.Warn("authorization header missing or invalid", observability.WithContext(request.Context()))
				writeError(writer, http.StatusUnauthorized, entity.ErrUnauthorized)
				return
			}

			accessToken := strings.TrimSpace(strings.TrimPrefix(authorization, entity.AuthorizationBearer))
			userID, err := authenticator.AuthenticateAccessToken(request.Context(), accessToken)
			if err != nil {
				log.Warn("access token authentication failed", observability.WithContext(request.Context()), logOption.Error(err))
				writeError(writer, http.StatusUnauthorized, entity.ErrUnauthorized)
				return
			}
			if !entity.IsUUID(userID.String()) {
				log.Warn("authenticated user id invalid", observability.WithContext(request.Context()), logOption.Attribute("user.id", userID.String()))
				writeError(writer, http.StatusUnauthorized, entity.ErrUnauthorized)
				return
			}

			ctx := context.WithValue(request.Context(), userIDContextKey, userID)
			log.Info("request authenticated", observability.WithContext(ctx), logOption.Attribute("user.id", userID.String()))
			next.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}

func Auth(authenticator inbound.LoginUsecase) func(http.Handler) http.Handler {
	return AuthMiddleware(authenticator)
}

func UserIDFromContext(ctx context.Context) (entity.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(entity.UUID)
	return userID, ok && userID != entity.EmptyString
}
