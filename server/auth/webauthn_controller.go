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
			return sendError(ctx, err.Error())
		}
		email := p.Email
		if !validEmail(email) {
			return sendError(ctx, "Invalid email")
		}

		user, err := wc.UserStore.FindUserByEmail(ctx.Request().Context(), email)
		if err != nil {
			user, err = wc.UserStore.CreateUser(ctx.Request().Context(), email)
			if err != nil {
				return sendError(ctx, err.Error())
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
			return sendError(ctx, err.Error())
		}

		err = CreateSession(ctx, "registration", sessionData)
		if err != nil {
			return sendError(ctx, err.Error())
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) FinishRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := GetSession(ctx, "registration")
		if err != nil {
			return sendError(ctx, err.Error())
		}

		user, err := wc.UserStore.FindUserByID(ctx.Request().Context(), sessionData.UserID)
		if err != nil {
			return sendError(ctx, err.Error())
		}

		credential, err := wc.WebAuthnAPI.FinishRegistration(user, *sessionData, ctx.Request())
		if err != nil {
			return sendError(ctx, err.Error())
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			return sendError(ctx, "User not present or not verified")
		}

		if err := wc.UserStore.AddWebauthnCredential(ctx.Request().Context(), user.ID, credential); err != nil {
			return sendError(ctx, err.Error())
		}

		DeleteSession(ctx.Request().Context(), sessionID)

		return sendOK(ctx)
	}
}

func (wc WebAuthnController) BeginLogin() echo.HandlerFunc {
	return wc.assertionOptions(wc.getCredentialAssertion)
}

func (wc WebAuthnController) FinishLogin() echo.HandlerFunc {
	return wc.assertionResult(wc.getCredential)
}

func (wc WebAuthnController) BeginDiscoverableLogin() echo.HandlerFunc {
	return wc.assertionOptions(wc.getDiscoverableCredentialAssertion)
}

func (wc WebAuthnController) FinishDiscoverableLogin() echo.HandlerFunc {
	return wc.assertionResult(wc.getDiscoverableCredential)
}

func (wc WebAuthnController) assertionOptions(getCredentialAssertion func(ctx echo.Context) (*protocol.CredentialAssertion, *webauthn.SessionData, error)) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		options, sessionData, err := getCredentialAssertion(ctx)
		if err != nil {
			return sendError(ctx, err.Error())
		}

		if err := CreateSession(ctx, "login", sessionData); err != nil {
			return sendError(ctx, err.Error())
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) assertionResult(getCredential func(ctx echo.Context, sessionData *webauthn.SessionData) (*webauthn.Credential, error)) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := GetSession(ctx, "login")
		if err != nil {
			return sendError(ctx, err.Error())
		}

		credential, err := getCredential(ctx, sessionData)
		if err != nil {
			return sendError(ctx, err.Error())
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			return sendError(ctx, "User not present or not verified")
		}

		if credential.Authenticator.CloneWarning {
			return sendError(ctx, "Authenticator is cloned")
		}

		DeleteSession(ctx.Request().Context(), sessionID)

		return sendOK(ctx)
	}
}

func (wc WebAuthnController) getCredentialAssertion(ctx echo.Context) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	var p Params
	if err := ctx.Bind(&p); err != nil {
		return nil, nil, err
	}

	user, err := wc.UserStore.FindUserByEmail(ctx.Request().Context(), p.Email)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		return nil, nil, err
	}

	options, sessionData, err := wc.WebAuthnAPI.BeginLogin(user)
	if err != nil {
		return nil, nil, err
	}

	return options, sessionData, nil
}

func (wc WebAuthnController) getDiscoverableCredentialAssertion(ctx echo.Context) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	options, sessionData, err := wc.WebAuthnAPI.BeginDiscoverableLogin()
	return options, sessionData, err
}

func (wc WebAuthnController) getCredential(ctx echo.Context, sessionData *webauthn.SessionData) (*webauthn.Credential, error) {
	user, err := wc.UserStore.FindUserByID(ctx.Request().Context(), sessionData.UserID)
	if err != nil {
		return nil, err
	}

	credential, err := wc.WebAuthnAPI.FinishLogin(user, *sessionData, ctx.Request())
	if err != nil {
		return nil, err
	}
	return credential, nil
}

func (wc WebAuthnController) getDiscoverableCredential(ctx echo.Context, sessionData *webauthn.SessionData) (*webauthn.Credential, error) {
	credential, err := wc.WebAuthnAPI.FinishDiscoverableLogin(func(rawId []byte, userID []byte) (user webauthn.User, err error) {
		return wc.UserStore.FindUserByID(ctx.Request().Context(), userID)
	}, *sessionData, ctx.Request())
	return credential, err
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func sendError(ctx echo.Context, err string) error {
	return ctx.JSON(http.StatusBadRequest, FIDO2Response{
		Status:       "error",
		ErrorMessage: err,
	})
}

func sendOK(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, FIDO2Response{
		Status:       "ok",
		ErrorMessage: "",
	})
}
