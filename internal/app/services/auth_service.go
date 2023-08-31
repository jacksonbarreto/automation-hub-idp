package services

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"idp-automations-hub/internal/app/config"
	"idp-automations-hub/internal/app/dto"
	"idp-automations-hub/internal/app/services/iservice"
	"idp-automations-hub/internal/app/utils"
	"os"
	"strconv"
	"time"
)

type authService struct {
	userService          iservice.UserService
	hasher               utils.PasswordHasher
	blockListService     iservice.TokenBlockListService
	logger               Logger
	jwtSecret            string
	RefreshTokenDuration int
	AccessTokenDuration  int
}

func NewAuthService(userService iservice.UserService, hasher utils.PasswordHasher,
	blockListService iservice.TokenBlockListService, logger Logger, jwtSecret string) iservice.AuthService {
	return &authService{
		userService:          userService,
		hasher:               hasher,
		blockListService:     blockListService,
		logger:               logger,
		jwtSecret:            jwtSecret,
		RefreshTokenDuration: getEnvExpire(config.RefreshTokenDurationDays, 7),
		AccessTokenDuration:  getEnvExpire(config.AccessTokenDurationMinutes, 15),
	}
}

func (a *authService) Login(email, password string) (*dto.TokenDetails, error) {
	user, err := a.userService.GetUserByEmail(email)
	if err != nil {
		a.logger.Error("Error fetching user by email: %v", err)
		return nil, errors.New("invalid credentials")
	}

	hashErr := a.hasher.Compare(user.Password, password)
	if hashErr != nil {
		a.logger.Warn("Hash comparison failed for user %s: %v", email, hashErr)
		return nil, errors.New("invalid credentials")
	}

	if user.IsBlocked {
		a.logger.Warn("Login attempt for blocked user: %s", email)
		return nil, errors.New("account is blocked")
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

func (a *authService) Logout(accessToken string) error {
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
	atDuration := time.Until(time.Unix(int64(atExpires), 0))

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

func (a *authService) RefreshToken(refreshToken string) (*dto.TokenDetails, error) {
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

func (a *authService) IsUserAuthenticated(accessToken string) (bool, error) {
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

func (a *authService) generateAccessToken(userID uuid.UUID, refreshUUID string, refreshExp int64) (string, int64, error) {
	expires := time.Now().Add(time.Minute * time.Duration(a.AccessTokenDuration)).Unix()

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

func (a *authService) generateRefreshToken(userID uuid.UUID) (string, string, int64, error) {
	refreshUUID := uuid.New().String()
	expires := time.Now().Add(time.Hour * 24 * time.Duration(a.RefreshTokenDuration)).Unix()

	claims := jwt.MapClaims{}
	claims["refresh_uuid"] = refreshUUID
	claims["user_id"] = userID.String()
	claims["exp"] = expires

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString([]byte(a.jwtSecret))
	return refreshToken, refreshUUID, expires, err
}

func getEnvExpire(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		intVal, err := strconv.Atoi(value)
		if err == nil {
			return intVal
		}
	}
	return defaultVal
}

func (a *authService) parseAndValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
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