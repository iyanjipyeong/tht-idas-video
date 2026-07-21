package postgres

import "idas-video/internal/usecase/outbound"

var (
	_ outbound.UserRepository         = (*Store)(nil)
	_ outbound.TierRepository         = (*Store)(nil)
	_ outbound.VideoRepository        = (*Store)(nil)
	_ outbound.SubscriptionRepository = (*Store)(nil)
	_ outbound.TransactionRepository  = (*Store)(nil)
)
