package usecase

import (
	"context"
	"sort"
	"strings"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
	"idas-video/internal/usecase/outbound"
)

const (
	defaultVideoListPage   = 1
	defaultVideoListOffset = 0
	defaultVideoListSortBy = "createdAtDesc"
)

type VideoUsecase struct {
	videos        outbound.VideoRepository
	subscriptions outbound.SubscriptionRepository
	logger        outbound.Logger
}

func NewVideoUsecase(videos outbound.VideoRepository, subscriptions outbound.SubscriptionRepository) *VideoUsecase {
	return NewVideoUsecaseWithLogger(videos, subscriptions, nil)
}

func NewVideoUsecaseWithLogger(videos outbound.VideoRepository, subscriptions outbound.SubscriptionRepository, log outbound.Logger) *VideoUsecase {
	return &VideoUsecase{videos: videos, subscriptions: subscriptions, logger: fallbackLogger(log)}
}

func (usecase *VideoUsecase) ListAccessibleVideos(ctx context.Context, userID entity.UUID, query inbound.VideoListQuery) (*inbound.VideoListResult, error) {
	query = normalizeVideoListQuery(query)
	usecase.logger.Info(ctx, "usecase.video", "list accessible videos started", "video.list.started", outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "page", Value: query.Page}, outbound.LogField{Key: "offset", Value: query.Offset}, outbound.LogField{Key: "sort", Value: query.SortBy})

	subscription, err := usecase.subscriptions.GetActiveSubscriptionByUserID(ctx, userID)
	if err != nil {
		usecase.logger.Warn(ctx, "usecase.video", "active subscription lookup failed", "video.list.subscription_failed", outbound.LogField{Key: "user.id", Value: userID.String()})
		return nil, err
	}

	videos, err := usecase.videos.ListVideos(ctx)
	if err != nil {
		usecase.logger.Error(ctx, "usecase.video", "list videos failed", "video.list.repository_failed", err)
		return nil, err
	}

	accessibleVideos := make([]entity.Video, 0, len(videos))
	for _, video := range videos {
		if entity.CanAccessVideo(subscription.TierCode, video.Category) {
			accessibleVideos = append(accessibleVideos, video)
		}
	}

	sortAccessibleVideos(accessibleVideos, query.SortBy)
	total := len(accessibleVideos)
	items := paginateAccessibleVideos(accessibleVideos, query.Page, query.Offset)

	usecase.logger.Info(ctx, "usecase.video", "list accessible videos completed", "video.list.completed", outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "video.count", Value: len(items)}, outbound.LogField{Key: "video.total", Value: total})
	return &inbound.VideoListResult{
		Items:  items,
		Total:  total,
		Page:   query.Page,
		Offset: query.Offset,
		SortBy: query.SortBy,
	}, nil
}

func (usecase *VideoUsecase) GetAccessibleVideoByID(ctx context.Context, userID entity.UUID, videoID entity.UUID) (*entity.Video, error) {
	usecase.logger.Info(ctx, "usecase.video", "get accessible video started", "video.detail.started", outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "video.id", Value: videoID.String()})

	video, err := usecase.videos.GetVideoByID(ctx, videoID)
	if err != nil {
		usecase.logger.Warn(ctx, "usecase.video", "get video by id failed", "video.detail.lookup_failed", outbound.LogField{Key: "video.id", Value: videoID.String()})
		return nil, err
	}

	if err := usecase.canAccessVideo(ctx, userID, video); err != nil {
		usecase.logger.Warn(ctx, "usecase.video", "video access denied", "video.detail.access_denied", outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "video.id", Value: videoID.String()})
		return nil, err
	}

	usecase.logger.Info(ctx, "usecase.video", "get accessible video completed", "video.detail.completed", outbound.LogField{Key: "user.id", Value: userID.String()}, outbound.LogField{Key: "video.id", Value: videoID.String()})
	return video, nil
}

func (usecase *VideoUsecase) CanAccessVideo(ctx context.Context, userID entity.UUID, videoID entity.UUID) error {
	video, err := usecase.videos.GetVideoByID(ctx, videoID)
	if err != nil {
		return err
	}

	return usecase.canAccessVideo(ctx, userID, video)
}

func (usecase *VideoUsecase) canAccessVideo(ctx context.Context, userID entity.UUID, video *entity.Video) error {
	subscription, err := usecase.subscriptions.GetActiveSubscriptionByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if !entity.CanAccessVideo(subscription.TierCode, video.Category) {
		return entity.ErrForbiddenTier
	}

	return nil
}

func normalizeVideoListQuery(query inbound.VideoListQuery) inbound.VideoListQuery {
	if query.Page <= 0 {
		query.Page = defaultVideoListPage
	}
	if query.Offset < 0 {
		query.Offset = defaultVideoListOffset
	}
	switch query.SortBy {
	case "createdAtAsc", "createdAtDesc", "titleAsc", "titleDesc":
	default:
		query.SortBy = defaultVideoListSortBy
	}
	return query
}

func sortAccessibleVideos(videos []entity.Video, sortBy string) {
	sort.SliceStable(videos, func(i int, j int) bool {
		left := videos[i]
		right := videos[j]
		switch sortBy {
		case "createdAtAsc":
			if left.CreatedAt.Equal(right.CreatedAt) {
				return strings.Compare(left.ID.String(), right.ID.String()) < 0
			}
			return left.CreatedAt.Before(right.CreatedAt)
		case "titleAsc":
			if left.Title == right.Title {
				return strings.Compare(left.ID.String(), right.ID.String()) < 0
			}
			return strings.Compare(strings.ToLower(left.Title), strings.ToLower(right.Title)) < 0
		case "titleDesc":
			if left.Title == right.Title {
				return strings.Compare(left.ID.String(), right.ID.String()) < 0
			}
			return strings.Compare(strings.ToLower(left.Title), strings.ToLower(right.Title)) > 0
		default:
			if left.CreatedAt.Equal(right.CreatedAt) {
				return strings.Compare(left.ID.String(), right.ID.String()) < 0
			}
			return left.CreatedAt.After(right.CreatedAt)
		}
	})
}

func paginateAccessibleVideos(videos []entity.Video, page int, offset int) []entity.Video {
	if len(videos) == 0 {
		return []entity.Video{}
	}
	start := offset
	if start >= len(videos) {
		return []entity.Video{}
	}
	end := start + page
	if end > len(videos) {
		end = len(videos)
	}
	return videos[start:end]
}
