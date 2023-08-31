package services

import (
	"errors"
	"github.com/google/uuid"
	"idp-automations-hub/internal/app/models"
	"idp-automations-hub/internal/app/repositories/irepository"
	"idp-automations-hub/internal/app/services/iservice"
	"idp-automations-hub/internal/app/utils"
)

type EmailService interface {
	SendEmail(to, subject, body string) error
	SendTemplatedEmail(to, subject, templateName string, data map[string]interface{}) error
}

type userServiceImpl struct {
	userRepo              irepository.UserRepository
	logger                iservice.Logger
	passwordResetTemplate string
}

func NewUserService(repo irepository.UserRepository, logger iservice.Logger) iservice.UserService {
	return &userServiceImpl{
		userRepo: repo,
		logger:   logger,
	}
}

func (s *userServiceImpl) CreateUser(user models.User) (*models.User, error) {
	if existingUser, _ := s.userRepo.FindByEmail(user.Email); existingUser != nil {
		return nil, errors.New("user already exists")
	}

	return s.userRepo.Create(&user)
}

func (s *userServiceImpl) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userServiceImpl) GetAllUsers(p *utils.Pagination) ([]*models.User, error) {
	if p == nil {
		defaultPagination := utils.DefaultPagination()
		p = &defaultPagination
	}

	return s.userRepo.FindAll(*p)
}

func (s *userServiceImpl) UpdateUser(user models.User) (*models.User, error) {
	currentUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return nil, err
	}

	if currentUser.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(user.Email)
		if err == nil && existingUser.ID != user.ID {
			return nil, errors.New("email already exists")
		}
	}
	user.Password = currentUser.Password
	return s.userRepo.Update(&user)
}

func (s *userServiceImpl) DeleteUser(id uuid.UUID) error {
	err := s.userRepo.Delete(id)
	if err != nil {
		s.logger.Error("Error deleting user with ID: %s, %v", id, err)
		return errors.New("error deleting user")
	}
	return nil
}

func (s *userServiceImpl) UpdatePassword(id uuid.UUID, newPassword string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	user.Password = newPassword
	_, err = s.userRepo.Update(user)
	return err
}

func (s *userServiceImpl) BlockUser(id uuid.UUID) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Error finding user with ID: %s, %v", id, err)
		return err
	}

	user.IsBlocked = true

	_, err = s.userRepo.Update(user)
	if err != nil {
		s.logger.Error("Error blocking user with ID: %s, %v", id, err)
	}

	return err
}

func (s *userServiceImpl) UnblockUser(id uuid.UUID) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Error finding user with ID: %s, %v", id, err)
		return err
	}

	user.IsBlocked = false

	_, err = s.userRepo.Update(user)
	if err != nil {
		s.logger.Error("Error unblocking user with ID: %s, %v", id, err)
	}

	return err
}

func (s *userServiceImpl) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		s.logger.Error("Failed to fetch user with email: %s, %v", email, err)
		return nil, errors.New("failed to fetch user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *userServiceImpl) GetUserByResetToken(token string) (*models.User, error) {
	user, err := s.userRepo.FindByResetToken(token)
	if err != nil {
		s.logger.Error("Failed to fetch user with reset token: %s, %v", token, err)
		return nil, errors.New("failed to fetch user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
