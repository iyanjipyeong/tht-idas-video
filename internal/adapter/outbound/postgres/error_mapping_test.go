package postgres

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"idas-video/internal/entity"
)

func TestMapConstraintErrorMapsTransactionUserForeignKey(t *testing.T) {
	err := mapConstraintError(&pgconn.PgError{Code: "23503", ConstraintName: "transactions_user_id_fkey"})
	if err != entity.ErrUserNotFound {
		t.Fatalf("mapConstraintError() = %v, want %v", err, entity.ErrUserNotFound)
	}
}

func TestMapConstraintErrorMapsTransactionTierForeignKey(t *testing.T) {
	err := mapConstraintError(&pgconn.PgError{Code: "23503", ConstraintName: "transactions_tier_id_fkey"})
	if err != entity.ErrInvalidSubscription {
		t.Fatalf("mapConstraintError() = %v, want %v", err, entity.ErrInvalidSubscription)
	}
}
