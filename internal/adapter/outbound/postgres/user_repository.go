package postgres

import (
	"context"

	"idas-video/internal/entity"
)

func (store *Store) GetUserByID(ctx context.Context, id entity.UUID) (*entity.User, error) {
	var model userModel
	if err := store.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, mapRecordNotFound(err, entity.ErrUserNotFound)
	}

	user := model.toDomain()
	return &user, nil
}

func (store *Store) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model userModel
	if err := store.db.WithContext(ctx).First(&model, "email = ?", email).Error; err != nil {
		return nil, mapRecordNotFound(err, entity.ErrUserNotFound)
	}

	user := model.toDomain()
	return &user, nil
}
