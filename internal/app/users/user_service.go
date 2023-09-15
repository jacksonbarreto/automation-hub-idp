package users

import (
	"automation-hub-idp/internal/app/models"
	"automation-hub-idp/internal/app/repositories/irepository"
	"automation-hub-idp/internal/app/services/iservice"
	"automation-hub-idp/internal/app/utils"
	"errors"
	"github.com/google/uuid"
)

type userServiceImpl struct {
	userRepo irepository.UserRepository
	logger   iservice.Logger
}

func NewUserService(repo irepository.UserRepository, logger iservice.Logger) UserService {
	return &userServiceImpl{
		userRepo: repo,
		logger:   logger,
	}
}

func (s *userServiceImpl) CreateUser(user models.User) (*models.User, error) {
	if existingUser, _ := s.userRepo.FindByEmail(user.Email); existingUser != nil {
		s.logger.Error("User already exists with email: %s", user.Email)
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
		s.logger.Error("Error fetching user with ID: %s, %v", user.ID, err)
		return nil, errors.New("error fetching user by ID")
	}

	if currentUser.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(user.Email)
		if err == nil && existingUser.ID != user.ID {
			s.logger.Error("Email already exists: %s", user.Email)
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
		s.logger.Error("Error fetching user with ID: %s, %v", id, err)
		return errors.New("error fetching user")
	}

	user.Password = newPassword
	user.FirstAccess = false
	_, err = s.userRepo.Update(user)
	if err != nil {
		s.logger.Error("Error updating user with ID: %s, %v", id, err)
		return errors.New("error updating user")
	}
	return nil
}

func (s *userServiceImpl) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		s.logger.Error("Failed to fetch user with email: %s, %v", email, err)
		return nil, errors.New("failed to fetch user")
	}

	if user == nil {
		s.logger.Error("User not found with email: %s", email)
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
		s.logger.Error("User not found with reset token: %s", token)
		return nil, errors.New("user not found")
	}

	return user, nil
}
