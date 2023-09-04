package config

import (
	"os"
	"strconv"
)

const (
	RedisAddr string = "REDIS_ADDR"
)

var (
	AuthenticationConfig *authenticationConfig
)

func Setup() error {
	var err error
	AuthenticationConfig, err = newAuthenticationConfig()
	if err != nil {
		return err
	}

	return nil
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		intVal, err := strconv.Atoi(value)
		if err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
