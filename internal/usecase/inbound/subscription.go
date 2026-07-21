package inbound

import (
	"context"

	"idas-video/internal/entity"
)

type ActiveSubscriptionUsecase interface {
	GetActiveSubscription(ctx context.Context, userID entity.UUID) (*entity.Subscription, error)
}

type SubscriptionPurchaseUsecase interface {
	ActiveSubscriptionUsecase
	CreateSubscriptionTransaction(ctx context.Context, userID entity.UUID, tierID entity.UUID, subscriptionAction entity.SubscriptionAction, durationDays int) (*entity.Transaction, error)
}
