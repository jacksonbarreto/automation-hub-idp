package utils

import (
	"os"
	"strconv"
	"time"
)

type PasswordResetOptions struct {
	Domain       string
	Endpoint     string
	EmailSubject string
	TokenExpiry  time.Duration
}

const (
	DefaultAppDomain     = "https://localhost:3000" // TODO: Change this to the actual domain
	DefaultResetEndpoint = "/reset-password"
	DefaultEmailSubject  = "Password Reset"
	DefaultTokenExpiry   = 1 * time.Hour
)

func NewPasswordResetOptions(domain, endpoint, emailSubject string, tokenExpiry time.Duration) PasswordResetOptions {
	return PasswordResetOptions{
		Domain:       domain,
		Endpoint:     endpoint,
		EmailSubject: emailSubject,
		TokenExpiry:  tokenExpiry,
	}
}

func DefaultPasswordResetOptions() PasswordResetOptions {
	domain := os.Getenv("APP_DOMAIN")
	if domain == "" {
		domain = DefaultAppDomain
	}

	endpoint := os.Getenv("RESET_ENDPOINT")
	if endpoint == "" {
		endpoint = DefaultResetEndpoint
	}

	emailSubject := os.Getenv("RESET_EMAIL_SUBJECT")
	if emailSubject == "" {
		emailSubject = DefaultEmailSubject
	}

	tokenExpiresStr := os.Getenv("EXPIRATION_TIME_RESET_TOKEN_IN_HOURS")
	var tokenExpiry time.Duration = DefaultTokenExpiry
	if tokenExpiresStr != "" {
		tokenExpiresInt, err := strconv.Atoi(tokenExpiresStr)
		if err == nil && tokenExpiresInt > 0 {
			tokenExpiry = time.Duration(tokenExpiresInt) * time.Hour
		}
	}

	return PasswordResetOptions{
		Domain:       domain,
		Endpoint:     endpoint,
		EmailSubject: emailSubject,
		TokenExpiry:  tokenExpiry,
	}
}
