package authentication

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"idp-automations-hub/internal/app/config"
	"idp-automations-hub/internal/app/dto"
	"idp-automations-hub/internal/app/models"
	"idp-automations-hub/internal/app/repositories"
	"idp-automations-hub/internal/app/services"
	"idp-automations-hub/internal/app/services/iservice"
	"idp-automations-hub/internal/app/utils"
	"idp-automations-hub/internal/infra"
	"math"
	"time"
)

type service struct {
	userService      iservice.UserService
	hasher           utils.PasswordHasher
	blockListService iservice.TokenBlockListService
	logger           iservice.Logger
	sender           iservice.MessageSender
	jwtSecret        string
}

func NewService(userService iservice.UserService, hasher utils.PasswordHasher, sender iservice.MessageSender,
	blockListService iservice.TokenBlockListService, logger iservice.Logger, jwtSecret string) IService {
	return &service{
		userService:      userService,
		hasher:           hasher,
		blockListService: blockListService,
		logger:           logger,
		sender:           sender,
		jwtSecret:        jwtSecret,
	}
}

func GetDefaultAuthService() (IService, error) {
	logger, err := services.NewKafkaLogger(config.KafkaConfig.BrokersAddr, config.KafkaConfig.LoggerTopic)
	if err != nil {
		return nil, err
	}
	database, err := infra.GetDefaultDB()
	if err != nil {
		return nil, err
	}
	userRepository := repositories.NewGormUserRepository(database, logger)
	userService := services.NewUserService(userRepository, logger)
	hasher := utils.DefaultBcryptHasher()
	sender, err := services.NewKafkaMessageSender()
	if err != nil {
		return nil, err
	}
	blockListService := services.NewRedisTokenBlockListService()
	return NewService(userService, hasher, sender, blockListService, logger, config.AuthenticationConfig.JwtSecret), nil
}

func (a *service) Register(userDTO dto.UserDTO) (*dto.UserResponse, error) {
	hashedPassword, err := a.hasher.Hash(userDTO.Password)
	if err != nil {
		a.logger.Error("Error generating hashed password for user with email: %s, %v", userDTO.Email, err)
		return nil, errors.New("failed to register user due to internal error")
	}

	user := models.User{
		Email:    userDTO.Email,
		Password: hashedPassword,
	}

	userCreated, err := a.userService.CreateUser(user)
	if err != nil {
		a.logger.Error("Error creating user: %v", err)
		return nil, errors.New("failed to create user")
	}

	a.logger.Info("Successfully registered user: %s", user.Email)
	msg := struct {
		Email string
	}{
		Email: user.Email,
	}
	err = a.sender.Send(config.AuthenticationConfig.AccountCreatedTopic, msg)
	if err != nil {
		a.logger.Error("Error sending account created message: %v", err)
	}

	return &dto.UserResponse{
		ID:    userCreated.ID,
		Email: userCreated.Email,
	}, nil
}

