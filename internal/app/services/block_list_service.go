package services

import (
	"context"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)
import "idp-automations-hub/internal/app/services/iservice"

type tokenBlockListServiceImpl struct {
	client *redis.Client
	ctx    context.Context
}

func NewTokenBlockListService() iservice.TokenBlockListService {
	redisAddr := os.Getenv("REDIS_ADDR")
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
	err := r.client.Set(r.ctx, jwtUUID, true, expirationTime).Err()
	return err
}

func (r *tokenBlockListServiceImpl) IsInBlockList(jwtUUID string) (bool, error) {
	val, err := r.client.Get(r.ctx, jwtUUID).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return val == "true", err
}
