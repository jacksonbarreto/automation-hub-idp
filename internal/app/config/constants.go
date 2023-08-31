package config

const (
	LOGIN_ATTEMPT_BASE                   string = "LOGIN_ATTEMPT_BASE"
	MAX_LOGIN_ATTEMPTS_BEFORE_BLOCK      string = "MAX_LOGIN_ATTEMPTS_BEFORE_BLOCK"
	DOMAIN                               string = "DOMAIN"
	RESET_ENDPOINT                       string = "RESET_ENDPOINT"
	RESET_EMAIL_SUBJECT                  string = "RESET_EMAIL_SUBJECT"
	EXPIRATION_TIME_RESET_TOKEN_IN_HOURS string = "EXPIRATION_TIME_RESET_TOKEN_IN_HOURS"
	AccessTokenDurationMinutes           string = "ACCESS_TOKEN_DURATION_MINUTES"
	RefreshTokenDurationDays             string = "REFRESH_TOKEN_DURATION_DAYS"
	REDIS_ADDR                           string = "REDIS_ADDR"
)
