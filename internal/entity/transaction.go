package entity

import "time"

type SubscriptionAction string

type TransactionStatus string

const (
	SubscriptionActionNew       SubscriptionAction = "new"
	SubscriptionActionRenew     SubscriptionAction = "renew"
	SubscriptionActionUpgrade   SubscriptionAction = "upgrade"
	SubscriptionActionDowngrade SubscriptionAction = "downgrade"

	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusPaid      TransactionStatus = "paid"
	TransactionStatusProcessed TransactionStatus = "processed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

func (action SubscriptionAction) IsValid() bool {
	switch action {
	case SubscriptionActionNew, SubscriptionActionRenew, SubscriptionActionUpgrade, SubscriptionActionDowngrade:
		return true
	default:
		return false
	}
}

type Transaction struct {
	ID                    UUID
	ExternalTransactionID string
	GatewayTransactionID  string
	OrderID               string
	UserID                UUID
	SubscriptionID        UUID
	TierID                UUID
	TierCode              Tier
	TierSnapshot          TierSnapshot
	SubscriptionAction    SubscriptionAction
	SubscriptionDays      int
	CurrentSubscriptionID UUID
	CurrentTierSnapshot   *TierSnapshot
	CurrentEndDate        *time.Time
	ProratedCredit        float64
	FinalAmount           float64
	GrossAmount           float64
	Currency              string
	PaymentType           string
	TransactionStatus     TransactionStatus
	PaymentStatus         string
	FraudStatus           string
	SignatureKey          string
	TransactionTime       *time.Time
	SettlementTime        *time.Time
	ExpiryTime            *time.Time
	SettlementTimeRaw     string
	RawPayload            []byte
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
