package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

func TestTransactionUsecaseListTransactionsByUserID(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("ListTransactionsByUserID", mock.Anything, userID).Return([]entity.Transaction{{ID: entity.UUID("77777777-7777-4777-8777-777777777777"), UserID: userID}}, nil).Once()

	usecase := NewTransactionUsecase(repositoryContext)
	transactions, err := usecase.ListTransactionsByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListTransactionsByUserID() error = %v", err)
	}
	if len(transactions) != 1 || transactions[0].UserID != userID {
		t.Fatalf("transactions = %#v", transactions)
	}
}

func TestTransactionUsecaseGetTransactionByID(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	transactionID := entity.UUID("77777777-7777-4777-8777-777777777777")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetTransactionByID", mock.Anything, userID, transactionID).Return(&entity.Transaction{ID: transactionID, UserID: userID}, nil).Once()

	usecase := NewTransactionUsecase(repositoryContext)
	transaction, err := usecase.GetTransactionByID(context.Background(), userID, transactionID)
	if err != nil {
		t.Fatalf("GetTransactionByID() error = %v", err)
	}
	if transaction == nil || transaction.ID != transactionID || transaction.UserID != userID {
		t.Fatalf("transaction = %#v", transaction)
	}
}
