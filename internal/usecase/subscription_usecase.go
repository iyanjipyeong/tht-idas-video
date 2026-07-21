package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

type SubscriptionUsecase struct {
	subscriptions outbound.SubscriptionRepository
	tiers         outbound.TierRepository
	transactions  outbound.TransactionRepository
	logger        outbound.Logger
}

func NewSubscriptionUsecase(subscriptions outbound.SubscriptionRepository, tiers outbound.TierRepository, transactions outbound.TransactionRepository) *SubscriptionUsecase {
	return NewSubscriptionUsecaseWithLogger(subscriptions, tiers, transactions, nil)
}

func NewSubscriptionUsecaseWithLogger(subscriptions outbound.SubscriptionRepository, tiers outbound.TierRepository, transactions outbound.TransactionRepository, log outbound.Logger) *SubscriptionUsecase {
	return &SubscriptionUsecase{subscriptions: subscriptions, tiers: tiers, transactions: transactions, logger: fallbackLogger(log)}
}

func (usecase *SubscriptionUsecase) GetActiveSubscription(ctx context.Context, userID entity.UUID) (*entity.Subscription, error) {
	usecase.logger.Info(ctx, "usecase.subscription", "fetch active subscription", "subscription.active.fetch", outbound.LogField{Key: "user.id", Value: userID.String()})
	return usecase.subscriptions.GetActiveSubscriptionByUserID(ctx, userID)
}

func (usecase *SubscriptionUsecase) ActivateSubscription(ctx context.Context, userID entity.UUID, tier entity.Tier, durationDays int) (*outbound.SubscriptionActivationResult, error) {
	if userID == entity.EmptyString || !tier.IsValid() || durationDays <= 0 {
		usecase.logger.Warn(ctx, "usecase.subscription", "activate subscription rejected", "subscription.activate.rejected", outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "tier", Value: tier}, outbound.LogField{Key: "duration.days", Value: durationDays})
		return nil, entity.ErrInvalidSubscription
	}

	tierDetail, err := usecase.tiers.GetTierByCode(ctx, tier)
	if err != nil {
		usecase.logger.Error(ctx, "usecase.subscription", "tier lookup failed", "subscription.activate.tier_lookup_failed", err, outbound.LogField{Key: "tier", Value: tier})
		return nil, err
	}

	now := time.Now()
	result := &outbound.SubscriptionActivationResult{Action: entity.SubscriptionActionNew, FinalAmount: tierDetail.Price}
	current, err := usecase.subscriptions.GetActiveSubscriptionByUserID(ctx, userID)
	if err == nil && current != nil && current.IsActive(now) {
		result.Current = current
		switch {
		case current.TierCode == tier:
			result.Action = entity.SubscriptionActionRenew
			result.FinalAmount = tierDetail.Price
		case current.TierCode.Level() > tier.Level():
			return nil, entity.ErrDowngradeTier
		case current.TierCode.Level() < tier.Level():
			result.Action = entity.SubscriptionActionUpgrade
			remainingDays := current.EndDate.Sub(now).Hours() / 24
			if remainingDays < 0 {
				remainingDays = 0
			}
			proratedCredit := (remainingDays / float64(durationDays)) * current.TierSnapshot.TierPrice
			if proratedCredit > tierDetail.Price {
				proratedCredit = tierDetail.Price
			}
			result.ProratedCredit = proratedCredit
			result.FinalAmount = tierDetail.Price - proratedCredit
		}
	}

	startDate := now
	if result.Action == entity.SubscriptionActionRenew && result.Current != nil {
		startDate = result.Current.StartDate
		now = result.Current.EndDate
	}

	subscription := &entity.Subscription{
		ID:       entity.EmptyString,
		UserID:   userID,
		TierID:   tierDetail.ID,
		TierCode: tier,
		TierSnapshot: entity.TierSnapshot{
			TierID:       tierDetail.ID,
			TierCode:     tier,
			TierName:     tierDetail.Name,
			TierLevel:    tierDetail.Level,
			TierPrice:    tierDetail.Price,
			TierCurrency: tierDetail.Currency,
		},
		Status:    entity.SubscriptionStatusActive,
		StartDate: startDate,
		EndDate:   now.AddDate(0, 0, durationDays),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if result.Current != nil {
		subscription.ID = result.Current.ID
	}

	if err := usecase.subscriptions.UpsertSubscription(ctx, subscription); err != nil {
		usecase.logger.Error(ctx, "usecase.subscription", "activate subscription failed", "subscription.activate.failed", err, outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "tier", Value: tier})
		return nil, err
	}

	result.Subscription = subscription
	usecase.logger.Info(ctx, "usecase.subscription", "subscription activated", "subscription.activate.completed", outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "tier", Value: tier}, outbound.LogField{Key: "duration.days", Value: durationDays}, outbound.LogField{Key: "action", Value: result.Action}, outbound.LogField{Key: "final.amount", Value: result.FinalAmount}, outbound.LogField{Key: "prorated.credit", Value: result.ProratedCredit})
	return result, nil
}