func (a *service) Login(email, password string) (*dto.TokenDetails, error) {
	user, err := a.userService.GetUserByEmail(email)
	if err != nil {
		a.logger.Error("Error fetching user by email: %v", err)
		return nil, errors.New("invalid credentials")
	}

	// Check if account is blocked and if the block time hasn't expired
	now := time.Now()
	if user.IsBlocked && user.BlockedUntil != nil && now.Before(*user.BlockedUntil) {
		a.logger.Warn("Login attempt for blocked user: %s", email)
		return nil, errors.New("account is blocked")
	}

	// Check for rapid subsequent login attempts
	if user.LastAttempt != nil && now.Sub(*user.LastAttempt) < config.AuthenticationConfig.MinTimeBetweenAttemptsSeconds*time.Second {
		a.logger.Warn("Rapid subsequent login attempt detected for user: %s", email)
		return nil, errors.New("please wait a moment before trying again")
	}

	// If the account was blocked but the block time has expired, unblock the account
	if user.IsBlocked && (user.BlockedUntil == nil || now.After(*user.BlockedUntil)) {
		user.IsBlocked = false
		user.FailedAttempts = 0
		user.BlockedUntil = nil
		user, err = a.userService.UpdateUser(*user)
		if err != nil {
			return nil, errors.New("failed to unblock account")
		}
	}

	hashErr := a.hasher.Compare(user.Password, password)
	if hashErr != nil {
		user.FailedAttempts++
		user.LastAttempt = &now
		if user.FailedAttempts >= config.AuthenticationConfig.MaxLoginAttemptsBeforeBlock {
			blockDuration := calculateBlockDuration(user.FailedAttempts)
			blockedUntil := now.Add(blockDuration)
			user.BlockedUntil = &blockedUntil
			user.IsBlocked = true
			a.logger.Warn("User %s is blocked until %s", email, blockedUntil.String())
			msg := struct {
				Email        string
				BlockedUntil time.Time
			}{
				Email:        email,
				BlockedUntil: blockedUntil,
			}
			err = a.sender.Send(config.AuthenticationConfig.AccountBlockedTopic, msg)
		}
		_, updateErr := a.userService.UpdateUser(*user)
		if updateErr != nil {
			a.logger.Error("Failed to update user after failed login: %v", updateErr)
		}
		a.logger.Warn("Hash comparison failed for user %s: %v", email, hashErr)
		return nil, errors.New("invalid credentials")
	}

	// Reset FailedAttempts since login is successful
	user.FailedAttempts = 0
	user.LastAttempt = &now
	_, updateErr := a.userService.UpdateUser(*user)
	if updateErr != nil {
		a.logger.Error("Failed to reset failed attempts for user %s: %v", email, updateErr)
	}

	td := &dto.TokenDetails{}
	td.RefreshToken, td.RefreshUUID, td.RtExpires, err = a.generateRefreshToken(user.ID)
	if err != nil {
		a.logger.Error("Failed to generate refresh token for user %s: %v", email, err)
		return nil, errors.New("failed to generate refresh token")
	}
	td.AccessToken, td.AtExpires, err = a.generateAccessToken(user.ID, td.RefreshUUID, td.RtExpires)
	if err != nil {
		a.logger.Error("Failed to generate access token for user %s: %v", email, err)
		return nil, errors.New("failed to generate access token")
	}

	a.logger.Info("Successfully logged in user: %s", email)

	return td, nil
}

func calculateBlockDuration(failedLoginAttempts int) time.Duration {
	exponent := float64(failedLoginAttempts - config.AuthenticationConfig.MaxLoginAttemptsBeforeBlock)
	initialBlockDuration := time.Duration(config.AuthenticationConfig.BaseBlockDurationMinutes) * time.Minute
	return initialBlockDuration * time.Duration(math.Pow(2, exponent))
}

func (a *service) Logout(accessToken string) error {
	_, claims, err := a.parseAndValidateToken(accessToken)

	userID, ok := claims["user_id"].(string)
	if !ok {
		a.logger.Error("Error parsing user ID from claims: %v", err)
		return err
	}

	accessUUID, ok := claims["access_uuid"].(string)
	if !ok {
		a.logger.Warn("Access UUID not found in the token for user: %s", userID)
		return errors.New("access UUID not found in the token")
	}

	refreshUUID, ok := claims["refresh_uuid"].(string)
	if !ok {
		a.logger.Warn("Refresh UUID not found in the token for user: %s", userID)
		return errors.New("refresh UUID not found in the token")
	}

	// Calculates the expiration time of the tokens to define the time they remain on the block list.
	refreshExp, ok := claims["refresh_exp"].(int64)
	if !ok {
		a.logger.Warn("Refresh expiration time not found in the token for user: %s", userID)
		return errors.New("refresh expiration time not found in the token")
	}
	rtDuration := time.Until(time.Unix(refreshExp, 0))

	atExpires, ok := claims["exp"].(int64)
	if !ok {
		a.logger.Warn("Expiration time not found in the token for user: %s", userID)
		return errors.New("expiration time not found in the token")
	}
	atDuration := time.Until(time.Unix(atExpires, 0))

	// Add the access token and refresh token UUIDs to the block list
	err = a.blockListService.AddToBlockList(accessUUID, atDuration)
	if err != nil {
		a.logger.Error("Failed to add access token to block list for user: %s, Error: %v", userID, err)
		return err
	}
	err = a.blockListService.AddToBlockList(refreshUUID, rtDuration)
	if err != nil {
		a.logger.Error("Failed to add refresh token to block list for user: %s, Error: %v", userID, err)
		return err
	}

	a.logger.Info("Successfully logged out and blocked tokens for user: %s with accessUUID: %s and refreshUUID: %s", userID, accessUUID, refreshUUID)
	return nil
}

