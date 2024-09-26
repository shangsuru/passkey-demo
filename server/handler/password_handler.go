package handler

import (
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

		if err = createSession(ctx, user.ID.String()); err != nil {
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

		if err = createSession(ctx, user.ID.String()); err != nil {
			return sendError(ctx, "Session could not be created", http.StatusInternalServerError)
		}

		return sendOK(ctx)
	}
}

func (handler PasswordController) Logout() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if err := terminateSession(ctx); err != nil {
			return sendError(ctx, "Not logged in.", http.StatusUnauthorized)
		}
		return sendOK(ctx)
	}
}
