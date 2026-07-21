package entity

import "time"

type AuthUser struct {
	ID           UUID      `json:"id"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiredAt    time.Time `json:"expiredAt"`
}
