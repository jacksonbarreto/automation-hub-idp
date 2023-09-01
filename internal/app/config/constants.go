package config

const (
	LoginAttemptBase                string = "LOGIN_ATTEMPT_BASE"
	MaxLoginAttemptsBeforeBlock     string = "MAX_LOGIN_ATTEMPTS_BEFORE_BLOCK"
	MinTimeBetweenAttemptsInSeconds string = "MIN_TIME_BETWEEN_ATTEMPTS_IN_SECONDS"
	ExpirationTimeResetTokenInHours string = "EXPIRATION_TIME_RESET_TOKEN_IN_HOURS"
	AccessTokenDurationMinutes      string = "ACCESS_TOKEN_DURATION_MINUTES"
	RefreshTokenDurationDays        string = "REFRESH_TOKEN_DURATION_DAYS"
	RedisAddr                       string = "REDIS_ADDR"
	PasswordResetTopic              string = "PASSWORD_RESET_TOPIC"
	AccountBlockedTopic             string = "ACCOUNT_BLOCKED_TOPIC"
	AccountCreatedTopic             string = "ACCOUNT_CREATED_TOPIC"
)
