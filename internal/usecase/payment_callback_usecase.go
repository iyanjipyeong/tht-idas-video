package usecase

import (
	"context"
	"time"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
	"idas-video/internal/usecase/outbound"
)

type PaymentCallbackUsecase struct {
	transactions  outbound.TransactionRepository
	subscriptions outbound.SubscriptionActivator
	tiers         outbound.TierRepository
	logger        outbound.Logger
}

func NewPaymentCallbackUsecase(transactions outbound.TransactionRepository, subscriptions outbound.SubscriptionActivator, tiers outbound.TierRepository) *PaymentCallbackUsecase {
	return NewPaymentCallbackUsecaseWithLogger(transactions, subscriptions, tiers, nil)
}

func NewPaymentCallbackUsecaseWithLogger(transactions outbound.TransactionRepository, subscriptions outbound.SubscriptionActivator, tiers outbound.TierRepository, log outbound.Logger) *PaymentCallbackUsecase {
	return &PaymentCallbackUsecase{transactions: transactions, subscriptions: subscriptions, tiers: tiers, logger: fallbackLogger(log)}
}

func (usecase *PaymentCallbackUsecase) ProcessPaymentCallback(ctx context.Context, payload inbound.PaymentCallbackRequest) error {
	usecase.logger.Info(ctx, "usecase.payment", "payment callback processing started", "payment.callback.started", outbound.LogField{Key: "transaction.id", Value: payload.TransactionID}, outbound.LogField{Key: "user.id", Value: payload.UserID.String()}, outbound.LogField{Key: "payment.status", Value: payload.PaymentStatus})
	if payload.TransactionID == entity.EmptyString || payload.UserID == entity.EmptyString || !entity.IsUUID(payload.UserID.String()) || !payload.Tier.IsValid() || payload.SubscriptionDays <= 0 || payload.PaymentStatus == "" {
		usecase.logger.Warn(ctx, "usecase.payment", "payment callback rejected", "payment.callback.rejected")
		return entity.ErrInvalidRequest
	}

	tierDetail, err := usecase.tiers.GetTierByCode(ctx, payload.Tier)
	if err != nil {
		usecase.logger.Error(ctx, "usecase.payment", "payment callback tier lookup failed", "payment.callback.tier_lookup_failed", err, outbound.LogField{Key: "tier", Value: payload.Tier.String()})
		return err
	}

	existing, err := usecase.transactions.GetTransactionByExternalID(ctx, payload.TransactionID)
	if err != nil && err != entity.ErrTransactionNotFound {
		usecase.logger.Error(ctx, "usecase.payment", "payment callback transaction lookup failed", "payment.callback.transaction_lookup_failed", err)
		return err
	}
	if err == entity.ErrTransactionNotFound || existing == nil {
		usecase.logger.Warn(ctx, "usecase.payment", "payment callback rejected because reserved transaction was not found", "payment.callback.transaction_not_found", outbound.LogField{Key: "transaction.id", Value: payload.TransactionID}, outbound.LogField{Key: "order.id", Value: payload.OrderID}, outbound.LogField{Key: "user.id", Value: payload.UserID.String()})
		return entity.ErrTransactionMismatch
	}

	transaction := existing
	if err := validateCallbackAgainstReservedTransaction(existing, payload, tierDetail); err != nil {
		usecase.logger.Warn(ctx, "usecase.payment", "payment callback transaction mismatch", "payment.callback.mismatch", outbound.LogField{Key: "transaction.id", Value: payload.TransactionID})
		return err
	}
	if existing.TransactionStatus == entity.TransactionStatusProcessed {
		usecase.logger.Warn(ctx, "usecase.payment", "payment callback ignored because transaction already processed", "payment.callback.already_processed", outbound.LogField{Key: "transaction.id", Value: payload.TransactionID})
		return entity.ErrTransactionProcessed
	}
	mergeCallbackIntoTransaction(existing, payload)
	transaction = existing

	var activation *outbound.SubscriptionActivationResult
	switch payload.PaymentStatus {
	case entity.PaymentStatusPaid:
		transaction.TransactionStatus = entity.TransactionStatusPaid
		activation, err = usecase.subscriptions.ActivateSubscription(ctx, payload.UserID, payload.Tier, payload.SubscriptionDays)
		if err != nil {
			transaction.TransactionStatus = entity.TransactionStatusFailed
			transaction.UpdatedAt = time.Now()
			_ = usecase.transactions.UpdateTransaction(ctx, transaction)
			usecase.logger.Error(ctx, "usecase.payment", "payment callback subscription activation failed", "payment.callback.activation_failed", err)
			return err
		}
	case entity.PaymentStatusPending:
		transaction.TransactionStatus = entity.TransactionStatusPending
	case "failed", "deny", "cancel", "expire", "expired":
		transaction.TransactionStatus = entity.TransactionStatusFailed
	default:
		transaction.TransactionStatus = entity.TransactionStatusPending
	}

	if activation != nil {
		if activation.Subscription != nil {
			transaction.TierID = activation.Subscription.TierID
			transaction.TierSnapshot = activation.Subscription.TierSnapshot
			transaction.SubscriptionID = activation.Subscription.ID
		}
		transaction.SubscriptionAction = activation.Action
		transaction.ProratedCredit = activation.ProratedCredit
		transaction.FinalAmount = activation.FinalAmount
		if activation.Current != nil {
			transaction.CurrentSubscriptionID = activation.Current.ID
			transaction.CurrentTierSnapshot = &activation.Current.TierSnapshot
			transaction.CurrentEndDate = &activation.Current.EndDate
		}
		if transaction.GrossAmount == 0 {
			transaction.GrossAmount = activation.FinalAmount
		}
		transaction.TransactionStatus = entity.TransactionStatusProcessed
	}

	transaction.RawPayload = payload.RawPayload
	transaction.UpdatedAt = time.Now()
	if err := usecase.transactions.UpdateTransaction(ctx, transaction); err != nil {
		usecase.logger.Error(ctx, "usecase.payment", "payment callback update failed", "payment.callback.update_failed", err)
		return err
	}

	usecase.logger.Info(ctx, "usecase.payment", "payment callback processed", "payment.callback.completed", outbound.LogField{Key: "transaction.id", Value: payload.TransactionID}, outbound.LogField{Key: "transaction.status", Value: transaction.TransactionStatus}, outbound.LogField{Key: "payment.status", Value: transaction.PaymentStatus})
	return nil
}

