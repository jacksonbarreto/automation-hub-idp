package iservice

import "idp-automations-hub/internal/app/dto"

type AuthService interface {
	Register(email, password string) (*dto.TokenDetails, error)
	Login(email, password string) (*dto.TokenDetails, error)
	Logout(accessToken string) error
	RefreshToken(refreshToken string) (*dto.TokenDetails, error)
	VerifyToken(token string) (*dto.TokenDetails, error)
	RequestPasswordReset(email string) error
	ConfirmPasswordReset(token, newPassword string) error
}
