package postgres

import (
	"encoding/json"
	"time"

	"idas-video/internal/entity"
)

type userModel struct {
	ID        entity.UUID `gorm:"type:uuid;primaryKey;column:id;default:gen_random_uuid()"`
	Name      string      `gorm:"column:name"`
	Email     string      `gorm:"column:email"`
	Password  string      `gorm:"column:password"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (userModel) TableName() string { return "users" }

func (model userModel) toDomain() entity.User {
	return entity.User{ID: model.ID, Name: model.Name, Email: model.Email, Password: model.Password, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

type tierModel struct {
	ID          entity.UUID `gorm:"type:uuid;primaryKey;column:id;default:gen_random_uuid()"`
	Name        string      `gorm:"column:name"`
	Level       int         `gorm:"column:level"`
	Price       float64     `gorm:"column:price"`
	Currency    string      `gorm:"column:currency"`
	Description string      `gorm:"column:description"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (tierModel) TableName() string { return "tiers" }

func (model tierModel) toDomain() entity.TierDetail {
	code, _ := entity.ParseTier(model.Name)
	return entity.TierDetail{ID: model.ID, Code: code, Name: model.Name, Level: model.Level, Price: model.Price, Currency: model.Currency, Description: model.Description}
}

type videoModel struct {
	ID          entity.UUID `gorm:"type:uuid;primaryKey;column:id;default:gen_random_uuid()"`
	Title       string      `gorm:"column:title"`
	Description *string     `gorm:"column:description"`
	TierID      entity.UUID `gorm:"type:uuid;column:tier_id"`
	VideoURL    string      `gorm:"column:video_url"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (videoModel) TableName() string { return "videos" }

func (model videoModel) toDomain() entity.Video {
	description := ""
	if model.Description != nil {
		description = *model.Description
	}
	category := inferTierCodeFromUUID(model.TierID)
	return entity.Video{ID: model.ID, Title: model.Title, Description: description, Category: category, VideoURL: model.VideoURL, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

type subscriptionModel struct {
	ID           entity.UUID               `gorm:"type:uuid;primaryKey;column:id;default:gen_random_uuid()"`
	UserID       entity.UUID               `gorm:"type:uuid;column:user_id"`
	TierID       entity.UUID               `gorm:"type:uuid;column:tier_id"`
	TierName     string                    `gorm:"column:tier_name"`
	TierLevel    int                       `gorm:"column:tier_level"`
	TierPrice    float64                   `gorm:"column:tier_price"`
	TierCurrency string                    `gorm:"column:tier_currency"`
	Status       entity.SubscriptionStatus `gorm:"column:status"`
	StartDate    time.Time                 `gorm:"column:start_date"`
	EndDate      time.Time                 `gorm:"column:end_date"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (subscriptionModel) TableName() string { return "subscriptions" }

func newSubscriptionModel(subscription *entity.Subscription) subscriptionModel {
	return subscriptionModel{ID: subscription.ID, UserID: subscription.UserID, TierID: subscription.TierID, TierName: subscription.TierSnapshot.TierName, TierLevel: subscription.TierSnapshot.TierLevel, TierPrice: subscription.TierSnapshot.TierPrice, TierCurrency: subscription.TierSnapshot.TierCurrency, Status: subscription.Status, StartDate: subscription.StartDate, EndDate: subscription.EndDate, CreatedAt: subscription.CreatedAt, UpdatedAt: subscription.UpdatedAt}
}

func (model subscriptionModel) toDomain() entity.Subscription {
	code, _ := entity.ParseTier(model.TierName)
	return entity.Subscription{ID: model.ID, UserID: model.UserID, TierID: model.TierID, TierCode: code, TierSnapshot: entity.TierSnapshot{TierID: model.TierID, TierCode: code, TierName: model.TierName, TierLevel: model.TierLevel, TierPrice: model.TierPrice, TierCurrency: model.TierCurrency}, Status: model.Status, StartDate: model.StartDate, EndDate: model.EndDate, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

type transactionModel struct {
	ID                    entity.UUID     `gorm:"type:uuid;primaryKey;column:id;default:gen_random_uuid()"`
	ExternalTransactionID string          `gorm:"column:external_transaction_id"`
	GatewayTransactionID  string          `gorm:"column:gateway_transaction_id"`
	OrderID               string          `gorm:"column:order_id"`
	UserID                entity.UUID     `gorm:"type:uuid;column:user_id"`
	SubscriptionID        *entity.UUID    `gorm:"type:uuid;column:subscription_id"`
	TierID                entity.UUID     `gorm:"type:uuid;column:tier_id"`
	TierName              string          `gorm:"column:tier_name"`
	TierLevel             int             `gorm:"column:tier_level"`
	TierPrice             float64         `gorm:"column:tier_price"`
	TierCurrency          string          `gorm:"column:tier_currency"`
	SubscriptionAction    string          `gorm:"column:subscription_action"`
	SubscriptionDays      int             `gorm:"column:subscription_days"`
	CurrentSubscriptionID *entity.UUID    `gorm:"type:uuid;column:current_subscription_id"`
	CurrentTierID         *entity.UUID    `gorm:"type:uuid;column:current_tier_id"`
	CurrentTierName       *string         `gorm:"column:current_tier_name"`
	CurrentTierLevel      *int            `gorm:"column:current_tier_level"`
	CurrentTierPrice      *float64        `gorm:"column:current_tier_price"`
	CurrentEndDate        *time.Time      `gorm:"column:current_end_date"`
	ProratedCredit        float64         `gorm:"column:prorated_credit"`
	FinalAmount           float64         `gorm:"column:final_amount"`
	GrossAmount           *float64        `gorm:"column:gross_amount"`
	Currency              string          `gorm:"column:currency"`
	PaymentType           *string         `gorm:"column:payment_type"`
	TransactionStatus     string          `gorm:"column:transaction_status"`
	PaymentStatus         string          `gorm:"column:payment_status"`
	FraudStatus           *string         `gorm:"column:fraud_status"`
	SignatureKey          *string         `gorm:"column:signature_key"`
	TransactionTime       *time.Time      `gorm:"column:transaction_time"`
	SettlementTime        *time.Time      `gorm:"column:settlement_time"`
	ExpiryTime            *time.Time      `gorm:"column:expiry_time"`
	SettlementTimeRaw     *string         `gorm:"column:settlement_time_raw"`
	RawPayload            json.RawMessage `gorm:"column:raw_payload;type:jsonb"`
	CreatedAt             time.Time       `gorm:"column:created_at"`
	UpdatedAt             time.Time       `gorm:"column:updated_at"`
}

func (transactionModel) TableName() string { return "transactions" }

func (model transactionModel) toDomain() entity.Transaction {
	transaction := entity.Transaction{
		ID:                    model.ID,
		ExternalTransactionID: model.ExternalTransactionID,
		GatewayTransactionID:  model.GatewayTransactionID,
		OrderID:               model.OrderID,
		UserID:                model.UserID,
		TierID:                model.TierID,
		TierCode:              inferTierCodeFromName(model.TierName),
		TierSnapshot: entity.TierSnapshot{
			TierID:       model.TierID,
			TierCode:     inferTierCodeFromName(model.TierName),
			TierName:     model.TierName,
			TierLevel:    model.TierLevel,
			TierPrice:    model.TierPrice,
			TierCurrency: model.TierCurrency,
		},
		SubscriptionAction: entity.SubscriptionAction(model.SubscriptionAction),
		SubscriptionDays:   model.SubscriptionDays,
		ProratedCredit:     model.ProratedCredit,
		FinalAmount:        model.FinalAmount,
		Currency:           model.Currency,
		TransactionStatus:  entity.TransactionStatus(model.TransactionStatus),
		PaymentStatus:      model.PaymentStatus,
		TransactionTime:    model.TransactionTime,
		SettlementTime:     model.SettlementTime,
		ExpiryTime:         model.ExpiryTime,
		RawPayload:         []byte(model.RawPayload),
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
	if model.SubscriptionID != nil {
		transaction.SubscriptionID = *model.SubscriptionID
	}
	if model.CurrentSubscriptionID != nil {
		transaction.CurrentSubscriptionID = *model.CurrentSubscriptionID
	}
	if model.CurrentTierID != nil {
		transaction.CurrentTierSnapshot = &entity.TierSnapshot{TierID: *model.CurrentTierID}
		if model.CurrentTierName != nil {
			transaction.CurrentTierSnapshot.TierName = *model.CurrentTierName
			transaction.CurrentTierSnapshot.TierCode = inferTierCodeFromName(*model.CurrentTierName)
		}
		if model.CurrentTierLevel != nil {
			transaction.CurrentTierSnapshot.TierLevel = *model.CurrentTierLevel
		}
		if model.CurrentTierPrice != nil {
			transaction.CurrentTierSnapshot.TierPrice = *model.CurrentTierPrice
		}
	}
	transaction.CurrentEndDate = model.CurrentEndDate
	if model.GrossAmount != nil {
		transaction.GrossAmount = *model.GrossAmount
	}
	if model.PaymentType != nil {
		transaction.PaymentType = *model.PaymentType
	}
	if model.FraudStatus != nil {
		transaction.FraudStatus = *model.FraudStatus
	}
	if model.SignatureKey != nil {
		transaction.SignatureKey = *model.SignatureKey
	}
	if model.SettlementTimeRaw != nil {
		transaction.SettlementTimeRaw = *model.SettlementTimeRaw
	}
	return transaction
}

func inferTierCodeFromUUID(id entity.UUID) entity.Tier {
	switch id.String() {
	case "11111111-1111-4111-8111-111111111101":
		return entity.TierBronze
	case "11111111-1111-4111-8111-111111111102":
		return entity.TierSilver
	case "11111111-1111-4111-8111-111111111103":
		return entity.TierGold
	default:
		return entity.TierUnknown
	}
}

func inferTierCodeFromName(name string) entity.Tier {
	tier, err := entity.ParseTier(name)
	if err != nil {
		return entity.TierUnknown
	}
	return tier
}
