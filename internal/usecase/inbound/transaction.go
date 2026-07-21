package inbound

import (
	"context"

	"idas-video/internal/entity"
)

type TransactionReaderUsecase interface {
	ListTransactionsByUserID(ctx context.Context, userID entity.UUID) ([]entity.Transaction, error)
	GetTransactionByID(ctx context.Context, userID entity.UUID, transactionID entity.UUID) (*entity.Transaction, error)
}
