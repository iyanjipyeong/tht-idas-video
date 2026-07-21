package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

func TestSubscriptionUsecaseActivateSubscriptionSuccess(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111103")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByCode", mock.Anything, entity.TierGold).Return(&entity.TierDetail{ID: tierID, Code: entity.TierGold, Name: "Gold", Level: 3, Price: 150000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetActiveSubscriptionByUserID", mock.Anything, userID).Return(nil, entity.ErrActiveSubscription).Once()
	repositoryContext.On("UpsertSubscription", mock.Anything, mock.MatchedBy(func(subscription *entity.Subscription) bool {
		return subscription != nil && subscription.UserID == userID && subscription.TierID == tierID && subscription.TierCode == entity.TierGold && subscription.Status == entity.SubscriptionStatusActive
	})).Return(nil).Once()

	usecase := NewSubscriptionUsecase(repositoryContext, repositoryContext, repositoryContext)
	result, err := usecase.ActivateSubscription(context.Background(), userID, entity.TierGold, 30)
	if err != nil {
		t.Fatalf("ActivateSubscription() error = %v", err)
	}
	if result == nil || result.Subscription == nil || result.Subscription.UserID != userID || result.Subscription.TierID != tierID {
		t.Fatal("ActivateSubscription() should return active subscription")
	}
}

func TestSubscriptionUsecaseActivateSubscriptionRejectsInvalidInput(t *testing.T) {
	usecase := NewSubscriptionUsecase(outbound.NewMockIRepositoryContext(t), outbound.NewMockIRepositoryContext(t), outbound.NewMockIRepositoryContext(t))
	if _, err := usecase.ActivateSubscription(context.Background(), "", entity.TierUnknown, 0); err != entity.ErrInvalidSubscription {
		t.Fatalf("ActivateSubscription() error = %v, want %v", err, entity.ErrInvalidSubscription)
	}
}

func TestSubscriptionUsecaseCreateSubscriptionTransactionSuccess(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111102")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByID", mock.Anything, tierID).Return(&entity.TierDetail{ID: tierID, Code: entity.TierSilver, Name: "Silver", Level: 2, Price: 100000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetActiveSubscriptionByUserID", mock.Anything, userID).Return(nil, entity.ErrActiveSubscription).Once()
	repositoryContext.On("FindPendingSubscriptionTransaction", mock.Anything, userID, tierID, entity.SubscriptionActionNew, 30).Return(nil, entity.ErrTransactionNotFound).Once()
	repositoryContext.On("CreateTransaction", mock.Anything, mock.MatchedBy(func(transaction *entity.Transaction) bool {
		return transaction != nil && transaction.UserID == userID && transaction.TierID == tierID && transaction.TransactionStatus == entity.TransactionStatusPending && transaction.PaymentStatus == entity.PaymentStatusPending
	})).Return(true, nil).Once()

	usecase := NewSubscriptionUsecase(repositoryContext, repositoryContext, repositoryContext)
	transaction, err := usecase.CreateSubscriptionTransaction(context.Background(), userID, tierID, entity.SubscriptionActionNew, 30)
	if err != nil {
		t.Fatalf("CreateSubscriptionTransaction() error = %v", err)
	}
	if transaction == nil || transaction.UserID != userID || transaction.TierID != tierID {
		t.Fatal("CreateSubscriptionTransaction() should return pending transaction")
	}
}

func TestSubscriptionUsecaseCreateSubscriptionTransactionIsIdempotentForSamePendingTier(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111102")
	existingTransactionID := entity.UUID("77777777-7777-4777-8777-777777777777")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByID", mock.Anything, tierID).Return(&entity.TierDetail{ID: tierID, Code: entity.TierSilver, Name: "Silver", Level: 2, Price: 100000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetActiveSubscriptionByUserID", mock.Anything, userID).Return(nil, entity.ErrActiveSubscription).Once()
	repositoryContext.On("FindPendingSubscriptionTransaction", mock.Anything, userID, tierID, entity.SubscriptionActionNew, 30).Return(&entity.Transaction{
		ID:                    existingTransactionID,
		ExternalTransactionID: "sub-existing",
		UserID:                userID,
		TierID:                tierID,
		TierCode:              entity.TierSilver,
		TierSnapshot:          entity.TierSnapshot{TierID: tierID, TierCode: entity.TierSilver, TierName: "Silver", TierLevel: 2, TierPrice: 100000, TierCurrency: "IDR"},
		SubscriptionAction:    entity.SubscriptionActionNew,
		SubscriptionDays:      30,
		TransactionStatus:     entity.TransactionStatusPending,
		PaymentStatus:         entity.PaymentStatusPending,
	}, nil).Once()

	usecase := NewSubscriptionUsecase(repositoryContext, repositoryContext, repositoryContext)
	transaction, err := usecase.CreateSubscriptionTransaction(context.Background(), userID, tierID, entity.SubscriptionActionNew, 30)
	if err != nil {
		t.Fatalf("CreateSubscriptionTransaction() error = %v", err)
	}
	if transaction == nil || transaction.ID != existingTransactionID {
		t.Fatal("CreateSubscriptionTransaction() should reuse existing pending transaction")
	}
}

func TestSubscriptionUsecaseRejectsRenewWithoutActiveSubscription(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111102")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByID", mock.Anything, tierID).Return(&entity.TierDetail{ID: tierID, Code: entity.TierSilver, Name: "Silver", Level: 2, Price: 100000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetActiveSubscriptionByUserID", mock.Anything, userID).Return(nil, entity.ErrActiveSubscription).Once()

	usecase := NewSubscriptionUsecase(repositoryContext, repositoryContext, repositoryContext)
	_, err := usecase.CreateSubscriptionTransaction(context.Background(), userID, tierID, entity.SubscriptionActionRenew, 30)
	if err != entity.ErrSubscriptionActionRequiresActive {
		t.Fatalf("CreateSubscriptionTransaction() error = %v, want %v", err, entity.ErrSubscriptionActionRequiresActive)
	}
}

func TestSubscriptionUsecaseRejectsUpgradeWhenTierNotHigher(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111102")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByID", mock.Anything, tierID).Return(&entity.TierDetail{ID: tierID, Code: entity.TierSilver, Name: "Silver", Level: 2, Price: 100000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetActiveSubscriptionByUserID", mock.Anything, userID).Return(&entity.Subscription{UserID: userID, TierID: tierID, TierCode: entity.TierGold, TierSnapshot: entity.TierSnapshot{TierID: tierID, TierCode: entity.TierGold, TierName: "Gold", TierLevel: 3, TierPrice: 150000, TierCurrency: "IDR"}, Status: entity.SubscriptionStatusActive, EndDate: time.Now().Add(time.Hour)}, nil).Once()

	usecase := NewSubscriptionUsecase(repositoryContext, repositoryContext, repositoryContext)
	_, err := usecase.CreateSubscriptionTransaction(context.Background(), userID, tierID, entity.SubscriptionActionUpgrade, 30)
	if err != entity.ErrSubscriptionActionMismatch {
		t.Fatalf("CreateSubscriptionTransaction() error = %v, want %v", err, entity.ErrSubscriptionActionMismatch)
	}
}
