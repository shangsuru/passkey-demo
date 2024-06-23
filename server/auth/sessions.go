package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

func GetSession(ctx echo.Context, sessionName string) (string, *webauthn.SessionData, error) {
	cookie, err := ctx.Cookie(sessionName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get session cookie: %v", err)
	}
	id := cookie.Value

	bytes, err := sessionStore.Get(ctx.Request().Context(), id).Bytes()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get session: %v", err)
	}

	var data *webauthn.SessionData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return "", nil, fmt.Errorf("failed to decode session data: %v", err)
	}

	return id, data, nil
}

func CreateSession(ctx echo.Context, sessionName string, data *webauthn.SessionData) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode session data: %v", err)
	}

	id := uuid.New().String()
	if err := sessionStore.Set(ctx.Request().Context(), id, bytes, duration).Err(); err != nil {
		return fmt.Errorf("failed to save session: %v", err)
	}

	ctx.SetCookie(&http.Cookie{
		Name:  sessionName,
		Value: id,
		Path:  "/",
	})

	return nil
}

func DeleteSession(ctx context.Context, id string) error {
	if err := sessionStore.Del(ctx, id).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	return nil
}
