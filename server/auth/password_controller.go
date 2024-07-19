package auth

import (
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"

	"github.com/shangsuru/passkey-demo/users"
)

type PasswordController struct {
	UserStore users.UserRepository
}

func (pc PasswordController) SignUp() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var p AuthParams
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

		_, err := pc.UserStore.FindUserByEmail(ctx.Request().Context(), email)
		if err == nil {
			return sendError(ctx, "An account with that email already exists.", http.StatusConflict)
		}

		passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
		if err != nil {
			return sendError(ctx, "Internal server error", http.StatusInternalServerError)
		}

		_, err = pc.UserStore.CreateUser(ctx.Request().Context(), email, passwordHash)
		if err != nil {
			return sendError(ctx, "Internal server error", http.StatusInternalServerError)
		}

		return sendOK(ctx)
	}
}

func (pc PasswordController) Login() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var p AuthParams
		if err := ctx.Bind(&p); err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}
		email := p.Email
		if !validEmail(email) {
			return sendError(ctx, "Invalid email", http.StatusBadRequest)
		}

		user, err := pc.UserStore.FindUserByEmail(ctx.Request().Context(), email)
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

		return sendOK(ctx)
	}
}
