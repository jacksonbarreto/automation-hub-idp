package config

import (
	"errors"
	"fmt"
	"time"
)

const (
	baseBlockDurationMinutes        string = "BLOCKING_TIME_EXPONENTIATION_BASIS"
	maxLoginAttemptsBeforeBlock     string = "MAX_LOGIN_ATTEMPTS_BEFORE_BLOCK"
	minTimeBetweenAttemptsInSeconds string = "MIN_TIME_BETWEEN_ATTEMPTS_IN_SECONDS"
	expirationTimeResetTokenInHours string = "EXPIRATION_TIME_RESET_TOKEN_IN_HOURS"
	accessTokenDurationMinutes      string = "ACCESS_TOKEN_DURATION_MINUTES"
	refreshTokenDurationDays        string = "REFRESH_TOKEN_DURATION_DAYS"
	passwordResetTopic              string = "PASSWORD_RESET_TOPIC"
	accountBlockedTopic             string = "ACCOUNT_BLOCKED_TOPIC"
	accountCreatedTopic             string = "ACCOUNT_CREATED_TOPIC"
	jwtSecret                              = "JWT_SECRET"
)

type authenticationConfig struct {
	BaseBlockDurationMinutes      int
	MaxLoginAttemptsBeforeBlock   int
	MinTimeBetweenAttemptsSeconds time.Duration
	ExpirationTimeResetTokenHours time.Duration
	AccessTokenDurationMinutes    time.Duration
	RefreshTokenDurationDays      time.Duration
	PasswordResetTopic            string
	AccountBlockedTopic           string
	AccountCreatedTopic           string
	JwtSecret                     string
}

func newAuthenticationConfig() (*authenticationConfig, error) {
	passwordResetTopicValue := getEnvString(passwordResetTopic, "NULL")
	accountBlockedTopicValue := getEnvString(accountBlockedTopic, "NULL")
	accountCreatedTopicValue := getEnvString(accountCreatedTopic, "NULL")
	if passwordResetTopicValue == "NULL" || accountBlockedTopicValue == "NULL" || accountCreatedTopicValue == "NULL" {
		errorMessage := fmt.Sprintf("error: One or more topics are not set, please check the environment variables: %s, %s, %s", passwordResetTopic, accountBlockedTopic, accountCreatedTopic)
		return nil, errors.New(errorMessage)
	}
	baseBlockDurationMinutesValue := getEnvInt(baseBlockDurationMinutes, 0)
	if baseBlockDurationMinutesValue == 0 {
		errorMessage := fmt.Sprintf("error: Base block duration minutes should be greater than 0, please check the environment variable: %s", baseBlockDurationMinutes)
		return nil, errors.New(errorMessage)
	}
	maxLoginAttemptsBeforeBlockValue := getEnvInt(maxLoginAttemptsBeforeBlock, 0)
	if maxLoginAttemptsBeforeBlockValue == 0 {
		errorMessage := fmt.Sprintf("error: Max login attempts before block should be greater than 0, please check the environment variable: %s", maxLoginAttemptsBeforeBlock)
		return nil, errors.New(errorMessage)
	}

	jwtSecret := getEnvString(jwtSecret, "NULL")
	if jwtSecret == "NULL" {
		errorMessage := fmt.Sprintf("error: JWT secret is not set, please check the environment variable: %s", jwtSecret)
		return nil, errors.New(errorMessage)
	}

	return &authenticationConfig{
		BaseBlockDurationMinutes:      baseBlockDurationMinutesValue,
		MaxLoginAttemptsBeforeBlock:   maxLoginAttemptsBeforeBlockValue,
		MinTimeBetweenAttemptsSeconds: time.Duration(getEnvInt(minTimeBetweenAttemptsInSeconds, 0)),
		ExpirationTimeResetTokenHours: time.Duration(getEnvInt(expirationTimeResetTokenInHours, 24)),
		AccessTokenDurationMinutes:    time.Duration(getEnvInt(accessTokenDurationMinutes, 15)),
		RefreshTokenDurationDays:      time.Duration(24*getEnvInt(refreshTokenDurationDays, 4)) * time.Hour,
		PasswordResetTopic:            passwordResetTopicValue,
		AccountBlockedTopic:           accountBlockedTopicValue,
		AccountCreatedTopic:           accountCreatedTopicValue,
		JwtSecret:                     jwtSecret,
	}, nil
}