func validateCallbackAgainstReservedTransaction(existing *entity.Transaction, payload inbound.PaymentCallbackRequest, tierDetail *entity.TierDetail) error {
	if existing == nil {
		return entity.ErrTransactionNotFound
	}
	if existing.UserID != payload.UserID || existing.TierID != tierDetail.ID || existing.SubscriptionDays != payload.SubscriptionDays {
		return entity.ErrTransactionMismatch
	}
	if existing.OrderID != "" && payload.OrderID != existing.OrderID {
		return entity.ErrTransactionMismatch
	}
	if existing.GrossAmount != payload.GrossAmount {
		return entity.ErrTransactionMismatch
	}
	return nil
}

func mergeCallbackIntoTransaction(transaction *entity.Transaction, payload inbound.PaymentCallbackRequest) {
	if payload.GatewayTransactionID != "" {
		transaction.GatewayTransactionID = payload.GatewayTransactionID
	}
	if payload.OrderID != "" {
		transaction.OrderID = payload.OrderID
	}
	if payload.GrossAmount > 0 {
		transaction.GrossAmount = payload.GrossAmount
		transaction.FinalAmount = payload.GrossAmount
	}
	if payload.Currency != "" {
		transaction.Currency = payload.Currency
	}
	transaction.PaymentType = payload.PaymentType
	transaction.PaymentStatus = payload.PaymentStatus
	transaction.FraudStatus = payload.FraudStatus
	transaction.SignatureKey = payload.SignatureKey
	transaction.TransactionTime = payload.TransactionTime
	transaction.SettlementTime = payload.SettlementTime
	transaction.ExpiryTime = payload.ExpiryTime
	transaction.SettlementTimeRaw = payload.SettlementTimeRaw
}
