package outbound

import (
	"context"

	"idas-video/internal/entity"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id entity.UUID) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}

type TierRepository interface {
	ListTiers(ctx context.Context) ([]entity.TierDetail, error)
	GetTierByID(ctx context.Context, id entity.UUID) (*entity.TierDetail, error)
	GetTierByCode(ctx context.Context, code entity.Tier) (*entity.TierDetail, error)
}

type VideoRepository interface {
	ListVideos(ctx context.Context) ([]entity.Video, error)
	GetVideoByID(ctx context.Context, id entity.UUID) (*entity.Video, error)
}

type SubscriptionRepository interface {
	GetActiveSubscriptionByUserID(ctx context.Context, userID entity.UUID) (*entity.Subscription, error)
	UpsertSubscription(ctx context.Context, subscription *entity.Subscription) error
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *entity.Transaction) (bool, error)
	UpdateTransaction(ctx context.Context, transaction *entity.Transaction) error
	TransactionExists(ctx context.Context, transactionID string) (bool, error)
	GetTransactionByExternalID(ctx context.Context, externalTransactionID string) (*entity.Transaction, error)
	FindPendingSubscriptionTransaction(ctx context.Context, userID entity.UUID, tierID entity.UUID, action entity.SubscriptionAction, durationDays int) (*entity.Transaction, error)
	ListTransactionsByUserID(ctx context.Context, userID entity.UUID) ([]entity.Transaction, error)
	GetTransactionByID(ctx context.Context, userID entity.UUID, transactionID entity.UUID) (*entity.Transaction, error)
}
