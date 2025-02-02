package session

import (
	"context"
	"time"

	"e-commerce-platform/pkg/redis"
)

type Session struct {
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionManager struct {
	redisClient *redis.Client
}

func NewSessionManager(redisClient *redis.Client) *SessionManager {
	return &SessionManager{redisClient: redisClient}
}

func (sm *SessionManager) CreateSession(ctx context.Context, userID, token string, duration time.Duration) error {
	session := Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(duration),
	}

	return sm.redisClient.Set(ctx, "session:"+token, session, duration)
}

func (sm *SessionManager) GetSession(ctx context.Context, token string) (*Session, error) {
	var session Session
	err := sm.redisClient.Get(ctx, "session:"+token, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (sm *SessionManager) DeleteSession(ctx context.Context, token string) error {
	return sm.redisClient.Delete(ctx, "session:"+token)
}
