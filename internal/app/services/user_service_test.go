package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idp-automations-hub/internal/app/models"
	"idp-automations-hub/internal/app/utils"
	"testing"
	"time"
)

func TestCreateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockHasher := new(MockPasswordHasher)
	mockLogger := new(MockLogger)
	email := "test@example.com"
	user := models.User{Email: email, Password: "test123"}
	hashedPassword := "hashedPassword"

	mockError := errors.New("user not found")
	mockRepo.On("FindByEmail", email).Return(nil, mockError)

	mockHasher.On("Hash", "test123").Return(hashedPassword, nil)
	expectedUser := user
	expectedUser.Password = hashedPassword
	mockRepo.On("Create", &expectedUser).Return(&expectedUser, nil)

	service := NewUserService(mockRepo, nil, mockLogger, mockHasher)

	// Act
	result, err := service.CreateUser(user)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, hashedPassword, result.Password)
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

func TestGetUserByID(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	id := uuid.New()
	user := models.User{ID: id, Email: "test@example.com"}

	mockRepo.On("FindByID", id).Return(&user, nil)

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	result, err := service.GetUserByID(id)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, &user, result)
	mockRepo.AssertExpectations(t)
}

func TestGetAllUsers_WithDefaultPagination(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	users := []*models.User{
		{Email: "test1@example.com"},
		{Email: "test2@example.com"},
	}

	defaultPagination := utils.DefaultPagination()
	mockRepo.On("FindAll", defaultPagination).Return(users, nil)

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	result, err := service.GetAllUsers(nil)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, users, result)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userID := uuid.New()
	existingUser := &models.User{ID: userID, Email: "existing@example.com", Password: "hashedPassword"}
	newUser := models.User{ID: userID, Email: "new@example.com"}

	mockRepo.On("FindByID", userID).Return(existingUser, nil)
	mockRepo.On("FindByEmail", "new@example.com").Return(&models.User{ID: uuid.New()}, nil) // Different user with same email

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	result, err := service.UpdateUser(newUser)

	// Assert
	assert.Nil(t, result)
	assert.EqualError(t, err, "email already exists")
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userID := uuid.New()
	existingUser := &models.User{ID: userID, Email: "existing@example.com", Password: "hashedPassword"}
	newUser := models.User{ID: userID, Email: "new@example.com", Password: "hashedPassword"}

	mockRepo.On("FindByID", userID).Return(existingUser, nil)
	mockRepo.On("FindByEmail", "new@example.com").Return(nil, errors.New("not found"))
	updatedUser := newUser
	updatedUser.Password = "hashedPassword"
	mockRepo.On("Update", &updatedUser).Return(&updatedUser, nil)

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	result, err := service.UpdateUser(newUser)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, &newUser, result)
	assert.Equal(t, "hashedPassword", result.Password)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_PasswordNotChanged(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userID := uuid.New()
	existingPassword := "hashedPassword"
	existingUser := &models.User{ID: userID, Email: "existing@example.com", Password: existingPassword}
	newUser := models.User{ID: userID, Email: "new@example.com", Password: "newHashedPassword"}

	mockRepo.On("FindByID", userID).Return(existingUser, nil)
	mockRepo.On("FindByEmail", "new@example.com").Return(nil, errors.New("not found"))
	var updatedUserToReturn *models.User

	mockRepo.On("Update", mock.AnythingOfType("*models.User")).Run(func(args mock.Arguments) {
		updatedUser := args.Get(0).(*models.User)
		assert.Equal(t, existingPassword, updatedUser.Password)
		updatedUserToReturn = updatedUser
	}).Return(func(args mock.Arguments) (*models.User, error) {
		return updatedUserToReturn, nil
	})
	mockLogger := new(MockLogger)
	mockLogger.On("Error", "Error deleting user with ID: %s, %v", mock.MatchedBy(func(args []interface{}) bool {
		// You can add further conditions to verify the contents of the slice if necessary.
		return true
	})).Return()

	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	result, err := service.UpdateUser(newUser)

	// Assert
	assert.Nil(t, err)
	newUser.Password = existingPassword
	assert.Equal(t, &newUser, result)

	assert.Equal(t, existingPassword, result.Password) // This ensures the password did not change
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	id := uuid.New()

	mockRepo.On("Delete", id).Return(nil)

	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	err := service.DeleteUser(id)

	// Assert
	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_RepoError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	id := uuid.New()

	mockRepo.On("Delete", id).Return(errors.New("db error"))
	mockLogger.On("Error", "Error deleting user with ID: %s, %v", mock.Anything, mock.Anything).Return()

	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	err := service.DeleteUser(id)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "error deleting user", err.Error())
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestResetPassword_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockEmailService := new(MockEmailService)
	email := "test@example.com"

	user := &models.User{
		Email: email,
	}

	mockRepo.On("FindByEmail", email).Return(user, nil)
	mockRepo.On("Update", user).Return(user, nil)
	mockEmailService.On("SendTemplatedEmail", email, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	service := NewUserService(mockRepo, mockEmailService, nil, nil)

	// Act
	err := service.ResetPassword(email)

	// Assert
	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestResetPassword_RepoFindError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	email := "test@example.com"

	mockRepo.On("FindByEmail", email).Return(nil, errors.New("not found"))

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	err := service.ResetPassword(email)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestResetPassword_RepoUpdateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockEmailService := new(MockEmailService)
	email := "test@example.com"

	user := &models.User{
		Email: email,
	}

	mockRepo.On("FindByEmail", email).Return(user, nil)
	mockRepo.On("Update", user).Return(user, errors.New("update error"))

	service := NewUserService(mockRepo, mockEmailService, nil, nil)

	// Act
	err := service.ResetPassword(email)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "update error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestResetPassword_EmailServiceError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockEmailService := new(MockEmailService)
	email := "test@example.com"

	user := &models.User{
		Email: email,
	}

	mockRepo.On("FindByEmail", email).Return(user, nil)
	mockRepo.On("Update", user).Return(user, nil)
	mockEmailService.On("SendTemplatedEmail", email, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("email error"))

	service := NewUserService(mockRepo, mockEmailService, nil, nil)

	// Act
	err := service.ResetPassword(email)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "email error", err.Error())
	mockRepo.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestVerifyPasswordResetToken_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	token := "sample-token"
	mockEmailService := new(MockEmailService)

	expiry := time.Now().Add(1 * time.Hour)
	user := &models.User{
		ResetTokenExpires: &expiry,
	}

	mockRepo.On("FindByResetToken", token).Return(user, nil)

	service := NewUserService(mockRepo, mockEmailService, nil, nil)

	// Act
	err := service.VerifyPasswordResetToken(token)

	// Assert
	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
}

