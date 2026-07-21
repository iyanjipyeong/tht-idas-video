package inbound

import (
	"context"

	"idas-video/internal/entity"
)

type TierListUsecase interface {
	ListTiers(ctx context.Context) ([]entity.TierDetail, error)
}
