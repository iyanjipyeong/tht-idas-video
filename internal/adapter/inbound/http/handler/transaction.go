package handler

import (
	"net/http"
	"time"

	"idas-video/internal/adapter/inbound/http/middleware"
	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

type transactionResponse struct {
	ID                    string  `json:"id"`
	ExternalTransactionID string  `json:"externalTransactionId"`
	GatewayTransactionID  string  `json:"gatewayTransactionId"`
	OrderID               string  `json:"orderId"`
	UserID                string  `json:"userId"`
	SubscriptionID        string  `json:"subscriptionId"`
	TierID                string  `json:"tierId"`
	TierCode              string  `json:"tierCode"`
	TierName              string  `json:"tierName"`
	TierLevel             int     `json:"tierLevel"`
	TierPrice             float64 `json:"tierPrice"`
	TierCurrency          string  `json:"tierCurrency"`
	SubscriptionAction    string  `json:"subscriptionAction"`
	SubscriptionDays      int     `json:"subscriptionDays"`
	CurrentSubscriptionID string  `json:"currentSubscriptionId"`
	CurrentTierID         string  `json:"currentTierId"`
	CurrentTierCode       string  `json:"currentTierCode"`
	CurrentTierName       string  `json:"currentTierName"`
	CurrentTierLevel      int     `json:"currentTierLevel"`
	CurrentTierPrice      float64 `json:"currentTierPrice"`
	CurrentEndDate        *int64  `json:"currentEndDate"`
	ProratedCredit        float64 `json:"proratedCredit"`
	FinalAmount           float64 `json:"finalAmount"`
	GrossAmount           float64 `json:"grossAmount"`
	Currency              string  `json:"currency"`
	PaymentType           string  `json:"paymentType"`
	TransactionStatus     string  `json:"transactionStatus"`
	PaymentStatus         string  `json:"paymentStatus"`
	FraudStatus           string  `json:"fraudStatus"`
	SignatureKey          string  `json:"signatureKey"`
	TransactionTime       *int64  `json:"transactionTime"`
	SettlementTime        *int64  `json:"settlementTime"`
	ExpiryTime            *int64  `json:"expiryTime"`
	SettlementTimeRaw     string  `json:"settlementTimeRaw"`
	CreatedAt             int64   `json:"createdAt"`
	UpdatedAt             int64   `json:"updatedAt"`
}

type TransactionHandler struct {
	usecase inbound.TransactionReaderUsecase
}

func NewTransactionHandler(usecase inbound.TransactionReaderUsecase) *TransactionHandler {
	return &TransactionHandler{usecase: usecase}
}

func (handler *TransactionHandler) ListTransactions(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		writeError(writer, ErrHTTPUnauthorized)
		return
	}

	transactions, err := handler.usecase.ListTransactionsByUserID(request.Context(), userID)
	if err != nil {
		writeError(writer, err)
		return
	}
	responses := make([]transactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		responses = append(responses, newTransactionResponse(transaction))
	}
	writeListSuccess(writer, responses, len(responses), defaultListPage, defaultListOffset, defaultListSortBy)
}

func (handler *TransactionHandler) GetTransaction(writer http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		writeError(writer, ErrHTTPUnauthorized)
		return
	}

	transactionID := request.PathValue("id")
	if !entity.IsUUID(transactionID) {
		writeError(writer, entity.ErrInvalidRequest)
		return
	}

	transaction, err := handler.usecase.GetTransactionByID(request.Context(), userID, entity.UUID(transactionID))
	if err != nil {
		writeError(writer, err)
		return
	}

	writeSuccess(writer, newTransactionResponse(*transaction))
}

func newTransactionResponse(transaction entity.Transaction) transactionResponse {
	response := transactionResponse{
		ID:                    transaction.ID.String(),
		ExternalTransactionID: transaction.ExternalTransactionID,
		GatewayTransactionID:  transaction.GatewayTransactionID,
		OrderID:               transaction.OrderID,
		UserID:                transaction.UserID.String(),
		SubscriptionID:        transaction.SubscriptionID.String(),
		TierID:                transaction.TierID.String(),
		TierCode:              transaction.TierCode.String(),
		TierName:              transaction.TierSnapshot.TierName,
		TierLevel:             transaction.TierSnapshot.TierLevel,
		TierPrice:             transaction.TierSnapshot.TierPrice,
		TierCurrency:          transaction.TierSnapshot.TierCurrency,
		SubscriptionAction:    string(transaction.SubscriptionAction),
		SubscriptionDays:      transaction.SubscriptionDays,
		CurrentSubscriptionID: transaction.CurrentSubscriptionID.String(),
		ProratedCredit:        transaction.ProratedCredit,
		FinalAmount:           transaction.FinalAmount,
		GrossAmount:           transaction.GrossAmount,
		Currency:              transaction.Currency,
		PaymentType:           transaction.PaymentType,
		TransactionStatus:     string(transaction.TransactionStatus),
		PaymentStatus:         transaction.PaymentStatus,
		FraudStatus:           transaction.FraudStatus,
		SignatureKey:          transaction.SignatureKey,
		SettlementTimeRaw:     transaction.SettlementTimeRaw,
		CreatedAt:             transaction.CreatedAt.UTC().Unix(),
		UpdatedAt:             transaction.UpdatedAt.UTC().Unix(),
	}
	if transaction.CurrentTierSnapshot != nil {
		response.CurrentTierID = transaction.CurrentTierSnapshot.TierID.String()
		response.CurrentTierCode = transaction.CurrentTierSnapshot.TierCode.String()
		response.CurrentTierName = transaction.CurrentTierSnapshot.TierName
		response.CurrentTierLevel = transaction.CurrentTierSnapshot.TierLevel
		response.CurrentTierPrice = transaction.CurrentTierSnapshot.TierPrice
	}
	response.CurrentEndDate = unixPtr(transaction.CurrentEndDate)
	response.TransactionTime = unixPtr(transaction.TransactionTime)
	response.SettlementTime = unixPtr(transaction.SettlementTime)
	response.ExpiryTime = unixPtr(transaction.ExpiryTime)
	return response
}

func unixPtr(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	unix := value.UTC().Unix()
	return &unix
}
