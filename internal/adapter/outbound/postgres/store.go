package postgres

import (
	"context"
	"database/sql"
	"errors"

	logOption "github.com/digitalrealmforgestudios/d-logger/option"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"idas-video/internal/entity"
	"idas-video/internal/infrastructure/logger"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func Open(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	log := logger.Child("postgres")
	log.Info("opening gorm connection", logger.WithContext(ctx))
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Error("gorm open failed", logger.WithContext(ctx), logOption.Error(err))
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("extract sql db failed", logger.WithContext(ctx), logOption.Error(err))
		return nil, err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		log.Error("database ping failed", logger.WithContext(ctx), logOption.Error(err))
		return nil, err
	}

	log.Info("database ping succeeded", logger.WithContext(ctx))
	return db, nil
}

func SQLDB(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

func mapRecordNotFound(err error, mapped error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mapped
	}
	if mappedErr := mapConstraintError(err); mappedErr != nil {
		return mappedErr
	}
	return err
}

func mapConstraintError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	if pgErr.Code != "23503" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "transactions_user_id_fkey", "subscriptions_user_id_fkey":
		return entity.ErrUserNotFound
	case "transactions_tier_id_fkey", "subscriptions_tier_id_fkey":
		return entity.ErrInvalidSubscription
	case "transactions_subscription_id_fkey":
		return entity.ErrInvalidSubscription
	default:
		return entity.ErrInvalidRequest
	}
}
