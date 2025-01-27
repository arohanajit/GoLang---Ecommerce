package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/arohanajit/e-commerce-platform/pkg/redis"
)

type UserCache struct {
	redisClient *redis.Client
}

func NewUserCache(redisClient *redis.Client) *UserCache {
	return &UserCache{redisClient: redisClient}
}

func (uc *UserCache) GetUser(ctx context.Context, userID string) (*User, error) {
	var user User
	err := uc.redisClient.Get(ctx, fmt.Sprintf("user:%s", userID), &user)
	return &user, err
}

func (uc *UserCache) SetUser(ctx context.Context, user *User) error {
	return uc.redisClient.Set(ctx, fmt.Sprintf("user:%s", user.ID), user, 30*time.Minute)
}

func (uc *UserCache) InvalidateUser(ctx context.Context, userID string) error {
	return uc.redisClient.Delete(ctx, fmt.Sprintf("user:%s", userID))
}
