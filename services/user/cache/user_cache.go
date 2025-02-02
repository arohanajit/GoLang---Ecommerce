package cache

import (
	"context"
	"fmt"
	"time"

	"e-commerce-platform/pkg/redis"

	"github.com/arohanajit/user-service/models"
)

type UserCache struct {
	redisClient *redis.Client
}

func NewUserCache(redisClient *redis.Client) *UserCache {
	return &UserCache{redisClient: redisClient}
}

func (uc *UserCache) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := uc.redisClient.Get(ctx, fmt.Sprintf("user:%s", userID), &user)
	return &user, err
}

func (uc *UserCache) SetUser(ctx context.Context, user *models.User) error {
	return uc.redisClient.Set(ctx, fmt.Sprintf("user:%s", user.ID), user, 30*time.Minute)
}

func (uc *UserCache) InvalidateUser(ctx context.Context, userID string) error {
	return uc.redisClient.Delete(ctx, fmt.Sprintf("user:%s", userID))
}
