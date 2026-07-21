package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
	"idas-video/internal/usecase/outbound"
)

func TestVideoUsecaseListAccessibleVideos(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	baseTime := time.Unix(1784628000, 0).UTC()
	videos := []entity.Video{
		{ID: entity.UUID("22222222-2222-4222-8222-222222222222"), Title: "Bronze Video", Category: entity.TierBronze, CreatedAt: baseTime.Add(-2 * time.Hour)},
		{ID: entity.UUID("33333333-3333-4333-8333-333333333333"), Title: "Silver Video", Category: entity.TierSilver, CreatedAt: baseTime.Add(-1 * time.Hour)},
		{ID: entity.UUID("44444444-4444-4444-8444-444444444444"), Title: "Gold Video", Category: entity.TierGold, CreatedAt: baseTime},
	}
	subscription := &entity.Subscription{ID: entity.UUID("55555555-5555-4555-8555-555555555555"), UserID: userID, TierID: entity.UUID("11111111-1111-4111-8111-111111111103"), TierCode: entity.TierGold, Status: entity.SubscriptionStatusActive, EndDate: time.Now().Add(time.Hour)}
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("ListVideos", mock.Anything).Return(videos, nil).Once()
	repositoryContext.On("GetActiveSubscriptionByUserID", mock.Anything, userID).Return(subscription, nil).Once()

	usecase := NewVideoUsecase(repositoryContext, repositoryContext)
	result, err := usecase.ListAccessibleVideos(context.Background(), userID, inbound.VideoListQuery{Page: 2, Offset: 1, SortBy: "createdAtDesc"})
	if err != nil {
		t.Fatalf("ListAccessibleVideos() error = %v", err)
	}
	if result.Total != 3 || result.Page != 2 || result.Offset != 1 || result.SortBy != "createdAtDesc" {
		t.Fatalf("metadata = %#v", result)
	}
	if len(result.Items) != 2 {
		t.Fatalf("ListAccessibleVideos() returned %d videos, want 2", len(result.Items))
	}
}
