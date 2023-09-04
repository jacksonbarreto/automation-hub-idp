package config

const (
	redisAddr string = "REDIS_ADDR"
)

type redisConfig struct {
	RedisAddr string
}

func newRedisConfig() (*redisConfig, error) {
	return &redisConfig{
		RedisAddr: getEnvString(redisAddr, "redis:6379"),
	}, nil
}
