package irepository

import (
	"github.com/google/uuid"
	"idp-automations-hub/internal/app/models"
	"idp-automations-hub/internal/app/utils"
)

type UserRepository interface {
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Create(user *models.User) (*models.User, error)
	Update(user *models.User) (*models.User, error)
	Delete(id uuid.UUID) error
	FindAll(p utils.Pagination) ([]*models.User, error)
	FindByResetToken(token string) (*models.User, error)
}