func (a *service) RefreshToken(refreshToken string) (*dto.TokenDetails, error) {
	_, claims, err := a.parseAndValidateToken(refreshToken)

	refreshUUID, ok := claims["refresh_uuid"].(string)
	if !ok {
		a.logger.Warn("Refresh UUID not found in the token")
		return nil, errors.New("refresh UUID not found in the token")
	}

	// Check if the refresh token is on the block list
	isBlocked, err := a.blockListService.IsInBlockList(refreshUUID)
	if err != nil {
		a.logger.Error("Failed to check blockList status: %v", err)
		return nil, errors.New("error checking blockList status")
	}
	if isBlocked {
		a.logger.Warn("Refresh token is blocked")
		return nil, errors.New("refresh token is blocked")
	}

	// Renew the access token using the refresh token's claims
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		a.logger.Error("Error parsing user ID from claims: %v", err)
		return nil, err
	}

	refreshExp, ok := claims["refresh_exp"].(int64)
	if !ok {
		a.logger.Warn("Refresh expiration time not found in the token for user: %s", userID)
		return nil, errors.New("refresh expiration time not found in the token")
	}
	newAccessToken, atExpires, err := a.generateAccessToken(userID, refreshUUID, refreshExp)
	if err != nil {
		a.logger.Error("Failed to generate new access token: %v", err)
		return nil, err
	}

	td := &dto.TokenDetails{
		AccessToken:  newAccessToken,
		AtExpires:    atExpires,
		RefreshToken: refreshToken,
		RefreshUUID:  refreshUUID,
		RtExpires:    refreshExp,
	}

	a.logger.Info("Successfully renewed access token for user: %s", userID.String())

	return td, nil
}

func (a *service) IsUserAuthenticated(accessToken string) (bool, error) {
	_, claims, err := a.parseAndValidateToken(accessToken)
	if err != nil {
		a.logger.Error("Error parsing accessToken: %v", err)
		return false, err
	}
	// Check if the accessToken is in the blockList
	accessUUID, ok := claims["access_uuid"].(string)
	if !ok {
		a.logger.Warn("Access UUID not found in the accessToken")
		return false, errors.New("invalid accessToken")
	}

	isBlocked, err := a.blockListService.IsInBlockList(accessUUID)
	if err != nil {
		a.logger.Error("Error checking accessToken in block list: %v", err)
		return false, err
	}

	if isBlocked {
		a.logger.Warn("Token is blocked")
		return false, errors.New("accessToken is blocked")
	}

	return true, nil
}

func (a *service) RequestPasswordReset(email string) (string, time.Time, error) {
	user, err := a.userService.GetUserByEmail(email)
	if err != nil {
		a.logger.Error("Error fetching user by email: %v", err)
		return "", time.Time{}, errors.New("invalid email")
	}

	// Generate a reset token
	resetToken := uuid.New().String()
	resetTokenExpires := time.Now().Add(time.Hour * config.AuthenticationConfig.ExpirationTimeResetTokenHours)

	// Add the reset token to the user
	user.ResetPasswordToken = resetToken
	user.ResetTokenExpires = &resetTokenExpires

	_, err = a.userService.UpdateUser(*user)
	if err != nil {
		a.logger.Error("Error updating user: %v", err)
		return "", time.Time{}, errors.New("failed to update user")
	}

	// Send the message with the reset token
	msg := struct {
		Email          string
		ResetToken     string
		TokenExpiresIn int64
	}{
		Email:          email,
		ResetToken:     resetToken,
		TokenExpiresIn: resetTokenExpires.Unix(),
	}
	err = a.sender.Send(config.AuthenticationConfig.PasswordResetTopic, msg)
	if err != nil {
		a.logger.Error("Error sending reset token message: %v", err)
		return "", time.Time{}, errors.New("failed to send reset token")
	}

	a.logger.Info("Successfully sent reset token to user: %s", email)
	return resetToken, resetTokenExpires, nil
}

