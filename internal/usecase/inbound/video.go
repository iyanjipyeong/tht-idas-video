package inbound

import (
	"context"

	"idas-video/internal/entity"
)

type VideoListQuery struct {
	Page   int
	Offset int
	SortBy string
}

type VideoListResult struct {
	Items  []entity.Video
	Total  int
	Page   int
	Offset int
	SortBy string
}

type VideoAccessUsecase interface {
	ListAccessibleVideos(ctx context.Context, userID entity.UUID, query VideoListQuery) (*VideoListResult, error)
	GetAccessibleVideoByID(ctx context.Context, userID entity.UUID, videoID entity.UUID) (*entity.Video, error)
	CanAccessVideo(ctx context.Context, userID entity.UUID, videoID entity.UUID) error
}
