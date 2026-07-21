package postgres

import (
	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

var _ outbound.UserRepository = (*Store)(nil)
var _ outbound.TierRepository = (*Store)(nil)
var _ outbound.VideoRepository = (*Store)(nil)
var _ outbound.SubscriptionRepository = (*Store)(nil)
var _ outbound.TransactionRepository = (*Store)(nil)

var _ entity.UUID = userModel{}.ID
var _ entity.UUID = tierModel{}.ID
var _ entity.UUID = videoModel{}.ID
var _ entity.UUID = subscriptionModel{}.ID
var _ entity.UUID = subscriptionModel{}.UserID
var _ entity.UUID = transactionModel{}.ID
var _ entity.UUID = transactionModel{}.UserID
