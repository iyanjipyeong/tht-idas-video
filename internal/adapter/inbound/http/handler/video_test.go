package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idas-video/internal/adapter/inbound/http/middleware"
	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

type fakeVideoAccessUsecase struct{}

func (fakeVideoAccessUsecase) ListAccessibleVideos(context.Context, entity.UUID, inbound.VideoListQuery) (*inbound.VideoListResult, error) {
	return &inbound.VideoListResult{}, nil
}

func (fakeVideoAccessUsecase) GetAccessibleVideoByID(context.Context, entity.UUID, entity.UUID) (*entity.Video, error) {
	return nil, nil
}

func (fakeVideoAccessUsecase) CanAccessVideo(context.Context, entity.UUID, entity.UUID) error {
	return nil
}

type fakeLoginUsecase struct{}

func (fakeLoginUsecase) Login(context.Context, string, string) (*entity.AuthUser, error) {
	return nil, nil
}

func (fakeLoginUsecase) AuthenticateAccessToken(context.Context, string) (entity.UUID, error) {
	return entity.UUID("11111111-1111-4111-8111-111111111111"), nil
}

func (fakeLoginUsecase) AuthenticateRefreshToken(context.Context, string) (entity.UUID, error) {
	return "", nil
}

func (fakeLoginUsecase) Refresh(context.Context, string) (*entity.AuthUser, error) {
	return &entity.AuthUser{ExpiredAt: time.Now()}, nil
}

func TestVideoHandlerGetVideoRejectsInvalidUUID(t *testing.T) {
	handler := NewVideoHandler(fakeVideoAccessUsecase{})
	mux := http.NewServeMux()
	mux.Handle("GET /videos/{id}", middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(handler.GetVideo)))

	request := httptest.NewRequest(http.MethodGet, "/videos/not-a-uuid", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("GET /videos/{id} status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if payload["success"] != false || payload["code"] != "E_REQUEST_001" || payload["message"] != entity.ErrInvalidRequest.Error() {
		t.Fatalf("error response = %#v", payload)
	}
}
