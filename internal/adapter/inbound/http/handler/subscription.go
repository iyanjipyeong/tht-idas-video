package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"idas-video/internal/adapter/inbound/http/middleware"
	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

type activeSubscriptionResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"userId"`
	TierID       string  `json:"tierId"`
	TierCode     string  `json:"tierCode"`
	TierName     string  `json:"tierName"`
	TierLevel    int     `json:"tierLevel"`
	TierPrice    float64 `json:"tierPrice"`
	TierCurrency string  `json:"tierCurrency"`
	Status       string  `json:"status"`
	StartDate    int64   `json:"startDate"`
	EndDate      int64   `json:"endDate"`
	CreatedAt    int64   `json:"createdAt"`
	UpdatedAt    int64   `json:"updatedAt"`
}

type SubscriptionHandler struct {
	usecase inbound.SubscriptionPurchaseUsecase
}

type subscribeTierRequest struct {
	TierID             string `json:"tierId"`
	SubscriptionAction string `json:"subscriptionAction"`
	SubscriptionDays   int    `json:"subscriptionDays"`
}

type subscribeTierResponse struct {
	ID                    string   `json:"id"`
	ExternalTransactionID string   `json:"externalTransactionId"`
	OrderID               string   `json:"orderId"`
	PaymentReference      string   `json:"paymentReference"`
	PaymentChannel        string   `json:"paymentChannel"`
	PaymentInstructions   []string `json:"paymentInstructions"`
	ExpiresAt             int64    `json:"expiresAt"`
	UserID                string   `json:"userId"`
	TierID                string   `json:"tierId"`
	TierCode              string   `json:"tierCode"`
	TierName              string   `json:"tierName"`
	TierLevel             int      `json:"tierLevel"`
	TierPrice             float64  `json:"tierPrice"`
	TierCurrency          string   `json:"tierCurrency"`
	RequestedAction       string   `json:"requestedAction"`
	EffectiveAction       string   `json:"effectiveAction"`
	SubscriptionAction    string   `json:"subscriptionAction"`
	SubscriptionDays      int      `json:"subscriptionDays"`
	CurrentSubscriptionID string   `json:"currentSubscriptionId"`
	CurrentTierID         string   `json:"currentTierId"`
	CurrentTierCode       string   `json:"currentTierCode"`
	CurrentTierName       string   `json:"currentTierName"`
	CurrentTierLevel      int      `json:"currentTierLevel"`
	CurrentTierPrice      float64  `json:"currentTierPrice"`
	CurrentEndDate        int64    `json:"currentEndDate"`
	ProratedCredit        float64  `json:"proratedCredit"`
	TransactionStatus     string   `json:"transactionStatus"`
	PaymentStatus         string   `json:"paymentStatus"`
	FinalAmount           float64  `json:"finalAmount"`
	GrossAmount           float64  `json:"grossAmount"`
	Currency              string   `json:"currency"`
	CreatedAt             int64    `json:"createdAt"`
	UpdatedAt             int64    `json:"updatedAt"`
}

func NewSubscriptionHandler(usecase inbound.SubscriptionPurchaseUsecase) *SubscriptionHandler {
	return &SubscriptionHandler{usecase: usecase}
}

func (handler *SubscriptionHandler) Subscribe(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		writeError(writer, ErrHTTPUnauthorized)
		return
	}

	var payload subscribeTierRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeError(writer, entity.ErrInvalidRequest)
		return
	}

	if !entity.IsUUID(payload.TierID) {
		writeError(writer, entity.ErrInvalidSubscription)
		return
	}

	action := entity.SubscriptionAction(payload.SubscriptionAction)
	if action == "" {
		action = entity.SubscriptionActionNew
	}
	transaction, err := handler.usecase.CreateSubscriptionTransaction(request.Context(), userID, entity.UUID(payload.TierID), action, payload.SubscriptionDays)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeSuccess(writer, newSubscribeTierResponse(*transaction))
}

func (handler *SubscriptionHandler) GetActive(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		writeError(writer, ErrHTTPUnauthorized)
		return
	}

	subscription, err := handler.usecase.GetActiveSubscription(request.Context(), userID)
	if err != nil {
		if errors.Is(err, entity.ErrActiveSubscription) {
			writeError(writer, entity.ErrNoActiveSubscription)
			return
		}
		writeError(writer, err)
		return
	}

	writeSuccess(writer, newActiveSubscriptionResponse(*subscription))
}

func newSubscribeTierResponse(transaction entity.Transaction) subscribeTierResponse {
	response := subscribeTierResponse{
		ID:                    transaction.ID.String(),
		ExternalTransactionID: transaction.ExternalTransactionID,
		OrderID:               transaction.OrderID,
		PaymentReference:      transaction.ExternalTransactionID,
		PaymentChannel:        "manual-simulated-gateway",
		PaymentInstructions: []string{
			"Review the transaction reference before sending callback payload.",
			"Use /payment/callback with paymentStatus=paid to simulate a successful notification.",
			"Send the same transactionId to simulate idempotent gateway retries.",
		},
		ExpiresAt:          unixSecondsValue(transaction.ExpiryTime),
		UserID:             transaction.UserID.String(),
		TierID:             transaction.TierID.String(),
		TierCode:           transaction.TierCode.String(),
		TierName:           transaction.TierSnapshot.TierName,
		TierLevel:          transaction.TierSnapshot.TierLevel,
		TierPrice:          transaction.TierSnapshot.TierPrice,
		TierCurrency:       transaction.TierSnapshot.TierCurrency,
		RequestedAction:    string(transaction.SubscriptionAction),
		EffectiveAction:    string(transaction.SubscriptionAction),
		SubscriptionAction: string(transaction.SubscriptionAction),
		SubscriptionDays:   transaction.SubscriptionDays,
		ProratedCredit:     transaction.ProratedCredit,
		TransactionStatus:  string(transaction.TransactionStatus),
		PaymentStatus:      transaction.PaymentStatus,
		FinalAmount:        transaction.FinalAmount,
		GrossAmount:        transaction.GrossAmount,
		Currency:           transaction.Currency,
		CreatedAt:          transaction.CreatedAt.UTC().Unix(),
		UpdatedAt:          transaction.UpdatedAt.UTC().Unix(),
	}
	if transaction.CurrentSubscriptionID != entity.EmptyString {
		response.CurrentSubscriptionID = transaction.CurrentSubscriptionID.String()
	}
	if transaction.CurrentTierSnapshot != nil {
		response.CurrentTierID = transaction.CurrentTierSnapshot.TierID.String()
		response.CurrentTierCode = transaction.CurrentTierSnapshot.TierCode.String()
		response.CurrentTierName = transaction.CurrentTierSnapshot.TierName
		response.CurrentTierLevel = transaction.CurrentTierSnapshot.TierLevel
		response.CurrentTierPrice = transaction.CurrentTierSnapshot.TierPrice
	}
	if transaction.CurrentEndDate != nil {
		response.CurrentEndDate = transaction.CurrentEndDate.UTC().Unix()
	}
	return response
}

func unixSecondsValue(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	return value.UTC().Unix()
}

func newActiveSubscriptionResponse(subscription entity.Subscription) activeSubscriptionResponse {
	return activeSubscriptionResponse{
		ID:           subscription.ID.String(),
		UserID:       subscription.UserID.String(),
		TierID:       subscription.TierID.String(),
		TierCode:     subscription.TierCode.String(),
		TierName:     subscription.TierSnapshot.TierName,
		TierLevel:    subscription.TierSnapshot.TierLevel,
		TierPrice:    subscription.TierSnapshot.TierPrice,
		TierCurrency: subscription.TierSnapshot.TierCurrency,
		Status:       subscription.Status.String(),
		StartDate:    subscription.StartDate.UTC().Unix(),
		EndDate:      subscription.EndDate.UTC().Unix(),
		CreatedAt:    subscription.CreatedAt.UTC().Unix(),
		UpdatedAt:    subscription.UpdatedAt.UTC().Unix(),
	}
}
