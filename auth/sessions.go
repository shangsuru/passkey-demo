package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const duration time.Duration = 5 * time.Minute

var sessionStore *redis.Client

func init() {
	sessionStore = redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "",
		DB:       0,
	})
}

func GetSession(ctx context.Context, id string) (*webauthn.SessionData, error) {
	bytes, err := sessionStore.Get(ctx, id).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}

	var data *webauthn.SessionData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("failed to decode session data: %v", err)
	}

	return data, nil
}

func CreateSession(ctx context.Context, data *webauthn.SessionData) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to encode session data: %v", err)
	}

	id := uuid.New().String()
	if err := sessionStore.Set(ctx, id, bytes, duration).Err(); err != nil {
		return "", fmt.Errorf("failed to save session: %v", err)
	}

	return id, nil
}

func DeleteSession(ctx context.Context, id string) error {
	if err := sessionStore.Del(ctx, id).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	return nil
}
