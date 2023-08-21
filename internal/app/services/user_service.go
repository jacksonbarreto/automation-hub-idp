package services

import (
	"errors"
	"github.com/google/uuid"
	"idp-automations-hub/internal/app/models"
	"idp-automations-hub/internal/app/repositories/irepository"
	"idp-automations-hub/internal/app/services/iservice"
	"idp-automations-hub/internal/app/utils"
	"time"
)

type Logger interface {
	Info(message string, args ...interface{})
	Error(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Debug(message string, args ...interface{})
}

type EmailService interface {
	SendEmail(to, subject, body string) error
	SendTemplatedEmail(to, subject, templateName string, data map[string]interface{}) error
}

type userServiceImpl struct {
	userRepo              irepository.UserRepository
	emailService          EmailService
	logger                Logger
	passwordResetTemplate string
	passwordHasher        utils.PasswordHasher
}

func NewUserService(repo irepository.UserRepository, emailService EmailService, logger Logger,
	hasher utils.PasswordHasher, resetTemplate ...string) iservice.UserService {

	defaultTemplate := "Your reset link is: {{.Link}}"
	template := defaultTemplate
	if len(resetTemplate) > 0 {
		template = resetTemplate[0]
	}

	return &userServiceImpl{
		userRepo:              repo,
		emailService:          emailService,
		logger:                logger,
		passwordResetTemplate: template,
		passwordHasher:        hasher,
	}
}

func (s *userServiceImpl) CreateUser(user models.User) (*models.User, error) {
	if existingUser, _ := s.userRepo.FindByEmail(user.Email); existingUser != nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := s.passwordHasher.Hash(user.Password)
	if err != nil {
		s.logger.Error("Error generating hashed password for user with Email: %s, %v", user.Email, err)
		return nil, errors.New("failed to create user due to internal error")
	}

	user.Password = string(hashedPassword)
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

func (s *userServiceImpl) ResetPassword(email string, opts ...utils.PasswordResetOptions) error {
	options := utils.DefaultPasswordResetOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return err
	}

	resetToken := uuid.New().String()
	expirationTime := time.Now().Add(options.TokenExpiry)
	user.ResetPasswordToken = resetToken
	user.ResetTokenExpires = &expirationTime

	if _, err := s.userRepo.Update(user); err != nil {
		return err
	}

	magicLink := options.Domain + options.Endpoint + "?token=" + resetToken

	data := map[string]interface{}{
		"Link": magicLink,
	}
	err = s.emailService.SendTemplatedEmail(email, options.EmailSubject, s.passwordResetTemplate, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *userServiceImpl) VerifyPasswordResetToken(token string) error {
	user, err := s.userRepo.FindByResetToken(token)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("invalid token")
	}

	if user.ResetTokenExpires.Before(time.Now()) {
		return errors.New("token expired")
	}

	return nil
}

func (s *userServiceImpl) ChangePassword(id uuid.UUID, newPassword string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	hashedPassword, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		s.logger.Error("Error generating hashed password for user with ID: %s, %v", id, err)
		return errors.New("failed to change password due to internal error")
	}

	user.Password = string(hashedPassword)
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
