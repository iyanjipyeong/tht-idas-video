package middleware

import (
	"net/http"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

func VideoAccess(authorizer inbound.VideoAccessUsecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			userID, ok := UserIDFromContext(request.Context())
			if !ok {
				writeError(writer, http.StatusUnauthorized, entity.ErrUnauthorized)
				return
			}

			videoID := entity.UUID(request.PathValue("id"))
			if err := authorizer.CanAccessVideo(request.Context(), userID, videoID); err != nil {
				statusCode := http.StatusForbidden
				if err == entity.ErrVideoNotFound {
					statusCode = http.StatusNotFound
				}
				writeError(writer, statusCode, err)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}
}
