package services

import (
	"errors"
	"github.com/benbjohnson/clock"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"idp-automations-hub/internal/app/models"
	"idp-automations-hub/internal/app/services/service_mock"
	"idp-automations-hub/internal/app/utils/utils_mock"
	"os"
	"testing"
	"time"
)

func TestLogin(t *testing.T) {
	user := models.SimulateUser()
	mockUserSvc := new(service_mock.MockUserService)
	mockHasher := new(utils_mock.MockHasher)

	a := &authService{
		userService: mockUserSvc,
		hasher:      mockHasher,
		jwtSecret:   "test-secret",
	}

	t.Run("Successful Login Test", func(t *testing.T) {
		mockUserSvc.On("GetUserByEmail", user.Email).Return(&user, nil)
		mockHasher.On("Compare", user.Password, mock.Anything).Return(nil)

		td, err := a.Login(user.Email, "correctPassword")
		assert.NoError(t, err)
		assert.NotNil(t, td)
		assert.NotEmpty(t, td.AccessToken)
		assert.NotEmpty(t, td.RefreshToken)
	})

	t.Run("Invalid Credentials Test", func(t *testing.T) {
		mockUserSvc.On("GetUserByEmail", user.Email).Return(&user, nil)
		mockHasher.On("Compare", user.Password, mock.Anything).Return(errors.New("invalid password"))

		td, err := a.Login(user.Email, "wrongPassword")
		assert.Error(t, err)
		assert.Nil(t, td)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("Blocked Account Test", func(t *testing.T) {
		user.IsBlocked = true
		mockUserSvc.On("GetUserByEmail", user.Email).Return(&user, nil)

		td, err := a.Login(user.Email, "correctPassword")
		assert.Error(t, err)
		assert.Nil(t, td)
		assert.Equal(t, "account is blocked", err.Error())
	})
}

func TestRefreshToken(t *testing.T) {
	mockBlockListSvc := new(service_mock.MockBlockListService)
	mockUserSvc := new(service_mock.MockUserService)
	mockHasher := new(utils_mock.MockHasher)

	a := &authService{
		userService:      mockUserSvc,
		hasher:           mockHasher,
		jwtSecret:        "test-secret",
		blockListService: mockBlockListSvc,
	}

	userID := uuid.New()
	refreshUUID := uuid.New().String()
	expires := time.Now().Add(time.Hour * 24 * 7).Unix()

	// Generate a valid refresh token
	claims := jwt.MapClaims{}
	claims["refresh_uuid"] = refreshUUID
	claims["user_id"] = userID.String()
	claims["exp"] = expires
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	validRefreshToken, _ := token.SignedString([]byte(a.jwtSecret))

	// Generate a refresh token with a different signing method (RS256)
	token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	invalidSignatureToken, _ := token.SignedString([]byte(a.jwtSecret))

	t.Run("Valid Refresh Token Test", func(t *testing.T) {
		mockBlockListSvc.On("IsInBlockList", refreshUUID).Return(false, nil)

		td, err := a.RefreshToken(validRefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, td)
		assert.NotEmpty(t, td.AccessToken)
		assert.Equal(t, validRefreshToken, td.RefreshToken)
	})

	t.Run("Invalid Signature Method Test", func(t *testing.T) {
		td, err := a.RefreshToken(invalidSignatureToken)
		assert.Error(t, err)
		assert.Nil(t, td)
		assert.Contains(t, err.Error(), "unexpected signing method")
	})

	t.Run("Invalid Refresh Token Test", func(t *testing.T) {
		invalidToken := validRefreshToken + "invalid"
		td, err := a.RefreshToken(invalidToken)
		assert.Error(t, err)
		assert.Nil(t, td)
		assert.Equal(t, "invalid refresh token", err.Error())
	})

	t.Run("Missing Refresh UUID Test", func(t *testing.T) {
		claims["refresh_uuid"] = nil
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenWithoutUUID, _ := token.SignedString([]byte(a.jwtSecret))

		td, err := a.RefreshToken(tokenWithoutUUID)
		assert.Error(t, err)
		assert.Nil(t, td)
		assert.Equal(t, "refresh UUID not found in the token", err.Error())
	})

	t.Run("Blocked Refresh Token Test", func(t *testing.T) {
		mockBlockListSvc.On("IsInBlockList", refreshUUID).Return(true, nil)

		td, err := a.RefreshToken(validRefreshToken)
		assert.Error(t, err)
		assert.Nil(t, td)
		assert.Equal(t, "refresh token is blocked", err.Error())
	})
}

func TestGetEnvExpire(t *testing.T) {
	testKey := "TEST_EXPIRE"

	// 1. Environment variable exists and has a valid integer value
	err := os.Setenv(testKey, "123")
	require.NoError(t, err)
	assert.Equal(t, 123, getEnvExpire(testKey, 456))

	// 2. Environment variable exists but has a non-integer value
	err = os.Setenv(testKey, "abc")
	require.NoError(t, err)
	assert.Equal(t, 456, getEnvExpire(testKey, 456))

	// 3. Environment variable does not exist
	err = os.Unsetenv(testKey)
	require.NoError(t, err)
	assert.Equal(t, 456, getEnvExpire(testKey, 456))
}

func TestGenerateToken(t *testing.T) {
	// Setup
	a := &authService{
		jwtSecret: "super-secret",
	}
	userID := uuid.New()
	var mockClock = clock.NewMock()

	t.Run("Basic Token Generation Test", func(t *testing.T) {
		td, err := a.generateToken(userID)

		assert.NoError(t, err)
		assert.NotEmpty(t, td.AccessToken)
		assert.NotEmpty(t, td.RefreshToken)

		_, err = jwt.Parse(td.AccessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.jwtSecret), nil
		})
		assert.NoError(t, err)

		_, err = jwt.Parse(td.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.jwtSecret), nil
		})
		assert.NoError(t, err)
	})

	t.Run("Token Expiration Test", func(t *testing.T) {
		td, err := a.generateToken(userID)
		assert.NoError(t, err)

		// Mock time to simulate token expiration
		mockClock.Add(16 * time.Minute)
		token, err := jwt.Parse(td.AccessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.jwtSecret), nil
		})
		assert.Error(t, err) // Should error because token is expired
		assert.NotNil(t, token.Claims.(jwt.MapClaims)["exp"])
	})

	t.Run("Token Claims Test", func(t *testing.T) {
		td, err := a.generateToken(userID)
		assert.NoError(t, err)

		token, err := jwt.Parse(td.AccessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.jwtSecret), nil
		})
		assert.NoError(t, err)
		claims := token.Claims.(jwt.MapClaims)

		assert.Equal(t, userID.String(), claims["user_id"])
		assert.Equal(t, td.AccessUUID, claims["access_uuid"])
	})

	t.Run("Token Signature Failure Test", func(t *testing.T) {
		a.jwtSecret = "wrong-secret"
		_, err := a.generateToken(userID)
		assert.Error(t, err) // It should fail because the secret is wrong
	})
}
