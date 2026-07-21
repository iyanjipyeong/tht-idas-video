package handler

import (
	"bytes"
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

type shapeLoginUsecase struct{}

func (shapeLoginUsecase) Login(context.Context, string, string) (*entity.AuthUser, error) {
	return &entity.AuthUser{
		ID:           entity.UUID("11111111-1111-4111-8111-111111111111"),
		Email:        "demo@example.com",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiredAt:    time.Unix(1784714400, 0).UTC(),
	}, nil
}

func (shapeLoginUsecase) AuthenticateAccessToken(context.Context, string) (entity.UUID, error) {
	return "", nil
}

func (shapeLoginUsecase) AuthenticateRefreshToken(context.Context, string) (entity.UUID, error) {
	return "", nil
}

func (shapeLoginUsecase) Refresh(context.Context, string) (*entity.AuthUser, error) {
	return nil, nil
}

type shapePaymentUsecase struct{}

func (shapePaymentUsecase) ProcessPaymentCallback(context.Context, inbound.PaymentCallbackRequest) error {
	return nil
}

type shapeSubscriptionUsecase struct{}

func (shapeSubscriptionUsecase) GetActiveSubscription(context.Context, entity.UUID) (*entity.Subscription, error) {
	startedAt := time.Unix(1784628000, 0).UTC()
	return &entity.Subscription{
		ID:       entity.UUID("55555555-5555-4555-8555-555555555555"),
		UserID:   entity.UUID("11111111-1111-4111-8111-111111111111"),
		TierID:   entity.UUID("11111111-1111-4111-8111-111111111103"),
		TierCode: entity.TierGold,
		TierSnapshot: entity.TierSnapshot{
			TierID:       entity.UUID("11111111-1111-4111-8111-111111111103"),
			TierCode:     entity.TierGold,
			TierName:     "Gold",
			TierLevel:    3,
			TierPrice:    150000,
			TierCurrency: "IDR",
		},
		Status:    entity.SubscriptionStatusActive,
		StartDate: startedAt,
		EndDate:   startedAt.AddDate(0, 0, 30),
		CreatedAt: startedAt,
		UpdatedAt: startedAt,
	}, nil
}

func (shapeSubscriptionUsecase) CreateSubscriptionTransaction(context.Context, entity.UUID, entity.UUID, entity.SubscriptionAction, int) (*entity.Transaction, error) {
	createdAt := time.Unix(1784628000, 0).UTC()
	currentEndDate := createdAt.AddDate(0, 0, 7)
	return &entity.Transaction{
		ID:                    entity.UUID("77777777-7777-4777-8777-777777777777"),
		ExternalTransactionID: "sub-11111111-1111-4111-8111-111111111111-1784628000000000000",
		OrderID:               "ORDER-SUB-1784628000000000000",
		UserID:                entity.UUID("11111111-1111-4111-8111-111111111111"),
		TierID:                entity.UUID("11111111-1111-4111-8111-111111111103"),
		TierCode:              entity.TierGold,
		TierSnapshot:          entity.TierSnapshot{TierID: entity.UUID("11111111-1111-4111-8111-111111111103"), TierCode: entity.TierGold, TierName: "Gold", TierLevel: 3, TierPrice: 150000, TierCurrency: "IDR"},
		SubscriptionAction:    entity.SubscriptionActionNew,
		SubscriptionDays:      30,
		CurrentSubscriptionID: entity.UUID("55555555-5555-4555-8555-555555555555"),
		CurrentTierSnapshot:   &entity.TierSnapshot{TierID: entity.UUID("11111111-1111-4111-8111-111111111102"), TierCode: entity.TierSilver, TierName: "Silver", TierLevel: 2, TierPrice: 100000, TierCurrency: "IDR"},
		CurrentEndDate:        &currentEndDate,
		ProratedCredit:        25000,
		FinalAmount:           150000,
		GrossAmount:           150000,
		Currency:              "IDR",
		TransactionStatus:     entity.TransactionStatusPending,
		PaymentStatus:         entity.PaymentStatusPending,
		CreatedAt:             createdAt,
		UpdatedAt:             createdAt,
	}, nil
}

type shapeTransactionUsecase struct{}

func (shapeTransactionUsecase) ListTransactionsByUserID(context.Context, entity.UUID) ([]entity.Transaction, error) {
	createdAt := time.Unix(1784628000, 0).UTC()
	return []entity.Transaction{{
		ID:                    entity.UUID("77777777-7777-4777-8777-777777777777"),
		ExternalTransactionID: "trx-001",
		GatewayTransactionID:  "gw-trx-001",
		OrderID:               "ORDER-001",
		UserID:                entity.UUID("11111111-1111-4111-8111-111111111111"),
		TierID:                entity.UUID("11111111-1111-4111-8111-111111111103"),
		TierCode:              entity.TierGold,
		TierSnapshot:          entity.TierSnapshot{TierID: entity.UUID("11111111-1111-4111-8111-111111111103"), TierCode: entity.TierGold, TierName: "Gold", TierLevel: 3, TierPrice: 150000, TierCurrency: "IDR"},
		SubscriptionAction:    entity.SubscriptionActionNew,
		SubscriptionDays:      30,
		FinalAmount:           150000,
		GrossAmount:           150000,
		Currency:              "IDR",
		TransactionStatus:     entity.TransactionStatusProcessed,
		PaymentStatus:         entity.PaymentStatusPaid,
		CreatedAt:             createdAt,
		UpdatedAt:             createdAt,
	}}, nil
}

func (shapeTransactionUsecase) GetTransactionByID(context.Context, entity.UUID, entity.UUID) (*entity.Transaction, error) {
	createdAt := time.Unix(1784628000, 0).UTC()
	return &entity.Transaction{
		ID:                    entity.UUID("77777777-7777-4777-8777-777777777777"),
		ExternalTransactionID: "trx-001",
		GatewayTransactionID:  "gw-trx-001",
		OrderID:               "ORDER-001",
		UserID:                entity.UUID("11111111-1111-4111-8111-111111111111"),
		TierID:                entity.UUID("11111111-1111-4111-8111-111111111103"),
		TierCode:              entity.TierGold,
		TierSnapshot:          entity.TierSnapshot{TierID: entity.UUID("11111111-1111-4111-8111-111111111103"), TierCode: entity.TierGold, TierName: "Gold", TierLevel: 3, TierPrice: 150000, TierCurrency: "IDR"},
		SubscriptionAction:    entity.SubscriptionActionNew,
		SubscriptionDays:      30,
		FinalAmount:           150000,
		GrossAmount:           150000,
		Currency:              "IDR",
		TransactionStatus:     entity.TransactionStatusProcessed,
		PaymentStatus:         entity.PaymentStatusPaid,
		CreatedAt:             createdAt,
		UpdatedAt:             createdAt,
	}, nil
}

type shapeVideoUsecase struct {
	videos       []entity.Video
	canAccessErr error
	getVideoErr  error
}

func (usecase shapeVideoUsecase) ListAccessibleVideos(context.Context, entity.UUID, inbound.VideoListQuery) (*inbound.VideoListResult, error) {
	return &inbound.VideoListResult{Items: usecase.videos, Total: len(usecase.videos), Page: 1, Offset: 0, SortBy: "createdAtDesc"}, nil
}

func (usecase shapeVideoUsecase) GetAccessibleVideoByID(context.Context, entity.UUID, entity.UUID) (*entity.Video, error) {
	if usecase.getVideoErr != nil {
		return nil, usecase.getVideoErr
	}
	videoTime := time.Unix(1784628000, 0).UTC()
	return &entity.Video{
		ID:          entity.UUID("44444444-4444-4444-8444-444444444444"),
		Title:       "Gold Video",
		Description: "Video for Gold tier",
		Category:    entity.TierGold,
		VideoURL:    "https://example.com/gold.mp4",
		CreatedAt:   videoTime,
		UpdatedAt:   videoTime,
	}, nil
}

func (usecase shapeVideoUsecase) CanAccessVideo(context.Context, entity.UUID, entity.UUID) error {
	return usecase.canAccessErr
}

func TestHealthHandlerReturnsSuccessEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	handler := &HealthHandler{buildID: "development", buildTime: "2026-07-22T09:00:00Z", startedAt: time.Now().Add(-5 * time.Minute)}
	handler.ServeHTTP(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	data := payload["data"].(map[string]any)
	if data["buildId"] != "development" || data["buildTime"] != "2026-07-22T09:00:00Z" {
		t.Fatalf("health data = %#v", data)
	}
	if _, ok := data["uptime"].(string); !ok {
		t.Fatalf("health uptime = %#v", data["uptime"])
	}
}

func TestAuthHandlerLoginReturnsSuccessEnvelope(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"demo@example.com","password":"password"}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/login", body)

	NewAuthHandler(shapeLoginUsecase{}).Login(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	data := payload["data"].(map[string]any)
	if data["accessToken"] != "access-token" || data["refreshToken"] != "refresh-token" || data["expiredAt"] != float64(1784714400) {
		t.Fatalf("auth data = %#v", data)
	}
}

func TestPaymentHandlerCallbackReturnsSuccessEnvelope(t *testing.T) {
	body := bytes.NewBufferString(`{"transactionId":"trx-001","userId":"11111111-1111-4111-8111-111111111111","tier":"Gold","paymentStatus":"paid","subscriptionDays":30}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payment/callback", body)

	NewPaymentHandler(shapePaymentUsecase{}).Callback(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	if payload["data"] != nil {
		t.Fatalf("payment data = %#v, want nil", payload["data"])
	}
}

func TestSubscriptionHandlerGetActiveReturnsSuccessEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions/active", nil)
	request.Header.Set("Authorization", "Bearer valid-token")

	middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(NewSubscriptionHandler(shapeSubscriptionUsecase{}).GetActive)).ServeHTTP(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	data := payload["data"].(map[string]any)
	if data["tierCode"] != "Gold" || data["tierName"] != "Gold" || data["tierLevel"] != float64(3) {
		t.Fatalf("subscription data = %#v", data)
	}
	if data["startDate"] != float64(1784628000) {
		t.Fatalf("subscription timestamps = %#v", data)
	}
}

func TestSubscriptionHandlerSubscribeReturnsPendingTransactionEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBufferString(`{"tierId":"11111111-1111-4111-8111-111111111103","subscriptionAction":"new","subscriptionDays":30}`))
	request.Header.Set("Authorization", "Bearer valid-token")

	middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(NewSubscriptionHandler(shapeSubscriptionUsecase{}).Subscribe)).ServeHTTP(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	data := payload["data"].(map[string]any)
	if data["tierCode"] != "Gold" || data["transactionStatus"] != "pending" || data["paymentStatus"] != "pending" || data["currentTierCode"] != "Silver" || data["proratedCredit"] != float64(25000) {
		t.Fatalf("subscribe data = %#v", data)
	}
}

func TestTransactionHandlerListReturnsListEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	request.Header.Set("Authorization", "Bearer valid-token")

	middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(NewTransactionHandler(shapeTransactionUsecase{}).ListTransactions)).ServeHTTP(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	data := payload["data"].(map[string]any)
	items := data["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items = %#v", items)
	}
}

func TestTransactionHandlerGetDetailReturnsSuccessEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/transactions/77777777-7777-4777-8777-777777777777", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	mux := http.NewServeMux()
	mux.Handle("GET /transactions/{id}", middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(NewTransactionHandler(shapeTransactionUsecase{}).GetTransaction)))

	mux.ServeHTTP(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	data := payload["data"].(map[string]any)
	if data["externalTransactionId"] != "trx-001" || data["userId"] != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("transaction data = %#v", data)
	}
}

func TestTransactionHandlerGetTransactionRejectsOtherUserAccess(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/transactions/77777777-7777-4777-8777-777777777777", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	mux := http.NewServeMux()
	mux.Handle("GET /transactions/{id}", middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(NewTransactionHandler(transactionDeniedUsecase{}).GetTransaction)))

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

type transactionDeniedUsecase struct{}

func (transactionDeniedUsecase) ListTransactionsByUserID(context.Context, entity.UUID) ([]entity.Transaction, error) {
	return nil, nil
}

func (transactionDeniedUsecase) GetTransactionByID(context.Context, entity.UUID, entity.UUID) (*entity.Transaction, error) {
	return nil, entity.ErrTransactionNotFound
}

func TestVideoHandlerListReturnsListEnvelope(t *testing.T) {
	videoTime := time.Unix(1784628000, 0).UTC()
	usecase := shapeVideoUsecase{videos: []entity.Video{{
		ID:        entity.UUID("22222222-2222-4222-8222-222222222222"),
		Title:     "Bronze Video",
		Category:  entity.TierBronze,
		VideoURL:  "https://example.com/bronze.mp4",
		CreatedAt: videoTime,
		UpdatedAt: videoTime,
	}}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/videos?page=2&offset=1&sort=titleAsc", nil)
	request.Header.Set("Authorization", "Bearer valid-token")

	middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(NewVideoHandler(usecase).ListVideos)).ServeHTTP(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	data := payload["data"].(map[string]any)
	metadata := data["metadata"].(map[string]any)
	if metadata["total"] != float64(1) || metadata["page"] != float64(1) || metadata["offset"] != float64(0) || metadata["sortBy"] != "createdAtDesc" {
		t.Fatalf("metadata = %#v", metadata)
	}
	items := data["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items length = %d, want 1", len(items))
	}
}

func TestVideoHandlerGetVideoReturnsDetailEnvelope(t *testing.T) {
	usecase := shapeVideoUsecase{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/videos/44444444-4444-4444-8444-444444444444", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	mux := http.NewServeMux()
	mux.Handle("GET /videos/{id}", middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(NewVideoHandler(usecase).GetVideo)))

	mux.ServeHTTP(response, request)

	payload := decodeResponse(t, response)
	assertEnvelope(t, payload, true, "200", "Successfully")
	data := payload["data"].(map[string]any)
	if data["id"] != "44444444-4444-4444-8444-444444444444" || data["videoUrl"] != "https://example.com/gold.mp4" {
		t.Fatalf("video data = %#v", data)
	}
	if data["createdAt"] != float64(1784628000) || data["updatedAt"] != float64(1784628000) {
		t.Fatalf("video timestamps = %#v", data)
	}
}

func TestVideoHandlerGetVideoReturnsUnauthorizedEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/videos/44444444-4444-4444-8444-444444444444", nil)
	mux := http.NewServeMux()
	mux.Handle("GET /videos/{id}", middleware.AuthMiddleware(fakeLoginUsecase{})(http.HandlerFunc(NewVideoHandler(shapeVideoUsecase{}).GetVideo)))

	mux.ServeHTTP(response, request)

	payload := decodeErrorResponse(t, response, http.StatusUnauthorized)
	assertEnvelope(t, payload, false, "E_AUTH_001", entity.ErrUnauthorized.Error())
	if payload["data"] != nil {
		t.Fatalf("unauthorized data = %#v", payload["data"])
	}
}

func TestVideoHandlerGetVideoReturnsForbiddenEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/videos/44444444-4444-4444-8444-444444444444", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	mux := http.NewServeMux()
	usecase := shapeVideoUsecase{canAccessErr: entity.ErrForbiddenTier}
	mux.Handle("GET /videos/{id}", middleware.AuthMiddleware(fakeLoginUsecase{})(middleware.VideoAccess(usecase)(http.HandlerFunc(NewVideoHandler(usecase).GetVideo))))

	mux.ServeHTTP(response, request)

	payload := decodeErrorResponse(t, response, http.StatusForbidden)
	assertEnvelope(t, payload, false, "E_ACCESS_001", entity.ErrForbiddenTier.Error())
}

func TestVideoHandlerGetVideoReturnsNotFoundEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/videos/44444444-4444-4444-8444-444444444444", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	mux := http.NewServeMux()
	usecase := shapeVideoUsecase{canAccessErr: entity.ErrVideoNotFound}
	mux.Handle("GET /videos/{id}", middleware.AuthMiddleware(fakeLoginUsecase{})(middleware.VideoAccess(usecase)(http.HandlerFunc(NewVideoHandler(usecase).GetVideo))))

	mux.ServeHTTP(response, request)

	payload := decodeErrorResponse(t, response, http.StatusNotFound)
	assertEnvelope(t, payload, false, "E_VIDEO_001", entity.ErrVideoNotFound.Error())
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

func decodeErrorResponse(t *testing.T, response *httptest.ResponseRecorder, wantStatus int) map[string]any {
	t.Helper()
	if response.Code != wantStatus {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, wantStatus, response.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

func assertEnvelope(t *testing.T, payload map[string]any, success bool, code string, message string) {
	t.Helper()
	if payload["success"] != success || payload["code"] != code || payload["message"] != message {
		t.Fatalf("envelope = %#v", payload)
	}
}
