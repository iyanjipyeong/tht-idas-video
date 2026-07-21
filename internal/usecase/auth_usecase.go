package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

type AuthUsecase struct {
	users       outbound.UserRepository
	logger      outbound.Logger
	tokenSecret []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

type tokenPayload struct {
	Sub string    `json:"sub"`
	Em  string    `json:"em"`
	Typ string    `json:"typ"`
	Exp time.Time `json:"exp"`
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

func NewAuthUsecase(users outbound.UserRepository, tokenSecret string) *AuthUsecase {
	return NewAuthUsecaseWithLogger(users, tokenSecret, nil)
}

func NewAuthUsecaseWithLogger(users outbound.UserRepository, tokenSecret string, log outbound.Logger) *AuthUsecase {
	return &AuthUsecase{
		users:       users,
		logger:      fallbackLogger(log),
		tokenSecret: []byte(tokenSecret),
		accessTTL:   24 * time.Hour,
		refreshTTL:  7 * 24 * time.Hour,
	}
}

func (usecase *AuthUsecase) Login(ctx context.Context, email string, password string) (*entity.AuthUser, error) {
	if strings.TrimSpace(email) == entity.EmptyString || strings.TrimSpace(password) == entity.EmptyString {
		usecase.logger.Warn(ctx, "usecase.auth", "login rejected due to empty credentials", "auth.login.rejected")
		return nil, entity.ErrInvalidRequest
	}

	usecase.logger.Info(ctx, "usecase.auth", "auth login started", "auth.login.started", outbound.LogField{Key: "user.email", Value: email})
	user, err := usecase.users.GetUserByEmail(ctx, email)
	if err != nil {
		usecase.logger.Warn(ctx, "usecase.auth", "user lookup failed during login", "auth.login.user_lookup_failed", outbound.LogField{Key: "user.email", Value: email})
		return nil, entity.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		usecase.logger.Warn(ctx, "usecase.auth", "login password verification failed", "auth.login.password_mismatch", outbound.LogField{Key: "user.id", Value: user.ID.String()})
		return nil, entity.ErrUnauthorized
	}

	authUser, err := usecase.issueAuthUser(user.ID, user.Email)
	if err != nil {
		usecase.logger.Error(ctx, "usecase.auth", "auth token issuance failed", "auth.login.token_issue_failed", err, outbound.LogField{Key: "user.id", Value: user.ID.String()})
		return nil, err
	}

	usecase.logger.Info(ctx, "usecase.auth", "auth login completed", "auth.login.completed", outbound.LogField{Key: "user.id", Value: user.ID.String()})
	return authUser, nil
}

func (usecase *AuthUsecase) AuthenticateUser(ctx context.Context, userID entity.UUID) error {
	if userID == entity.EmptyString {
		usecase.logger.Warn(ctx, "usecase.auth", "authenticate user rejected with empty user id", "auth.user_auth.rejected")
		return entity.ErrUnauthorized
	}

	_, err := usecase.users.GetUserByID(ctx, userID)
	if err != nil {
		usecase.logger.Warn(ctx, "usecase.auth", "authenticate user failed", "auth.user_auth.failed", outbound.LogField{Key: "user.id", Value: userID.String()})
		return entity.ErrUnauthorized
	}

	return nil
}

func (usecase *AuthUsecase) AuthenticateAccessToken(ctx context.Context, accessToken string) (entity.UUID, error) {
	claims, err := usecase.verifyToken(accessToken, entity.JWTPurposeAccess)
	if err != nil {
		usecase.logger.Warn(ctx, "usecase.auth", "access token verification failed", "auth.access_token.invalid")
		return "", entity.ErrUnauthorized
	}

	if _, err := usecase.users.GetUserByID(ctx, entity.UUID(claims.Sub)); err != nil {
		usecase.logger.Warn(ctx, "usecase.auth", "access token user lookup failed", "auth.access_token.user_not_found", outbound.LogField{Key: "user.id", Value: claims.Sub})
		return "", entity.ErrUnauthorized
	}

	return entity.UUID(claims.Sub), nil
}

func (usecase *AuthUsecase) AuthenticateRefreshToken(ctx context.Context, refreshToken string) (entity.UUID, error) {
	claims, err := usecase.verifyToken(refreshToken, entity.JWTPurposeRefresh)
	if err != nil {
		usecase.logger.Warn(ctx, "usecase.auth", "refresh token verification failed", "auth.refresh_token.invalid")
		return "", entity.ErrUnauthorized
	}

	if _, err := usecase.users.GetUserByID(ctx, entity.UUID(claims.Sub)); err != nil {
		usecase.logger.Warn(ctx, "usecase.auth", "refresh token user lookup failed", "auth.refresh_token.user_not_found", outbound.LogField{Key: "user.id", Value: claims.Sub})
		return "", entity.ErrUnauthorized
	}

	return entity.UUID(claims.Sub), nil
}

func (usecase *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*entity.AuthUser, error) {
	usecase.logger.Info(ctx, "usecase.auth", "token refresh started", "auth.refresh.started")
	userID, err := usecase.AuthenticateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := usecase.users.GetUserByID(ctx, userID)
	if err != nil {
		usecase.logger.Warn(ctx, "usecase.auth", "refresh user lookup failed", "auth.refresh.user_lookup_failed", outbound.LogField{Key: "user.id", Value: userID.String()})
		return nil, entity.ErrUnauthorized
	}

	authUser, err := usecase.issueAuthUser(user.ID, user.Email)
	if err != nil {
		usecase.logger.Error(ctx, "usecase.auth", "refresh token issuance failed", "auth.refresh.issue_failed", err, outbound.LogField{Key: "user.id", Value: user.ID.String()})
		return nil, err
	}

	usecase.logger.Info(ctx, "usecase.auth", "token refresh completed", "auth.refresh.completed", outbound.LogField{Key: "user.id", Value: user.ID.String()})
	return authUser, nil
}

func (usecase *AuthUsecase) issueAuthUser(userID entity.UUID, email string) (*entity.AuthUser, error) {
	accessToken, accessExp, err := usecase.signToken(userID, email, entity.JWTPurposeAccess, usecase.accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := usecase.signToken(userID, email, entity.JWTPurposeRefresh, usecase.refreshTTL)
	if err != nil {
		return nil, err
	}

	return &entity.AuthUser{
		ID:           userID,
		Email:        email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiredAt:    accessExp,
	}, nil
}

func (usecase *AuthUsecase) signToken(userID entity.UUID, email string, tokenType string, ttl time.Duration) (string, time.Time, error) {
	exp := time.Now().Add(ttl)
	payload := tokenPayload{Sub: userID.String(), Em: email, Typ: tokenType, Exp: exp}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}

	headerBytes, err := json.Marshal(jwtHeader{Algorithm: entity.JWTAlgorithmHS256, Type: entity.JWTTypeBearerToken})
	if err != nil {
		return "", time.Time{}, err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := encodedHeader + "." + encodedPayload
	signature := usecase.sign(signingInput)
	return signingInput + "." + signature, exp, nil
}

func (usecase *AuthUsecase) verifyToken(token string, expectedType string) (*tokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(usecase.sign(signingInput)), []byte(parts[2])) {
		return nil, fmt.Errorf("invalid token signature")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, err
	}

	if header.Algorithm != entity.JWTAlgorithmHS256 || header.Type != entity.JWTTypeBearerToken {
		return nil, fmt.Errorf("invalid token header")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	if payload.Typ != expectedType {
		return nil, fmt.Errorf("invalid token type")
	}
	if time.Now().After(payload.Exp) {
		return nil, fmt.Errorf("token expired")
	}

	return &payload, nil
}

func (usecase *AuthUsecase) sign(message string) string {
	h := hmac.New(sha256.New, usecase.tokenSecret)
	_, _ = h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
