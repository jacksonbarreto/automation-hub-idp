package iservice

import (
	"github.com/google/uuid"
	"idp-automations-hub/internal/app/models"
	"idp-automations-hub/internal/app/utils"
)

type UserService interface {
	CreateUser(user models.User) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByResetToken(token string) (*models.User, error)
	UpdateUser(user models.User) (*models.User, error)
	DeleteUser(id uuid.UUID) error
	GetAllUsers(p *utils.Pagination) ([]*models.User, error)
	UpdatePassword(id uuid.UUID, newPassword string) error
	BlockUser(id uuid.UUID) error
	UnblockUser(id uuid.UUID) error
}
