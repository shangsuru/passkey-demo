package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/random"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

const webauthnSessionDuration = 5 * time.Minute
const sessionDuration = 30 * 24 * time.Hour

type SessionRepository struct {
	redisClient *redis.Client
}

func NewSessionRepository() SessionRepository {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "",
		DB:       0,
	})

	return SessionRepository{
		redisClient: redisClient,
	}
}

func (sr *SessionRepository) GetWebauthnSession(ctx echo.Context, sessionName string) (string, *webauthn.SessionData, error) {
	cookie, err := ctx.Cookie(sessionName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get session cookie: %v", err)
	}
	id := cookie.Value

	bytes, err := sr.redisClient.Get(ctx.Request().Context(), id).Bytes()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get session: %v", err)
	}

	var data *webauthn.SessionData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return "", nil, fmt.Errorf("failed to decode session data: %v", err)
	}

	return id, data, nil
}

func (sr *SessionRepository) CreateWebauthnSession(ctx echo.Context, sessionName string, data *webauthn.SessionData) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode session data: %v", err)
	}

	id := uuid.New().String()
	if err := sr.redisClient.Set(ctx.Request().Context(), id, bytes, webauthnSessionDuration).Err(); err != nil {
		return fmt.Errorf("failed to save session: %v", err)
	}

	ctx.SetCookie(&http.Cookie{
		Name:  sessionName,
		Value: id,
		Path:  "/",
	})

	return nil
}

func (sr *SessionRepository) Login(ctx echo.Context, userID uuid.UUID) error {
	sessionID := random.String(20)
	if err := sr.redisClient.Set(ctx.Request().Context(), sessionID, userID.String(), sessionDuration).Err(); err != nil {
		return err
	}

	ctx.SetCookie(&http.Cookie{
		Name:  "webauthn",
		Value: sessionID,
		Path:  "/",
	})
	return nil
}

func (sr *SessionRepository) Logout(ctx context.Context, sessionID string) {
	_ = sr.DeleteSession(ctx, sessionID)
}

func (sr *SessionRepository) DeleteSession(ctx context.Context, id string) error {
	if err := sr.redisClient.Del(ctx, id).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	return nil
}