func (usecase *SubscriptionUsecase) CreateSubscriptionTransaction(ctx context.Context, userID entity.UUID, tierID entity.UUID, subscriptionAction entity.SubscriptionAction, durationDays int) (*entity.Transaction, error) {
	if userID == entity.EmptyString || tierID == entity.EmptyString || !entity.IsUUID(tierID.String()) || durationDays <= 0 || !subscriptionAction.IsValid() {
		usecase.logger.Warn(ctx, "usecase.subscription", "subscription transaction rejected", "subscription.transaction.rejected", outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "tier.id", Value: tierID.String()}, outbound.LogField{Key: "duration.days", Value: durationDays})
		return nil, entity.ErrInvalidSubscription
	}

	tierDetail, err := usecase.tiers.GetTierByID(ctx, tierID)
	if err != nil {
		usecase.logger.Error(ctx, "usecase.subscription", "subscription transaction tier lookup failed", "subscription.transaction.tier_lookup_failed", err, outbound.LogField{Key: "tier.id", Value: tierID.String()})
		return nil, err
	}
	tier := tierDetail.Code
	now := time.Now()
	action := entity.SubscriptionActionNew
	finalAmount := tierDetail.Price
	var proratedCredit float64
	var currentSubscriptionID entity.UUID
	var currentTierSnapshot *entity.TierSnapshot
	var currentEndDate *time.Time

	current, err := usecase.subscriptions.GetActiveSubscriptionByUserID(ctx, userID)
	hasActive := err == nil && current != nil && current.IsActive(now)
	if !hasActive {
		if subscriptionAction != entity.SubscriptionActionNew {
			return nil, entity.ErrSubscriptionActionRequiresActive
		}
	} else {
		switch subscriptionAction {
		case entity.SubscriptionActionNew:
			return nil, entity.ErrRepurchaseTier
		case entity.SubscriptionActionRenew:
			if current.TierCode != tier {
				return nil, entity.ErrSubscriptionActionMismatch
			}
			action = entity.SubscriptionActionRenew
			finalAmount = tierDetail.Price
		case entity.SubscriptionActionUpgrade:
			if tier.Level() <= current.TierCode.Level() {
				return nil, entity.ErrSubscriptionActionMismatch
			}
			action = entity.SubscriptionActionUpgrade
			remainingDays := current.EndDate.Sub(now).Hours() / 24
			if remainingDays < 0 {
				remainingDays = 0
			}
			proratedCredit = (remainingDays / float64(durationDays)) * current.TierSnapshot.TierPrice
			if proratedCredit > tierDetail.Price {
				proratedCredit = tierDetail.Price
			}
			finalAmount = tierDetail.Price - proratedCredit
		case entity.SubscriptionActionDowngrade:
			if tier.Level() >= current.TierCode.Level() {
				return nil, entity.ErrSubscriptionActionMismatch
			}
			return nil, entity.ErrDowngradeTier
		}
		currentSubscriptionID = current.ID
		currentTierSnapshot = &current.TierSnapshot
		currentEndDate = &current.EndDate
	}

	pending, err := usecase.transactions.FindPendingSubscriptionTransaction(ctx, userID, tierDetail.ID, action, durationDays)
	if err == nil && pending != nil {
		usecase.logger.Info(ctx, "usecase.subscription", "subscription transaction reused", "subscription.transaction.reused", outbound.LogField{Key: "transaction.id", Value: pending.ID.String()}, outbound.LogField{Key: "external.transaction.id", Value: pending.ExternalTransactionID}, outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "tier", Value: tier}, outbound.LogField{Key: "action", Value: action})
		return pending, nil
	}
	if err != nil && err != entity.ErrTransactionNotFound {
		usecase.logger.Error(ctx, "usecase.subscription", "subscription pending transaction lookup failed", "subscription.transaction.lookup_failed", err, outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "tier.id", Value: tierID.String()}, outbound.LogField{Key: "tier", Value: tier})
		return nil, err
	}

	transaction := &entity.Transaction{
		ExternalTransactionID: fmt.Sprintf("sub-%s-%d", userID.String(), now.UnixNano()),
		OrderID:               fmt.Sprintf("ORDER-SUB-%d", now.UnixNano()),
		UserID:                userID,
		TierID:                tierDetail.ID,
		TierCode:              tier,
		TierSnapshot:          tierDetail.ToSnapshot(),
		SubscriptionAction:    action,
		SubscriptionDays:      durationDays,
		CurrentSubscriptionID: currentSubscriptionID,
		CurrentTierSnapshot:   currentTierSnapshot,
		CurrentEndDate:        currentEndDate,
		ProratedCredit:        proratedCredit,
		FinalAmount:           finalAmount,
		GrossAmount:           finalAmount,
		Currency:              tierDetail.Currency,
		TransactionStatus:     entity.TransactionStatusPending,
		PaymentStatus:         entity.PaymentStatusPending,
		ExpiryTime:            timePtr(now.Add(24 * time.Hour)),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	transaction.RawPayload = mustMarshalSubscriptionTransaction(transaction)

	created, err := usecase.transactions.CreateTransaction(ctx, transaction)
	if err != nil {
		usecase.logger.Error(ctx, "usecase.subscription", "subscription transaction creation failed", "subscription.transaction.create_failed", err, outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "tier.id", Value: tierID.String()}, outbound.LogField{Key: "tier", Value: tier})
		return nil, err
	}
	if !created {
		return nil, entity.ErrDuplicateTransaction
	}

	usecase.logger.Info(ctx, "usecase.subscription", "subscription transaction created", "subscription.transaction.created", outbound.LogField{Key: "transaction.id", Value: transaction.ID.String()}, outbound.LogField{Key: "external.transaction.id", Value: transaction.ExternalTransactionID}, outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "tier", Value: tier}, outbound.LogField{Key: "action", Value: action}, outbound.LogField{Key: "final.amount", Value: finalAmount})
	return transaction, nil
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func mustMarshalSubscriptionTransaction(transaction *entity.Transaction) []byte {
	raw, err := json.Marshal(map[string]any{
		"source":                  "subscribe_api",
		"external_transaction_id": transaction.ExternalTransactionID,
		"order_id":                transaction.OrderID,
		"user_id":                 transaction.UserID.String(),
		"tier":                    transaction.TierCode.String(),
		"subscription_action":     transaction.SubscriptionAction,
		"subscription_days":       transaction.SubscriptionDays,
		"gross_amount":            transaction.GrossAmount,
		"currency":                transaction.Currency,
	})
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}