func TestVerifyPasswordResetToken_UserNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	token := "sample-token"

	mockRepo.On("FindByResetToken", token).Return(nil, nil)

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	err := service.VerifyPasswordResetToken(token)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "invalid token", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestVerifyPasswordResetToken_TokenExpired(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	token := "sample-token"

	expiry := time.Now().Add(-1 * time.Hour)
	user := &models.User{
		ResetTokenExpires: &expiry,
	}

	mockRepo.On("FindByResetToken", token).Return(user, nil)

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	err := service.VerifyPasswordResetToken(token)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "token expired", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestChangePassword_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockHasher := new(MockPasswordHasher)
	userID := uuid.New()
	newPassword := "new-password"

	user := &models.User{
		Password: "old-password",
	}

	hashedPassword := "hashed-password"

	mockRepo.On("FindByID", userID).Return(user, nil)
	mockHasher.On("Hash", newPassword).Return(hashedPassword, nil)
	mockRepo.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return u.Password == hashedPassword
	})).Return(user, nil)

	service := NewUserService(mockRepo, nil, nil, mockHasher)

	// Act
	err := service.UpdatePassword(userID, newPassword)

	// Assert
	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

func TestChangePassword_FindUserError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userID := uuid.New()

	mockRepo.On("FindByID", userID).Return(nil, errors.New("user not found"))

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	err := service.UpdatePassword(userID, "new-password")

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestChangePassword_HashError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockHasher := new(MockPasswordHasher)
	mockLogger := new(MockLogger)

	userID := uuid.New()

	user := &models.User{
		Password: "old-password",
	}

	mockRepo.On("FindByID", userID).Return(user, nil)
	mockHasher.On("Hash", mock.Anything).Return("", errors.New("hashing error"))
	mockLogger.On("Error", "Error generating hashed password for user with ID: %s, %v", mock.Anything, mock.Anything).Return()

	service := NewUserService(mockRepo, nil, mockLogger, mockHasher)

	// Act
	err := service.UpdatePassword(userID, "new-password")

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "failed to change password due to internal error", err.Error())
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

