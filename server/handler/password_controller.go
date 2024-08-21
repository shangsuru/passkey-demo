package handler

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/shangsuru/passkey-demo/repository"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"
)

type PasswordController struct {
	UserRepository repository.UserRepository
}

func (handler PasswordController) SignUp() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var p Params
		if err := ctx.Bind(&p); err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}
		email := p.Email
		if !validEmail(email) {
			return sendError(ctx, "Invalid email", http.StatusBadRequest)
		}
		password := p.Password
		if len(password) < 8 {
			return sendError(ctx, "Password must be at least 8 characters", http.StatusBadRequest)
		}

		_, err := handler.UserRepository.FindUserByEmail(ctx.Request().Context(), email)
		if err == nil {
			return sendError(ctx, "An account with that email already exists.", http.StatusConflict)
		}

		passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
		if err != nil {
			return sendError(ctx, "Internal server error", http.StatusInternalServerError)
		}

		user, err := handler.UserRepository.CreateUser(ctx.Request().Context(), email, passwordHash)
		if err != nil {
			return sendError(ctx, "Internal server error", http.StatusInternalServerError)
		}

		if err = handler.createSession(ctx, user.ID.String()); err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return sendOK(ctx)
	}
}

func (handler PasswordController) Login() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var p Params
		if err := ctx.Bind(&p); err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}
		email := p.Email
		if !validEmail(email) {
			return sendError(ctx, "Invalid email", http.StatusBadRequest)
		}

		user, err := handler.UserRepository.FindUserByEmail(ctx.Request().Context(), email)
		if err != nil {
			return sendError(ctx, "An account with that email does not exist.", http.StatusNotFound)
		}

		match, err := argon2id.ComparePasswordAndHash(p.Password, user.PasswordHash)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if !match {
			return sendError(ctx, "Invalid password.", http.StatusUnauthorized)
		}

		if err = handler.createSession(ctx, user.ID.String()); err != nil {
			return sendError(ctx, "Session could not be created", http.StatusInternalServerError)
		}

		return sendOK(ctx)
	}
}

func (handler PasswordController) Logout() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if err := handler.terminateSession(ctx); err != nil {
			return sendError(ctx, "Not logged in.", http.StatusUnauthorized)
		}
		return sendOK(ctx)
	}
}

func (handler PasswordController) createSession(ctx echo.Context, userId string) error {
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

func (handler PasswordController) terminateSession(ctx echo.Context) error {
	sess, err := session.Get("auth", ctx)
	if err != nil {
		return err
	}

	sess.Values["user"] = nil // Revoke users authentication
	return sess.Save(ctx.Request(), ctx.Response())
}
