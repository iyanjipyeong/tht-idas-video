package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"idas-video/internal/entity"
	"idas-video/internal/usecase/outbound"
)

func TestAuthUsecaseLoginSuccess(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetUserByEmail", mock.Anything, "demo@example.com").Return(&entity.User{
		ID:       entity.UUID("11111111-1111-4111-8111-111111111111"),
		Email:    "demo@example.com",
		Password: string(hashedPassword),
	}, nil).Once()

	usecase := NewAuthUsecase(repositoryContext, "secret-key")
	authUser, err := usecase.Login(context.Background(), "demo@example.com", "password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if authUser == nil || authUser.AccessToken == "" || authUser.RefreshToken == "" {
		t.Fatal("Login() should return access token and refresh token")
	}
}

func TestAuthUsecaseAuthenticateAccessTokenSuccess(t *testing.T) {
	userID := entity.UUID("11111111-1111-4111-8111-111111111111")
	repositoryContext := outbound.NewMockIRepositoryContext(t)
	repositoryContext.On("GetUserByID", mock.Anything, userID).Return(&entity.User{ID: userID, Email: "demo@example.com"}, nil).Once()

	usecase := NewAuthUsecase(repositoryContext, "secret-key")
	authUser, err := usecase.issueAuthUser(userID, "demo@example.com")
	if err != nil {
		t.Fatalf("issueAuthUser() error = %v", err)
	}

	authenticatedUserID, err := usecase.AuthenticateAccessToken(context.Background(), authUser.AccessToken)
	if err != nil {
		t.Fatalf("AuthenticateAccessToken() error = %v", err)
	}
	if authenticatedUserID != userID {
		t.Fatalf("AuthenticateAccessToken() userID = %q, want %q", authenticatedUserID, userID)
	}
}
