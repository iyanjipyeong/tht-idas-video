package handler

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

type successResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type listResponseData struct {
	Items    any          `json:"items"`
	Metadata listMetadata `json:"metadata"`
}

type listMetadata struct {
	Total  int    `json:"total"`
	Page   int    `json:"page"`
	Offset int    `json:"offset"`
	SortBy string `json:"sortBy"`
}

const (
	responseCodeSuccess = "200"
	responseMessageOK   = "Successfully"
	responseCodeServer  = "E_SERVER_001"
	responseCodeAuth    = "E_AUTH_001"
	responseCodeAccess  = "E_ACCESS_001"
	responseCodeVideo   = "E_VIDEO_001"
	responseCodeUser    = "E_USER_001"
	responseCodeRequest = "E_REQUEST_001"
	responseCodePayment = "E_PAYMENT_001"
	defaultListPage     = 1
	defaultListOffset   = 0
	defaultListSortBy   = "createdAtDesc"
)

func writeJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set(entity.HTTPHeaderContentType, entity.HTTPContentTypeJSON)
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}

func writeSuccess(writer http.ResponseWriter, data any) {
	writeJSON(writer, http.StatusOK, successResponse{Success: true, Code: responseCodeSuccess, Message: responseMessageOK, Data: data})
}

func writeListSuccess(writer http.ResponseWriter, items any, total int, page int, offset int, sortBy string) {
	writeSuccess(writer, listResponseData{Items: items, Metadata: listMetadata{Total: total, Page: page, Offset: offset, SortBy: sortBy}})
}

func writeError(writer http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	message := err.Error()
	switch {
	case errors.Is(err, entity.ErrUnauthorized), errors.Is(err, entity.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
	case errors.Is(err, entity.ErrActiveSubscription), errors.Is(err, entity.ErrNoActiveSubscription), errors.Is(err, entity.ErrForbiddenTier):
		statusCode = http.StatusForbidden
	case errors.Is(err, entity.ErrVideoNotFound), errors.Is(err, entity.ErrUserNotFound), errors.Is(err, entity.ErrTransactionNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, entity.ErrInvalidRequest), errors.Is(err, entity.ErrInvalidSubscription), errors.Is(err, entity.ErrDuplicateTransaction), errors.Is(err, entity.ErrTransactionMismatch), errors.Is(err, entity.ErrTransactionProcessed), errors.Is(err, entity.ErrDowngradeTier), errors.Is(err, entity.ErrRepurchaseTier), errors.Is(err, entity.ErrSubscriptionActionRequiresActive), errors.Is(err, entity.ErrSubscriptionActionMismatch):
		statusCode = http.StatusBadRequest
	}
	if statusCode == http.StatusInternalServerError {
		message = "internal server error"
	}
	writeJSON(writer, statusCode, errorResponse{Success: false, Code: errorCode(err), Message: message, Data: nil})
}

func errorCode(err error) string {
	switch {
	case errors.Is(err, entity.ErrUnauthorized), errors.Is(err, entity.ErrInvalidCredentials):
		return responseCodeAuth
	case errors.Is(err, entity.ErrActiveSubscription), errors.Is(err, entity.ErrNoActiveSubscription), errors.Is(err, entity.ErrForbiddenTier):
		return responseCodeAccess
	case errors.Is(err, entity.ErrVideoNotFound):
		return responseCodeVideo
	case errors.Is(err, entity.ErrTransactionNotFound):
		return responseCodePayment
	case errors.Is(err, entity.ErrUserNotFound):
		return responseCodeUser
	case errors.Is(err, entity.ErrDuplicateTransaction), errors.Is(err, entity.ErrTransactionMismatch), errors.Is(err, entity.ErrTransactionProcessed):
		return responseCodePayment
	case errors.Is(err, entity.ErrInvalidRequest), errors.Is(err, entity.ErrInvalidSubscription), errors.Is(err, entity.ErrDowngradeTier), errors.Is(err, entity.ErrRepurchaseTier), errors.Is(err, entity.ErrSubscriptionActionRequiresActive), errors.Is(err, entity.ErrSubscriptionActionMismatch):
		return responseCodeRequest
	default:
		return responseCodeServer
	}
}
