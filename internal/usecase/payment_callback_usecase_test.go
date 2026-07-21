package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
	"idas-video/internal/usecase/outbound"
)

func TestPaymentCallbackUsecaseProcessPaymentCallbackPaidSuccess(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111103")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByCode", mock.Anything, entity.TierGold).Return(&entity.TierDetail{ID: tierID, Code: entity.TierGold, Name: "Gold", Level: 3, Price: 150000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetTransactionByExternalID", mock.Anything, "trx-001").Return(&entity.Transaction{ExternalTransactionID: "trx-001", OrderID: "ORDER-TRX-001", UserID: userID, TierID: tierID, TierCode: entity.TierGold, TierSnapshot: entity.TierSnapshot{TierID: tierID, TierCode: entity.TierGold, TierName: "Gold", TierLevel: 3, TierPrice: 150000, TierCurrency: "IDR"}, SubscriptionAction: entity.SubscriptionActionNew, SubscriptionDays: 30, GrossAmount: 150000, FinalAmount: 150000, TransactionStatus: entity.TransactionStatusPending, PaymentStatus: entity.PaymentStatusPending}, nil).Once()
	repositoryContext.On("ActivateSubscription", mock.Anything, userID, entity.TierGold, 30).Return(&outbound.SubscriptionActivationResult{
		Action: entity.SubscriptionActionNew,
		Subscription: &entity.Subscription{
			ID:           entity.UUID("55555555-5555-4555-8555-555555555555"),
			UserID:       userID,
			TierID:       tierID,
			TierCode:     entity.TierGold,
			TierSnapshot: entity.TierSnapshot{TierID: tierID, TierCode: entity.TierGold, TierName: "Gold", TierLevel: 3, TierPrice: 150000, TierCurrency: "IDR"},
		},
		FinalAmount: 150000,
	}, nil).Once()
	repositoryContext.On("UpdateTransaction", mock.Anything, mock.MatchedBy(func(transaction *entity.Transaction) bool {
		return transaction != nil && transaction.ExternalTransactionID == "trx-001" && transaction.SubscriptionID == entity.UUID("55555555-5555-4555-8555-555555555555") && transaction.TransactionStatus == entity.TransactionStatusProcessed
	})).Return(nil).Once()

	usecase := NewPaymentCallbackUsecase(repositoryContext, repositoryContext, repositoryContext)
	err := usecase.ProcessPaymentCallback(context.Background(), inbound.PaymentCallbackRequest{TransactionID: "trx-001", OrderID: "ORDER-TRX-001", UserID: userID, Tier: entity.TierGold, PaymentStatus: entity.PaymentStatusPaid, GrossAmount: 150000, SubscriptionDays: 30, RawPayload: []byte(`{"transaction_id":"trx-001"}`)})
	if err != nil {
		t.Fatalf("ProcessPaymentCallback() error = %v", err)
	}
}

func TestPaymentCallbackUsecaseRejectsDuplicateTransaction(t *testing.T) {
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	tierID := entity.UUID("11111111-1111-4111-8111-111111111103")
	repositoryContext.On("GetTierByCode", mock.Anything, entity.TierGold).Return(&entity.TierDetail{ID: tierID, Code: entity.TierGold, Name: "Gold", Level: 3, Price: 150000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetTransactionByExternalID", mock.Anything, "trx-001").Return(nil, entity.ErrTransactionNotFound).Once()

	usecase := NewPaymentCallbackUsecase(repositoryContext, repositoryContext, repositoryContext)
	err := usecase.ProcessPaymentCallback(context.Background(), inbound.PaymentCallbackRequest{TransactionID: "trx-001", OrderID: "ORDER-TRX-001", UserID: entity.UUID("11111111-1111-4111-8111-111111111111"), Tier: entity.TierGold, PaymentStatus: entity.PaymentStatusPaid, SubscriptionDays: 30})
	if err != entity.ErrTransactionMismatch {
		t.Fatalf("ProcessPaymentCallback() error = %v, want %v", err, entity.ErrTransactionMismatch)
	}
}

func TestPaymentCallbackUsecaseProcessesReservedTransaction(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111103")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByCode", mock.Anything, entity.TierGold).Return(&entity.TierDetail{ID: tierID, Code: entity.TierGold, Name: "Gold", Level: 3, Price: 150000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetTransactionByExternalID", mock.Anything, "sub-001").Return(&entity.Transaction{
		ID:                    entity.UUID("77777777-7777-4777-8777-777777777777"),
		ExternalTransactionID: "sub-001",
		OrderID:               "ORDER-SUB-001",
		UserID:                userID,
		TierID:                tierID,
		TierCode:              entity.TierGold,
		TierSnapshot:          entity.TierSnapshot{TierID: tierID, TierCode: entity.TierGold, TierName: "Gold", TierLevel: 3, TierPrice: 150000, TierCurrency: "IDR"},
		SubscriptionAction:    entity.SubscriptionActionNew,
		SubscriptionDays:      30,
		FinalAmount:           150000,
		GrossAmount:           150000,
		Currency:              "IDR",
		TransactionStatus:     entity.TransactionStatusPending,
		PaymentStatus:         entity.PaymentStatusPending,
	}, nil).Once()
	repositoryContext.On("ActivateSubscription", mock.Anything, userID, entity.TierGold, 30).Return(&outbound.SubscriptionActivationResult{
		Action: entity.SubscriptionActionNew,
		Subscription: &entity.Subscription{
			ID:           entity.UUID("55555555-5555-4555-8555-555555555555"),
			UserID:       userID,
			TierID:       tierID,
			TierCode:     entity.TierGold,
			TierSnapshot: entity.TierSnapshot{TierID: tierID, TierCode: entity.TierGold, TierName: "Gold", TierLevel: 3, TierPrice: 150000, TierCurrency: "IDR"},
		},
		FinalAmount: 150000,
	}, nil).Once()
	repositoryContext.On("UpdateTransaction", mock.Anything, mock.MatchedBy(func(transaction *entity.Transaction) bool {
		return transaction != nil && transaction.ExternalTransactionID == "sub-001" && transaction.GatewayTransactionID == "gw-001" && transaction.TransactionStatus == entity.TransactionStatusProcessed && transaction.PaymentStatus == entity.PaymentStatusPaid
	})).Return(nil).Once()

	usecase := NewPaymentCallbackUsecase(repositoryContext, repositoryContext, repositoryContext)
	err := usecase.ProcessPaymentCallback(context.Background(), inbound.PaymentCallbackRequest{TransactionID: "sub-001", GatewayTransactionID: "gw-001", OrderID: "ORDER-SUB-001", UserID: userID, Tier: entity.TierGold, PaymentStatus: entity.PaymentStatusPaid, GrossAmount: 150000, SubscriptionDays: 30, RawPayload: []byte(`{"transaction_id":"sub-001"}`)})
	if err != nil {
		t.Fatalf("ProcessPaymentCallback() error = %v", err)
	}
}

func TestPaymentCallbackUsecaseRejectsReservedTransactionMismatch(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111103")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByCode", mock.Anything, entity.TierGold).Return(&entity.TierDetail{ID: tierID, Code: entity.TierGold, Name: "Gold", Level: 3, Price: 150000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetTransactionByExternalID", mock.Anything, "sub-001").Return(&entity.Transaction{ExternalTransactionID: "sub-001", OrderID: "ORDER-SUB-001", UserID: userID, TierID: tierID, SubscriptionDays: 30, GrossAmount: 150000, TransactionStatus: entity.TransactionStatusPending}, nil).Once()

	usecase := NewPaymentCallbackUsecase(repositoryContext, repositoryContext, repositoryContext)
	err := usecase.ProcessPaymentCallback(context.Background(), inbound.PaymentCallbackRequest{TransactionID: "sub-001", UserID: userID, Tier: entity.TierGold, OrderID: "OTHER-ORDER", PaymentStatus: entity.PaymentStatusPaid, GrossAmount: 150000, SubscriptionDays: 30})
	if err != entity.ErrTransactionMismatch {
		t.Fatalf("ProcessPaymentCallback() error = %v, want %v", err, entity.ErrTransactionMismatch)
	}
}

func TestPaymentCallbackUsecaseRejectsReservedTransactionWithoutMatchingOrderID(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111103")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByCode", mock.Anything, entity.TierGold).Return(&entity.TierDetail{ID: tierID, Code: entity.TierGold, Name: "Gold", Level: 3, Price: 150000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetTransactionByExternalID", mock.Anything, "sub-001").Return(&entity.Transaction{ExternalTransactionID: "sub-001", OrderID: "ORDER-SUB-001", UserID: userID, TierID: tierID, SubscriptionDays: 30, GrossAmount: 150000, TransactionStatus: entity.TransactionStatusPending}, nil).Once()

	usecase := NewPaymentCallbackUsecase(repositoryContext, repositoryContext, repositoryContext)
	err := usecase.ProcessPaymentCallback(context.Background(), inbound.PaymentCallbackRequest{TransactionID: "sub-001", UserID: userID, Tier: entity.TierGold, PaymentStatus: entity.PaymentStatusPaid, GrossAmount: 150000, SubscriptionDays: 30})
	if err != entity.ErrTransactionMismatch {
		t.Fatalf("ProcessPaymentCallback() error = %v, want %v", err, entity.ErrTransactionMismatch)
	}
}

func TestPaymentCallbackUsecaseRejectsReservedTransactionWithDifferentGrossAmount(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	tierID := entity.UUID("11111111-1111-4111-8111-111111111103")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTierByCode", mock.Anything, entity.TierGold).Return(&entity.TierDetail{ID: tierID, Code: entity.TierGold, Name: "Gold", Level: 3, Price: 150000, Currency: "IDR"}, nil).Once()
	repositoryContext.On("GetTransactionByExternalID", mock.Anything, "sub-001").Return(&entity.Transaction{ExternalTransactionID: "sub-001", OrderID: "ORDER-SUB-001", UserID: userID, TierID: tierID, SubscriptionDays: 30, GrossAmount: 150000, TransactionStatus: entity.TransactionStatusPending}, nil).Once()

	usecase := NewPaymentCallbackUsecase(repositoryContext, repositoryContext, repositoryContext)
	err := usecase.ProcessPaymentCallback(context.Background(), inbound.PaymentCallbackRequest{TransactionID: "sub-001", OrderID: "ORDER-SUB-001", UserID: userID, Tier: entity.TierGold, PaymentStatus: entity.PaymentStatusPaid, GrossAmount: 140000, SubscriptionDays: 30})
	if err != entity.ErrTransactionMismatch {
		t.Fatalf("ProcessPaymentCallback() error = %v, want %v", err, entity.ErrTransactionMismatch)
	}
}
