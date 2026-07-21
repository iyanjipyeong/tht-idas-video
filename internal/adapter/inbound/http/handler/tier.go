package handler

import (
	"net/http"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

type tierResponse struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Level       int     `json:"level"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

type TierHandler struct {
	usecase inbound.TierListUsecase
}

func NewTierHandler(usecase inbound.TierListUsecase) *TierHandler {
	return &TierHandler{usecase: usecase}
}

func (handler *TierHandler) ListTiers(writer http.ResponseWriter, request *http.Request) {
	tiers, err := handler.usecase.ListTiers(request.Context())
	if err != nil {
		writeError(writer, err)
		return
	}
	responses := make([]tierResponse, 0, len(tiers))
	for _, tier := range tiers {
		responses = append(responses, newTierResponse(tier))
	}
	writeListSuccess(writer, responses, len(responses), defaultListPage, defaultListOffset, defaultListSortBy)
}

func newTierResponse(tier entity.TierDetail) tierResponse {
	return tierResponse{ID: tier.ID.String(), Code: tier.Code.String(), Name: tier.Name, Level: tier.Level, Price: tier.Price, Currency: tier.Currency, Description: tier.Description}
}
