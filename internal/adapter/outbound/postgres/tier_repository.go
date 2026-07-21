package postgres

import (
	"context"

	"idas-video/internal/entity"
)

func (store *Store) ListTiers(ctx context.Context) ([]entity.TierDetail, error) {
	models := []tierModel{}
	if err := store.db.WithContext(ctx).Order("level ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	tiers := make([]entity.TierDetail, 0, len(models))
	for _, model := range models {
		tiers = append(tiers, model.toDomain())
	}
	return tiers, nil
}

func (store *Store) GetTierByID(ctx context.Context, id entity.UUID) (*entity.TierDetail, error) {
	var model tierModel
	if err := store.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, mapRecordNotFound(err, entity.ErrInvalidSubscription)
	}

	tier := model.toDomain()
	return &tier, nil
}

func (store *Store) GetTierByCode(ctx context.Context, code entity.Tier) (*entity.TierDetail, error) {
	var model tierModel
	if err := store.db.WithContext(ctx).First(&model, "name = ?", code.String()).Error; err != nil {
		return nil, mapRecordNotFound(err, entity.ErrInvalidSubscription)
	}

	tier := model.toDomain()
	return &tier, nil
}
