package handler

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

const webauthnSessionDuration = 5 * time.Minute

type WebAuthnSession struct {
	redisClient *redis.Client
}

func NewWebAuthnSession() WebAuthnSession {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "",
		DB:       0,
	})

	return WebAuthnSession{
		redisClient: redisClient,
	}
}

func (ws *WebAuthnSession) Get(ctx echo.Context, sessionName string) (string, *webauthn.SessionData, error) {
	cookie, err := ctx.Cookie(sessionName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get session cookie: %v", err)
	}
	id := cookie.Value

	bytes, err := ws.redisClient.Get(ctx.Request().Context(), id).Bytes()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get session: %v", err)
	}

	var data *webauthn.SessionData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return "", nil, fmt.Errorf("failed to decode session data: %v", err)
	}

	return id, data, nil
}

func (ws *WebAuthnSession) Create(ctx echo.Context, sessionName string, data *webauthn.SessionData) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode session data: %v", err)
	}

	id := uuid.New().String()
	if err := ws.redisClient.Set(ctx.Request().Context(), id, bytes, webauthnSessionDuration).Err(); err != nil {
		return fmt.Errorf("failed to save session: %v", err)
	}

	ctx.SetCookie(&http.Cookie{
		Name:  sessionName,
		Value: id,
		Path:  "/",
	})

	return nil
}

func (ws *WebAuthnSession) Delete(ctx context.Context, id string) error {
	if err := ws.redisClient.Del(ctx, id).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	return nil
}
