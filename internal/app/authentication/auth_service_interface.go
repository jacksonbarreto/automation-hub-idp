package authentication

import (
	"idp-automations-hub/internal/app/dto"
	"time"
)

type IService interface {
	Register(userDTO dto.UserDTO) (*dto.UserResponse, error)
	Login(email, password string) (*dto.TokenDetails, error)
	Logout(accessToken string) error
	RefreshToken(refreshToken string) (*dto.TokenDetails, error)
	IsUserAuthenticated(accessToken string) (bool, error)
	RequestPasswordReset(email string) (string, time.Time, error)
	ConfirmPasswordReset(token, newPassword string) error
	ChangePassword(email string, newPassword string) error
}
