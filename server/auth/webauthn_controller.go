package auth

import (
	"net/http"
	"net/mail"

	"github.com/go-webauthn/webauthn/protocol"
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

type Params struct {
	Email string
}

func (wc WebAuthnController) BeginRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var p Params
		if err := ctx.Bind(&p); err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}
		email := p.Email
		if !validEmail(email) {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: "Invalid email",
			})
		}

		user, err := wc.UserStore.FindUserByEmail(ctx.Request().Context(), email)
		if err != nil {
			user, err = wc.UserStore.CreateUser(ctx.Request().Context(), email)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
					Status:       "error",
					ErrorMessage: err.Error(),
				})
			}
		}

		authSelect := protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyRequired(),
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
		}

		options, sessionData, err := wc.WebAuthnAPI.BeginRegistration(
			user,
			webauthn.WithAuthenticatorSelection(authSelect),
			webauthn.WithExclusions(user.CredentialExcludeList()),
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		err = CreateSession(ctx, "registration", sessionData)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) FinishRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := GetSession(ctx, "registration")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		user, err := wc.UserStore.FindUserByID(ctx.Request().Context(), sessionData.UserID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		credential, err := wc.WebAuthnAPI.FinishRegistration(user, *sessionData, ctx.Request())
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: "User not present or not verified",
			})
		}

		if err := wc.UserStore.AddWebauthnCredential(ctx.Request().Context(), user.ID, credential); err != nil {
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

func (wc WebAuthnController) BeginDiscoverableLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		options, sessionData, err := wc.WebAuthnAPI.BeginDiscoverableLogin()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		if err := CreateSession(ctx, "login", sessionData); err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) BeginLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var p Params
		if err := ctx.Bind(&p); err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		user, err := wc.UserStore.FindUserByEmail(ctx.Request().Context(), p.Email)
		if err != nil {
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
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		if err := CreateSession(ctx, "login", sessionData); err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) FinishDiscoverableLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := GetSession(ctx, "login")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		credential, err := wc.WebAuthnAPI.FinishDiscoverableLogin(func(rawId []byte, userID []byte) (user webauthn.User, err error) {
			return wc.UserStore.FindUserByID(ctx.Request().Context(), userID)
		}, *sessionData, ctx.Request())
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: "User not present or not verified",
			})
		}

		if credential.Authenticator.CloneWarning {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: "Authenticator is cloned",
			})
		}

		DeleteSession(ctx.Request().Context(), sessionID)

		return ctx.JSON(http.StatusOK, FIDO2Response{
			Status:       "ok",
			ErrorMessage: "",
		})
	}
}

func (wc WebAuthnController) FinishLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := GetSession(ctx, "login")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		user, err := wc.UserStore.FindUserByID(ctx.Request().Context(), sessionData.UserID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		credential, err := wc.WebAuthnAPI.FinishLogin(user, *sessionData, ctx.Request())
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: err.Error(),
			})
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: "User not present or not verified",
			})
		}

		if credential.Authenticator.CloneWarning {
			return ctx.JSON(http.StatusBadRequest, FIDO2Response{
				Status:       "error",
				ErrorMessage: "Authenticator is cloned",
			})
		}

		DeleteSession(ctx.Request().Context(), sessionID)

		return ctx.JSON(http.StatusOK, FIDO2Response{
			Status:       "ok",
			ErrorMessage: "",
		})
	}
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
