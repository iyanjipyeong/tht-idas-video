package usecase

import (
	"context"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

type TierUsecase struct {
	tiers outbound.TierRepository
}

func NewTierUsecase(tiers outbound.TierRepository) *TierUsecase {
	return &TierUsecase{tiers: tiers}
}

func (usecase *TierUsecase) ListTiers(ctx context.Context) ([]entity.TierDetail, error) {
	return usecase.tiers.ListTiers(ctx)
}
