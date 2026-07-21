package outbound

import (
	"context"

	"idas-video/internal/entity"
)

type SubscriptionActivationResult struct {
	Subscription   *entity.Subscription
	Action         entity.SubscriptionAction
	Current        *entity.Subscription
	ProratedCredit float64
	FinalAmount    float64
}

type SubscriptionActivator interface {
	ActivateSubscription(ctx context.Context, userID entity.UUID, tier entity.Tier, durationDays int) (*SubscriptionActivationResult, error)
}
