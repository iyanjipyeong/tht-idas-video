package inbound

import (
	"context"
	"time"

	"idas-video/internal/entity"
)

type UserAuthenticator interface {
	AuthenticateUser(ctx context.Context, userID entity.UUID) error
}

type LoginUsecase interface {
	Login(ctx context.Context, email string, password string) (*entity.AuthUser, error)
	AuthenticateAccessToken(ctx context.Context, accessToken string) (entity.UUID, error)
	AuthenticateRefreshToken(ctx context.Context, refreshToken string) (entity.UUID, error)
	Refresh(ctx context.Context, refreshToken string) (*entity.AuthUser, error)
}

type TokenClaims struct {
	UserID    entity.UUID
	Email     string
	ExpiredAt time.Time
}
