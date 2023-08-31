package services

import (
	"context"
	"github.com/go-redis/redis/v8"
	"idp-automations-hub/internal/app/config"
	"os"
	"time"
)
import "idp-automations-hub/internal/app/services/iservice"

type tokenBlockListServiceImpl struct {
	client *redis.Client
	ctx    context.Context
}

func NewTokenBlockListService() iservice.TokenBlockListService {
	redisAddr := os.Getenv(config.RedisAddr)
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.TODO()

	return &tokenBlockListServiceImpl{
		client: rdb,
		ctx:    ctx,
	}
}

func (r *tokenBlockListServiceImpl) AddToBlockList(jwtUUID string, expirationTime time.Duration) error {
	err := r.client.Set(r.ctx, jwtUUID, 1, expirationTime).Err()
	return err
}

func (r *tokenBlockListServiceImpl) IsInBlockList(jwtUUID string) (bool, error) {
	_, err := r.client.Get(r.ctx, jwtUUID).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, err
}
