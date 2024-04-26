package auth

import (
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"

	"github.com/shangsuru/passkey-demo/users"
)

type WebAuthnController struct {
	UserStore   users.UserRepository
	WebAuthnAPI *webauthn.WebAuthn
}

func (wc WebAuthnController) BeginRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		username := ctx.Param("username")

		user, err := wc.UserStore.FindUserByName(ctx.Request().Context(), username)
		if err != nil {
			user, err = wc.UserStore.CreateUser(ctx.Request().Context(), username)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err.Error())
			}
		}

		options, sessionData, err := wc.WebAuthnAPI.BeginRegistration(
			user,
			webauthn.WithExclusions(user.CredentialExcludeList()),
		)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		if err := CreateSession(ctx.Request().Context(), username, sessionData); err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) FinishRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		username := ctx.Param("username")
		user, err := wc.UserStore.FindUserByName(ctx.Request().Context(), username)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		sessionData, err := GetSession(ctx.Request().Context(), username)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		credential, err := wc.WebAuthnAPI.FinishRegistration(user, *sessionData, ctx.Request())
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		if err := wc.UserStore.AddWebauthnCredential(ctx.Request().Context(), user.ID, credential); err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusOK, nil)
	}
}

func (wc WebAuthnController) BeginLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		username := ctx.Param("username")
		user, err := wc.UserStore.FindUserByName(ctx.Request().Context(), username)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		options, sessionData, err := wc.WebAuthnAPI.BeginLogin(user)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		if err := CreateSession(ctx.Request().Context(), username, sessionData); err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) FinishLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		username := ctx.Param("username")
		user, err := wc.UserStore.FindUserByName(ctx.Request().Context(), username)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		sessionData, err := GetSession(ctx.Request().Context(), username)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		// in an actual implementation we should perform additional
		// checks on the returned 'credential'
		_, err = wc.WebAuthnAPI.FinishLogin(user, *sessionData, ctx.Request())
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		return ctx.JSON(http.StatusOK, nil)
	}
}
