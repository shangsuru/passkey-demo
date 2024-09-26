package handler

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"net/http"
	"net/mail"

	"github.com/labstack/echo/v4"
)

type Params struct {
	Email    string
	Password string
}

type Response struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
}

func sendError(ctx echo.Context, err string, code int) error {
	return ctx.JSON(code, Response{
		Status:       "error",
		ErrorMessage: err,
	})
}

func sendOK(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, Response{
		Status:       "ok",
		ErrorMessage: "",
	})
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func createSession(ctx echo.Context, userId string) error {
	sess, err := session.Get("auth", ctx)
	if err != nil {
		return err
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
	}
	sess.Values["user"] = userId
	return sess.Save(ctx.Request(), ctx.Response())
}

func terminateSession(ctx echo.Context) error {
	sess, err := session.Get("auth", ctx)
	if err != nil {
		return err
	}

	sess.Values["user"] = nil // Revoke users authentication
	return sess.Save(ctx.Request(), ctx.Response())
}
