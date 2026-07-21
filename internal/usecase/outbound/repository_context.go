package outbound

type IRepositoryContext interface {
	UserRepository
	TierRepository
	VideoRepository
	SubscriptionRepository
	TransactionRepository
	SubscriptionActivator
}
