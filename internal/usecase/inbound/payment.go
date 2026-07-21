package inbound

import (
	"context"
	"time"

	"idas-video/internal/entity"
)

type PaymentCallbackRequest struct {
	TransactionID        string
	GatewayTransactionID string
	OrderID              string
	UserID               entity.UUID
	Tier                 entity.Tier
	PaymentType          string
	PaymentStatus        string
	FraudStatus          string
	GrossAmount          float64
	Currency             string
	SignatureKey         string
	TransactionTime      *time.Time
	SettlementTime       *time.Time
	ExpiryTime           *time.Time
	SettlementTimeRaw    string
	SubscriptionDays     int
	RawPayload           []byte
}

type PaymentCallbackProcessor interface {
	ProcessPaymentCallback(ctx context.Context, payload PaymentCallbackRequest) error
}
