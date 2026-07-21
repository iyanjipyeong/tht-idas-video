package usecase

import (
	"context"
	"testing"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

func TestTierUsecaseListTiers(t *testing.T) {
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("ListTiers", context.Background()).Return([]entity.TierDetail{
		{ID: entity.UUID("11111111-1111-4111-8111-111111111101"), Code: entity.TierBronze, Name: "Bronze", Level: 1, Price: 50000, Currency: "IDR"},
		{ID: entity.UUID("11111111-1111-4111-8111-111111111102"), Code: entity.TierSilver, Name: "Silver", Level: 2, Price: 100000, Currency: "IDR"},
	}, nil).Once()

	usecase := NewTierUsecase(repositoryContext)
	tiers, err := usecase.ListTiers(context.Background())
	if err != nil {
		t.Fatalf("ListTiers() error = %v", err)
	}
	if len(tiers) != 2 || tiers[0].Name != "Bronze" || tiers[1].Level != 2 {
		t.Fatalf("tiers = %#v", tiers)
	}
}
