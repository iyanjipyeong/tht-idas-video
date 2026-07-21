package usecase

import (
	"context"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

type TransactionUsecase struct {
	transactions outbound.TransactionRepository
}

func NewTransactionUsecase(transactions outbound.TransactionRepository) *TransactionUsecase {
	return &TransactionUsecase{transactions: transactions}
}

func (usecase *TransactionUsecase) ListTransactionsByUserID(ctx context.Context, userID entity.UUID) ([]entity.Transaction, error) {
	return usecase.transactions.ListTransactionsByUserID(ctx, userID)
}

func (usecase *TransactionUsecase) GetTransactionByID(ctx context.Context, userID entity.UUID, transactionID entity.UUID) (*entity.Transaction, error) {
	return usecase.transactions.GetTransactionByID(ctx, userID, transactionID)
}
