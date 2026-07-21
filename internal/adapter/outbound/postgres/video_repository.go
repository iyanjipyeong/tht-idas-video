package postgres

import (
	"context"

	"idas-video/internal/entity"
)

func (store *Store) ListVideos(ctx context.Context) ([]entity.Video, error) {
	models := []videoModel{}
	if err := store.db.WithContext(ctx).Order("created_at DESC, id ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	videos := make([]entity.Video, 0, len(models))
	for _, model := range models {
		videos = append(videos, model.toDomain())
	}

	return videos, nil
}

func (store *Store) GetVideoByID(ctx context.Context, id entity.UUID) (*entity.Video, error) {
	var model videoModel
	if err := store.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, mapRecordNotFound(err, entity.ErrVideoNotFound)
	}

	video := model.toDomain()
	return &video, nil
}