func TestChangePassword_UpdateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockHasher := new(MockPasswordHasher)
	mockLogger := new(MockLogger)
	userID := uuid.New()
	newPassword := "new-password"

	user := &models.User{
		Password: "old-password",
	}

	hashedPassword := "hashed-password"

	mockRepo.On("FindByID", userID).Return(user, nil)
	mockHasher.On("Hash", newPassword).Return(hashedPassword, nil)
	mockRepo.On("Update", mock.Anything).Return((*models.User)(nil), errors.New("update error"))

	mockLogger.On("Error", "Error : %s, %v", mock.Anything, mock.Anything).Return()

	service := NewUserService(mockRepo, nil, mockLogger, mockHasher)

	// Act
	err := service.UpdatePassword(userID, newPassword)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "update error", err.Error())
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

func TestBlockUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userID := uuid.New()

	user := &models.User{
		IsBlocked: false,
	}

	mockRepo.On("FindByID", userID).Return(user, nil)
	mockRepo.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return u.IsBlocked
	})).Return(user, nil)

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	err := service.BlockUser(userID)

	// Assert
	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBlockUser_FindUserError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	userID := uuid.New()

	mockRepo.On("FindByID", userID).Return(nil, errors.New("user not found"))
	mockLogger.On("Error", "Error finding user with ID: %s, %v", mock.Anything, mock.Anything).Return()

	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	err := service.BlockUser(userID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestBlockUser_UpdateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	userID := uuid.New()

	user := &models.User{
		IsBlocked: false,
	}

	updatedUser := &models.User{
		IsBlocked: true,
	}
	mockRepo.On("FindByID", userID).Return(user, nil)
	mockRepo.On("Update", mock.Anything).Return(updatedUser, errors.New("update error"))
	mockLogger.On("Error", "Error blocking user with ID: %s, %v", mock.Anything, mock.Anything).Return()

	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	err := service.BlockUser(userID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "update error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUnblockUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	userID := uuid.New()

	user := &models.User{
		IsBlocked: true,
	}

	mockRepo.On("FindByID", userID).Return(user, nil)
	mockRepo.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return !u.IsBlocked
	})).Return(user, nil)
	mockLogger.On("Error", "Error unblocking user with ID: %s, %v", mock.Anything, mock.Anything).Return()
	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	err := service.UnblockUser(userID)

	// Assert
	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUnblockUser_FindUserError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	userID := uuid.New()

	mockRepo.On("FindByID", userID).Return(nil, errors.New("user not found"))
	mockLogger.On("Error", "Error finding user with ID: %s, %v", mock.Anything, mock.Anything).Return()

	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	err := service.UnblockUser(userID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUnblockUser_UpdateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	userID := uuid.New()

	user := &models.User{
		IsBlocked: true,
	}

	mockRepo.On("FindByID", userID).Return(user, nil)
	dummyUser := &models.User{}
	mockRepo.On("Update", mock.Anything).Return(dummyUser, errors.New("update error"))
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	err := service.UnblockUser(userID)

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "update error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestGetUserByEmail_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	email := "test@example.com"

	user := &models.User{
		Email: email,
	}

	mockRepo.On("FindByEmail", email).Return(user, nil)

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	result, err := service.GetUserByEmail(email)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, email, result.Email)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByEmail_RepoError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockLogger := new(MockLogger)
	email := "test@example.com"

	mockRepo.On("FindByEmail", email).Return(nil, errors.New("database error"))
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	service := NewUserService(mockRepo, nil, mockLogger, nil)

	// Act
	result, err := service.GetUserByEmail(email)

	// Assert
	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Equal(t, "failed to fetch user", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestGetUserByEmail_UserNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	email := "test@example.com"

	mockRepo.On("FindByEmail", email).Return(nil, nil)

	service := NewUserService(mockRepo, nil, nil, nil)

	// Act
	result, err := service.GetUserByEmail(email)

	// Assert
	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertExpectations(t)
}
