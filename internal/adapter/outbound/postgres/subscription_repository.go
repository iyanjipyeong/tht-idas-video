package postgres

import (
	"context"

	"gorm.io/gorm/clause"

	"idas-video/internal/entity"
)

func (store *Store) GetActiveSubscriptionByUserID(ctx context.Context, userID entity.UUID) (*entity.Subscription, error) {
	var model subscriptionModel
	err := store.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND end_date > NOW()", userID, entity.SubscriptionStatusActive).
		Order("end_date DESC").
		First(&model).Error
	if err != nil {
		return nil, mapRecordNotFound(err, entity.ErrActiveSubscription)
	}

	subscription := model.toDomain()
	return &subscription, nil
}

func (store *Store) UpsertSubscription(ctx context.Context, subscription *entity.Subscription) error {
	model := newSubscriptionModel(subscription)
	return store.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Eq{
				Column: clause.Column{Name: "status"},
				Value:  entity.SubscriptionStatusActive,
			}}},
			DoUpdates: clause.AssignmentColumns([]string{
				"tier_id",
				"tier_name",
				"tier_level",
				"tier_price",
				"tier_currency",
				"status",
				"start_date",
				"end_date",
				"updated_at",
			}),
		}).
		Create(&model).Error
}