func (a *service) ConfirmPasswordReset(token, newPassword string) error {
	user, err := a.userService.GetUserByResetToken(token)
	if err != nil {
		a.logger.Error("Error fetching user by reset token: %v", err)
		return errors.New("invalid token")
	}

	if user == nil {
		return errors.New("invalid token")
	}

	if user.ResetTokenExpires.Before(time.Now()) {
		return errors.New("token expired")
	}

	hashedPassword, err := a.hasher.Hash(newPassword)
	if err != nil {
		a.logger.Error("Error generating hashed password for user with ID: %s, %v", user.ID, err)
		return errors.New("failed to change password due to internal error")
	}

	user.Password = hashedPassword
	user.ResetPasswordToken = ""
	user.ResetTokenExpires = nil

	err = a.userService.UpdatePassword(user.ID, newPassword)
	if err != nil {
		a.logger.Error("Error updating user: %v", err)
		return errors.New("failed to change password")
	}

	_, err = a.userService.UpdateUser(*user)
	if err != nil {
		a.logger.Error("Error updating user: %v", err)
		return errors.New("failed to update user")
	}

	return nil
}

func (a *service) ChangePassword(email string, newPassword string) error {
	user, err := a.userService.GetUserByEmail(email)
	if err != nil {
		a.logger.Error("Error fetching user by email: %v", err)
		return errors.New("invalid email")
	}

	hashedPassword, hashErr := a.hasher.Hash(newPassword)
	if hashErr != nil {
		a.logger.Error("Error hashing new password: %v", hashErr)
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	user.ResetPasswordToken = ""
	user.ResetTokenExpires = nil

	updateErr := a.userService.UpdatePassword(user.ID, newPassword)
	if updateErr != nil {
		a.logger.Error("Error updating user password: %v", updateErr)
		return errors.New("failed to update password")
	}
	_, err = a.userService.UpdateUser(*user)
	if err != nil {
		a.logger.Error("Error updating user: %v", err)
		return errors.New("failed to update user")
	}

	a.logger.Info("Successfully changed password for user: %s", email)
	return nil
}

func (a *service) generateAccessToken(userID uuid.UUID, refreshUUID string, refreshExp int64) (string, int64, error) {
	expires := time.Now().Add(time.Minute * config.AuthenticationConfig.AccessTokenDurationMinutes).Unix()

	claims := jwt.MapClaims{}
	claims["user_id"] = userID.String()
	claims["access_uuid"] = uuid.New().String()
	claims["refresh_uuid"] = refreshUUID
	claims["refresh_exp"] = refreshExp
	claims["exp"] = expires

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(a.jwtSecret))
	return accessToken, expires, err
}

func (a *service) generateRefreshToken(userID uuid.UUID) (string, string, int64, error) {
	refreshUUID := uuid.New().String()
	expires := time.Now().Add(config.AuthenticationConfig.RefreshTokenDurationDays).Unix()

	claims := jwt.MapClaims{}
	claims["refresh_uuid"] = refreshUUID
	claims["user_id"] = userID.String()
	claims["exp"] = expires

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString([]byte(a.jwtSecret))
	return refreshToken, refreshUUID, expires, err
}

func (a *service) parseAndValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			a.logger.Error("Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, nil, errors.New("invalid token")
	}

	return token, claims, nil
}
