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

type FIDO2Response struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
}

func (wc WebAuthnController) BeginRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		username := ctx.Param("username")

		user, err := wc.UserStore.FindUserByName(ctx.Request().Context(), username)
		if err != nil {
			user, err = wc.UserStore.CreateUser(ctx.Request().Context(), username)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
					Status:       "error",
					ErrorMessage: err.Error(),
				})
			}
		}

		options, sessionData, err := wc.WebAuthnAPI.BeginRegistration(
			user,
			webauthn.WithExclusions(user.CredentialExcludeList()),
		)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		sessionID, err := CreateSession(ctx.Request().Context(), sessionData)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		ctx.SetCookie(&http.Cookie{
			Name:  "registration",
			Value: sessionID,
			Path:  "/",
		})

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) FinishRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		username := ctx.Param("username")
		user, err := wc.UserStore.FindUserByName(ctx.Request().Context(), username)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		cookie, err := ctx.Cookie("registration")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}
		sessionID := cookie.Value

		sessionData, err := GetSession(ctx.Request().Context(), sessionID)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		credential, err := wc.WebAuthnAPI.FinishRegistration(user, *sessionData, ctx.Request())
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		if err := wc.UserStore.AddWebauthnCredential(ctx.Request().Context(), user.ID, credential); err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		DeleteSession(ctx.Request().Context(), sessionID)

		return ctx.JSON(http.StatusOK, FIDO2Response{
			Status:       "ok",
			ErrorMessage: "",
		})
	}
}

func (wc WebAuthnController) BeginLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		username := ctx.Param("username")
		user, err := wc.UserStore.FindUserByName(ctx.Request().Context(), username)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		if user == nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: "User does not exist",
			})
		}

		options, sessionData, err := wc.WebAuthnAPI.BeginLogin(user)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		sessionID, err := CreateSession(ctx.Request().Context(), sessionData)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		ctx.SetCookie(&http.Cookie{
			Name:  "login",
			Value: sessionID,
			Path:  "/",
		})

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) FinishLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		username := ctx.Param("username")
		user, err := wc.UserStore.FindUserByName(ctx.Request().Context(), username)
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		cookie, err := ctx.Cookie("login")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}
		sessionID := cookie.Value

		sessionData, err := GetSession(ctx.Request().Context(), sessionID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		// in an actual implementation we should perform additional
		// checks on the returned 'credential'
		_, err = wc.WebAuthnAPI.FinishLogin(user, *sessionData, ctx.Request())
		if err != nil {
			ctx.Logger().Error(err)
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		DeleteSession(ctx.Request().Context(), sessionID)

		return ctx.JSON(http.StatusOK, FIDO2Response{
			Status:       "ok",
			ErrorMessage: "",
		})
	}
}
