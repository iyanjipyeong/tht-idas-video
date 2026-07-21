package handler

import (
	"encoding/json"
	"net/http"

	logOption "github.com/digitalrealmforgestudios/d-logger/option"

	"idas-video/internal/adapter/inbound/http/observability"
	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiredAt    int64  `json:"expiredAt"`
}

type AuthHandler struct {
	usecase inbound.LoginUsecase
}

func NewAuthHandler(usecase inbound.LoginUsecase) *AuthHandler { return &AuthHandler{usecase: usecase} }

func (handler *AuthHandler) Login(writer http.ResponseWriter, request *http.Request) {
	log := observability.Child("http.handler.auth")
	log.Info("login request received", observability.WithContext(request.Context()), logOption.EventName("http.auth.login.started"))

	var payload loginPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		log.Warn("login payload invalid", observability.WithContext(request.Context()), logOption.Error(err))
		writeError(writer, entity.ErrInvalidRequest)
		return
	}

	authUser, err := handler.usecase.Login(request.Context(), payload.Email, payload.Password)
	if err != nil {
		log.Warn("login failed", observability.WithContext(request.Context()), logOption.Error(err), logOption.Attribute("user.email", payload.Email))
		writeError(writer, err)
		return
	}

	log.Info("login succeeded", observability.WithContext(request.Context()), logOption.Attribute("user.id", authUser.ID.String()), logOption.Attribute("user.email", authUser.Email))
	writeSuccess(writer, authResponse{
		AccessToken:  authUser.AccessToken,
		RefreshToken: authUser.RefreshToken,
		ExpiredAt:    authUser.ExpiredAt.UTC().Unix(),
	})
}
