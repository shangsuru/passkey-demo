package handler

import (
	"github.com/alexedwards/argon2id"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/shangsuru/passkey-demo/repository"
	"net/http"
)

type WebAuthnController struct {
	UserRepository    repository.UserRepository
	WebAuthnAPI       *webauthn.WebAuthn
	SessionRepository repository.SessionRepository
}

func (wc WebAuthnController) BeginRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var p Params
		if err := ctx.Bind(&p); err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}
		email := p.Email
		if !validEmail(email) {
			return sendError(ctx, "Invalid email.", http.StatusBadRequest)
		}

		_, err := wc.UserRepository.FindUserByEmail(ctx.Request().Context(), email)
		if err == nil {
			return sendError(ctx, "An account with that email already exists.", http.StatusConflict)
		}

		// create a random password to fulfill not null constraint in user.go
		passwordHash, err := argon2id.CreateHash(random.String(20), argon2id.DefaultParams)
		if err != nil {
			return sendError(ctx, "Internal server error", http.StatusInternalServerError)
		}

		user, err := wc.UserRepository.CreateUser(ctx.Request().Context(), email, passwordHash)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		authSelect := protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyRequired(),
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			UserVerification:   protocol.VerificationRequired,
		}

		options, sessionData, err := wc.WebAuthnAPI.BeginRegistration(
			user,
			webauthn.WithAuthenticatorSelection(authSelect),
			webauthn.WithExclusions(user.CredentialExcludeList()),
		)
		if err != nil {
			_ = wc.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		err = wc.SessionRepository.CreateWebauthnSession(ctx, "registration", sessionData)
		if err != nil {
			_ = wc.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) FinishRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := wc.SessionRepository.GetWebauthnSession(ctx, "registration")
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}

		user, err := wc.UserRepository.FindUserByID(ctx.Request().Context(), sessionData.UserID)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		credential, err := wc.WebAuthnAPI.FinishRegistration(user, *sessionData, ctx.Request())
		if err != nil {
			_ = wc.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			_ = wc.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, "User not present or not verified.", http.StatusBadRequest)
		}

		if err := wc.UserRepository.AddWebauthnCredential(ctx.Request().Context(), user.ID, credential); err != nil {
			_ = wc.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		_ = wc.SessionRepository.DeleteSession(ctx.Request().Context(), sessionID)

		if err = wc.SessionRepository.Login(ctx, user.ID); err != nil {
			_ = wc.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

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
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if err := wc.SessionRepository.CreateWebauthnSession(ctx, "login", sessionData); err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (wc WebAuthnController) assertionResult(getCredential func(ctx echo.Context, sessionData *webauthn.SessionData) (*webauthn.Credential, error)) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := wc.SessionRepository.GetWebauthnSession(ctx, "login")
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}

		credential, err := getCredential(ctx, sessionData)
		if err != nil {
			return sendError(ctx, "There is no passkey associated with this account.", http.StatusNotFound)
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			return sendError(ctx, "User not present or not verified.", http.StatusBadRequest)
		}

		if credential.Authenticator.CloneWarning {
			return sendError(ctx, "Authenticator is cloned.", http.StatusBadRequest)
		}

		_ = wc.SessionRepository.DeleteSession(ctx.Request().Context(), sessionID)

		userID, err := wc.UserRepository.FindUserIDByCredentialID(ctx.Request().Context(), credential.ID)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if err = wc.SessionRepository.Login(ctx, *userID); err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return sendOK(ctx)
	}
}

func (wc WebAuthnController) getCredentialAssertion(ctx echo.Context) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	var p Params
	if err := ctx.Bind(&p); err != nil {
		return nil, nil, err
	}

	user, err := wc.UserRepository.FindUserByEmail(ctx.Request().Context(), p.Email)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		return nil, nil, err
	}

	options, sessionData, err := wc.WebAuthnAPI.BeginLogin(user, webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return nil, nil, err
	}

	return options, sessionData, nil
}

func (wc WebAuthnController) getDiscoverableCredentialAssertion(echo.Context) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	options, sessionData, err := wc.WebAuthnAPI.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
	return options, sessionData, err
}

func (wc WebAuthnController) getCredential(ctx echo.Context, sessionData *webauthn.SessionData) (*webauthn.Credential, error) {
	user, err := wc.UserRepository.FindUserByID(ctx.Request().Context(), sessionData.UserID)
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
		return wc.UserRepository.FindUserByID(ctx.Request().Context(), userID)
	}, *sessionData, ctx.Request())
	return credential, err
}
