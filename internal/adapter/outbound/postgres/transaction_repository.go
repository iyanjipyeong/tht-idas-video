package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
	"idas-video/internal/entity"
)

func (store *Store) CreateTransaction(ctx context.Context, transaction *entity.Transaction) (bool, error) {
	model := newTransactionModel(transaction)
	err := store.db.WithContext(ctx).Create(&model).Error
	if err == nil {
		if transaction.ID == entity.EmptyString {
			transaction.ID = model.ID
		}
		return true, nil
	}
	if isUniqueConstraintError(err) {
		return false, nil
	}
	if mappedErr := mapConstraintError(err); mappedErr != nil {
		return false, mappedErr
	}
	return false, err
}

func (store *Store) UpdateTransaction(ctx context.Context, transaction *entity.Transaction) error {
	model := newTransactionModel(transaction)
	err := store.db.WithContext(ctx).
		Model(&transactionModel{}).
		Where("external_transaction_id = ?", transaction.ExternalTransactionID).
		Updates(map[string]any{
			"gateway_transaction_id":  model.GatewayTransactionID,
			"order_id":                model.OrderID,
			"subscription_id":         model.SubscriptionID,
			"tier_id":                 model.TierID,
			"tier_name":               model.TierName,
			"tier_level":              model.TierLevel,
			"tier_price":              model.TierPrice,
			"tier_currency":           model.TierCurrency,
			"subscription_action":     model.SubscriptionAction,
			"subscription_days":       model.SubscriptionDays,
			"current_subscription_id": model.CurrentSubscriptionID,
			"current_tier_id":         model.CurrentTierID,
			"current_tier_name":       model.CurrentTierName,
			"current_tier_level":      model.CurrentTierLevel,
			"current_tier_price":      model.CurrentTierPrice,
			"current_end_date":        model.CurrentEndDate,
			"prorated_credit":         model.ProratedCredit,
			"final_amount":            model.FinalAmount,
			"gross_amount":            model.GrossAmount,
			"currency":                model.Currency,
			"payment_type":            model.PaymentType,
			"transaction_status":      model.TransactionStatus,
			"payment_status":          model.PaymentStatus,
			"fraud_status":            model.FraudStatus,
			"signature_key":           model.SignatureKey,
			"transaction_time":        model.TransactionTime,
			"settlement_time":         model.SettlementTime,
			"expiry_time":             model.ExpiryTime,
			"settlement_time_raw":     model.SettlementTimeRaw,
			"raw_payload":             model.RawPayload,
			"updated_at":              model.UpdatedAt,
		}).Error
	if mappedErr := mapConstraintError(err); mappedErr != nil {
		return mappedErr
	}
	return err
}

func (store *Store) TransactionExists(ctx context.Context, transactionID string) (bool, error) {
	var count int64
	if err := store.db.WithContext(ctx).
		Model(&transactionModel{}).
		Where("external_transaction_id = ?", transactionID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (store *Store) GetTransactionByExternalID(ctx context.Context, externalTransactionID string) (*entity.Transaction, error) {
	var model transactionModel
	if err := store.db.WithContext(ctx).
		Where("external_transaction_id = ?", externalTransactionID).
		First(&model).Error; err != nil {
		return nil, mapRecordNotFound(err, entity.ErrTransactionNotFound)
	}

	transaction := model.toDomain()
	return &transaction, nil
}

func (store *Store) FindPendingSubscriptionTransaction(ctx context.Context, userID entity.UUID, tierID entity.UUID, action entity.SubscriptionAction, durationDays int) (*entity.Transaction, error) {
	var model transactionModel
	err := store.db.WithContext(ctx).
		Where("user_id = ? AND tier_id = ? AND subscription_action = ? AND subscription_days = ? AND transaction_status = ?", userID, tierID, string(action), durationDays, entity.TransactionStatusPending).
		Order("created_at DESC, id DESC").
		First(&model).Error
	if err != nil {
		return nil, mapRecordNotFound(err, entity.ErrTransactionNotFound)
	}

	transaction := model.toDomain()
	return &transaction, nil
}

func (store *Store) ListTransactionsByUserID(ctx context.Context, userID entity.UUID) ([]entity.Transaction, error) {
	models := []transactionModel{}
	if err := store.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	transactions := make([]entity.Transaction, 0, len(models))
	for _, model := range models {
		transactions = append(transactions, model.toDomain())
	}
	return transactions, nil
}

func (store *Store) GetTransactionByID(ctx context.Context, userID entity.UUID, transactionID entity.UUID) (*entity.Transaction, error) {
	var model transactionModel
	if err := store.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", transactionID, userID).
		First(&model).Error; err != nil {
		return nil, mapRecordNotFound(err, entity.ErrTransactionNotFound)
	}

	transaction := model.toDomain()
	return &transaction, nil
}

func newTransactionModel(transaction *entity.Transaction) transactionModel {
	model := transactionModel{
		ID:                    transaction.ID,
		ExternalTransactionID: transaction.ExternalTransactionID,
		GatewayTransactionID:  transaction.GatewayTransactionID,
		OrderID:               transaction.OrderID,
		UserID:                transaction.UserID,
		TierID:                transaction.TierID,
		TierName:              transaction.TierSnapshot.TierName,
		TierLevel:             transaction.TierSnapshot.TierLevel,
		TierPrice:             transaction.TierSnapshot.TierPrice,
		TierCurrency:          transaction.TierSnapshot.TierCurrency,
		SubscriptionAction:    string(transaction.SubscriptionAction),
		SubscriptionDays:      transaction.SubscriptionDays,
		ProratedCredit:        transaction.ProratedCredit,
		FinalAmount:           transaction.FinalAmount,
		Currency:              transaction.Currency,
		TransactionStatus:     string(transaction.TransactionStatus),
		PaymentStatus:         transaction.PaymentStatus,
		TransactionTime:       transaction.TransactionTime,
		SettlementTime:        transaction.SettlementTime,
		ExpiryTime:            transaction.ExpiryTime,
		RawPayload:            json.RawMessage(transaction.RawPayload),
		CreatedAt:             transaction.CreatedAt,
		UpdatedAt:             transaction.UpdatedAt,
	}
	if transaction.SubscriptionID != entity.EmptyString {
		model.SubscriptionID = &transaction.SubscriptionID
	}
	if transaction.CurrentSubscriptionID != entity.EmptyString {
		model.CurrentSubscriptionID = &transaction.CurrentSubscriptionID
	}
	if transaction.CurrentTierSnapshot != nil {
		model.CurrentTierID = &transaction.CurrentTierSnapshot.TierID
		model.CurrentTierName = &transaction.CurrentTierSnapshot.TierName
		model.CurrentTierLevel = &transaction.CurrentTierSnapshot.TierLevel
		model.CurrentTierPrice = &transaction.CurrentTierSnapshot.TierPrice
	}
	if transaction.CurrentEndDate != nil {
		model.CurrentEndDate = transaction.CurrentEndDate
	}
	if transaction.GrossAmount > 0 {
		model.GrossAmount = &transaction.GrossAmount
	}
	model.PaymentType = stringPtr(transaction.PaymentType)
	model.FraudStatus = stringPtr(transaction.FraudStatus)
	model.SignatureKey = stringPtr(transaction.SignatureKey)
	model.SettlementTimeRaw = stringPtr(transaction.SettlementTimeRaw)
	return model
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := err.Error()
	return message != "" && (contains(message, "duplicate key") || contains(message, "unique constraint"))
}

func contains(value string, needle string) bool {
	return strings.Contains(value, needle)
}
