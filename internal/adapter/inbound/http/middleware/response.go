package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"idas-video/internal/entity"
)

type errorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

const (
	responseCodeAuth    = "E_AUTH_001"
	responseCodeAccess  = "E_ACCESS_001"
	responseCodeVideo   = "E_VIDEO_001"
	responseCodeRequest = "E_REQUEST_001"
)

func writeError(writer http.ResponseWriter, statusCode int, err error) {
	writer.Header().Set(entity.HTTPHeaderContentType, entity.HTTPContentTypeJSON)
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(errorResponse{Success: false, Code: errorCode(err), Message: err.Error(), Data: nil})
}

func errorCode(err error) string {
	switch {
	case errors.Is(err, entity.ErrUnauthorized), errors.Is(err, entity.ErrInvalidCredentials):
		return responseCodeAuth
	case errors.Is(err, entity.ErrActiveSubscription), errors.Is(err, entity.ErrForbiddenTier):
		return responseCodeAccess
	case errors.Is(err, entity.ErrVideoNotFound):
		return responseCodeVideo
	default:
		return responseCodeRequest
	}
}
