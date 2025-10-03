package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shop/gateway/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisSessionRepository struct {
	client *redis.Client
	prefix string
}

func NewRedisSessionRepository(addr, password, prefix string) (*RedisSessionRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisSessionRepository{
		client: client,
		prefix: prefix,
	}, nil
}

func (r *RedisSessionRepository) CreateSession(ctx context.Context, session *model.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := r.getKey(session.ID)
	expiration := time.Until(session.ExpiresAt)

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set session in redis: %w", err)
	}

	return nil
}

func (r *RedisSessionRepository) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	key := r.getKey(sessionID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session from redis: %w", err)
	}

	var session model.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (r *RedisSessionRepository) UpdateSession(ctx context.Context, session *model.Session) error {
	return r.CreateSession(ctx, session)
}

func (r *RedisSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	key := r.getKey(sessionID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisSessionRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	keys, err := r.client.Keys(ctx, r.prefix+"*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		session, err := r.GetSession(ctx, key[len(r.prefix):])
		if err != nil {
			continue
		}
		if session.UserID == userID {
			if err := r.DeleteSession(ctx, session.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *RedisSessionRepository) getKey(sessionID string) string {
	return r.prefix + sessionID
}
