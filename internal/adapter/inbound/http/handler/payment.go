package handler

import (
	"encoding/json"
	"net/http"
	"time"

	logOption "github.com/digitalrealmforgestudios/d-logger/option"

	"idas-video/internal/adapter/inbound/http/observability"
	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

type paymentCallbackPayload struct {
	TransactionID        string      `json:"transactionId"`
	GatewayTransactionID string      `json:"gatewayTransactionId"`
	OrderID              string      `json:"orderId"`
	UserID               string      `json:"userId"`
	Tier                 entity.Tier `json:"tier"`
	PaymentType          string      `json:"paymentType"`
	PaymentStatus        string      `json:"paymentStatus"`
	FraudStatus          string      `json:"fraudStatus"`
	GrossAmount          float64     `json:"grossAmount"`
	Currency             string      `json:"currency"`
	SignatureKey         string      `json:"signatureKey"`
	TransactionTime      string      `json:"transactionTime"`
	SettlementTime       string      `json:"settlementTime"`
	ExpiryTime           string      `json:"expiryTime"`
	SubscriptionDays     int         `json:"subscriptionDays"`
}

type PaymentHandler struct {
	usecase inbound.PaymentCallbackProcessor
}

func NewPaymentHandler(usecase inbound.PaymentCallbackProcessor) *PaymentHandler {
	return &PaymentHandler{usecase: usecase}
}

func (handler *PaymentHandler) Callback(writer http.ResponseWriter, request *http.Request) {
	log := observability.Child("http.handler.payment")
	log.Info("payment callback received", observability.WithContext(request.Context()), logOption.EventName("http.payment.callback.started"))

	var payload paymentCallbackPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		log.Warn("payment callback payload invalid", observability.WithContext(request.Context()), logOption.Error(err))
		writeError(writer, entity.ErrInvalidRequest)
		return
	}

	if !entity.IsUUID(payload.UserID) {
		log.Warn("payment callback user id invalid", observability.WithContext(request.Context()), logOption.Attribute("user.id", payload.UserID))
		writeError(writer, entity.ErrInvalidRequest)
		return
	}

	if err := handler.usecase.ProcessPaymentCallback(request.Context(), inbound.PaymentCallbackRequest{
		TransactionID:        payload.TransactionID,
		GatewayTransactionID: payload.GatewayTransactionID,
		OrderID:              payload.OrderID,
		UserID:               entity.UUID(payload.UserID),
		Tier:                 payload.Tier,
		PaymentType:          payload.PaymentType,
		PaymentStatus:        payload.PaymentStatus,
		FraudStatus:          payload.FraudStatus,
		GrossAmount:          payload.GrossAmount,
		Currency:             payload.Currency,
		SignatureKey:         payload.SignatureKey,
		TransactionTime:      parseCallbackTime(payload.TransactionTime),
		SettlementTime:       parseCallbackTime(payload.SettlementTime),
		ExpiryTime:           parseCallbackTime(payload.ExpiryTime),
		SettlementTimeRaw:    payload.SettlementTime,
		SubscriptionDays:     payload.SubscriptionDays,
		RawPayload:           mustMarshalPayload(payload),
	}); err != nil {
		log.Warn("payment callback processing failed", observability.WithContext(request.Context()), logOption.Error(err), logOption.Attribute("transaction.id", payload.TransactionID))
		writeError(writer, err)
		return
	}

	log.Info("payment callback succeeded", observability.WithContext(request.Context()), logOption.Attribute("transaction.id", payload.TransactionID))
	writeSuccess(writer, nil)
}

func mustMarshalPayload(payload paymentCallbackPayload) []byte {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return rawPayload
}

func parseCallbackTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}
	return nil
}
