package handler

import (
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/shangsuru/passkey-demo/repository"
)

type WebAuthnController struct {
	UserRepository  repository.UserRepository
	WebAuthnAPI     *webauthn.WebAuthn
	WebAuthnSession WebAuthnSession
}

func (handler WebAuthnController) BeginRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var p Params
		if err := ctx.Bind(&p); err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}
		username := p.Username
		if len(username) == 0 {
			return sendError(ctx, "Empty username", http.StatusBadRequest)
		}

		_, err := handler.UserRepository.FindUserByUsername(ctx.Request().Context(), username)
		if err == nil {
			return sendError(ctx, "An account with that username already exists.", http.StatusConflict)
		}

		// create a random password to fulfill not null constraint in user_repository.go
		passwordHash, err := argon2id.CreateHash(random.String(20), argon2id.DefaultParams)
		if err != nil {
			return sendError(ctx, "Internal server error", http.StatusInternalServerError)
		}

		user, err := handler.UserRepository.CreateUser(ctx.Request().Context(), username, passwordHash)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		authSelect := protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyRequired(),
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			UserVerification:   protocol.VerificationRequired,
		}

		options, sessionData, err := handler.WebAuthnAPI.BeginRegistration(
			user,
			webauthn.WithAuthenticatorSelection(authSelect),
			webauthn.WithExclusions(user.CredentialExcludeList()),
		)
		if err != nil {
			_ = handler.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		err = handler.WebAuthnSession.Create(ctx, "registration", sessionData)
		if err != nil {
			_ = handler.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (handler WebAuthnController) FinishRegistration() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := handler.WebAuthnSession.Get(ctx, "registration")
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}

		user, err := handler.UserRepository.FindUserByID(ctx.Request().Context(), sessionData.UserID)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		credential, err := handler.WebAuthnAPI.FinishRegistration(user, *sessionData, ctx.Request())
		if err != nil {
			_ = handler.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			_ = handler.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, "User not present or not verified.", http.StatusBadRequest)
		}

		if err := handler.UserRepository.AddWebauthnCredential(ctx.Request().Context(), user.ID, credential); err != nil {
			_ = handler.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		_ = handler.WebAuthnSession.Delete(ctx.Request().Context(), sessionID)

		if err = createSession(ctx, user.ID.String()); err != nil {
			_ = handler.UserRepository.DeleteUser(ctx.Request().Context(), user)
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return sendOK(ctx)
	}
}

func (handler WebAuthnController) BeginLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		options, sessionData, err := handler.getCredentialAssertion(ctx)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if err := handler.WebAuthnSession.Create(ctx, "login", sessionData); err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (handler WebAuthnController) FinishLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := handler.WebAuthnSession.Get(ctx, "login")
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}

		credential, err := handler.getCredential(ctx, sessionData)
		if err != nil {
			return sendError(ctx, "There is no passkey associated with this account.", http.StatusNotFound)
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			return sendError(ctx, "User not present or not verified.", http.StatusBadRequest)
		}

		if credential.Authenticator.CloneWarning {
			return sendError(ctx, "Authenticator is cloned.", http.StatusBadRequest)
		}

		_ = handler.WebAuthnSession.Delete(ctx.Request().Context(), sessionID)

		userID, err := handler.UserRepository.FindUserIDByCredentialID(ctx.Request().Context(), credential.ID)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if err = createSession(ctx, (*userID).String()); err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return sendOK(ctx)
	}
}

func (handler WebAuthnController) BeginDiscoverableLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		options, sessionData, err := handler.getDiscoverableCredentialAssertion()
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if err := handler.WebAuthnSession.Create(ctx, "login", sessionData); err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return ctx.JSON(http.StatusOK, options)
	}
}

func (handler WebAuthnController) FinishDiscoverableLogin() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sessionID, sessionData, err := handler.WebAuthnSession.Get(ctx, "login")
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusBadRequest)
		}

		credential, err := handler.getDiscoverableCredential(ctx, sessionData)
		if err != nil {
			return sendError(ctx, "There is no passkey associated with this account.", http.StatusNotFound)
		}

		if !credential.Flags.UserPresent || !credential.Flags.UserVerified {
			return sendError(ctx, "User not present or not verified.", http.StatusBadRequest)
		}

		if credential.Authenticator.CloneWarning {
			return sendError(ctx, "Authenticator is cloned.", http.StatusBadRequest)
		}

		_ = handler.WebAuthnSession.Delete(ctx.Request().Context(), sessionID)

		userID, err := handler.UserRepository.FindUserIDByCredentialID(ctx.Request().Context(), credential.ID)
		if err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		if err = createSession(ctx, (*userID).String()); err != nil {
			return sendError(ctx, err.Error(), http.StatusInternalServerError)
		}

		return sendOK(ctx)
	}
}

func (handler WebAuthnController) getCredentialAssertion(ctx echo.Context) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	var p Params
	if err := ctx.Bind(&p); err != nil {
		return nil, nil, err
	}

	user, err := handler.UserRepository.FindUserByUsername(ctx.Request().Context(), p.Username)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		return nil, nil, err
	}

	options, sessionData, err := handler.WebAuthnAPI.BeginLogin(user, webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return nil, nil, err
	}

	return options, sessionData, nil
}

func (handler WebAuthnController) getDiscoverableCredentialAssertion() (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	options, sessionData, err := handler.WebAuthnAPI.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
	return options, sessionData, err
}

func (handler WebAuthnController) getCredential(ctx echo.Context, sessionData *webauthn.SessionData) (*webauthn.Credential, error) {
	user, err := handler.UserRepository.FindUserByID(ctx.Request().Context(), sessionData.UserID)
	if err != nil {
		return nil, err
	}

	credential, err := handler.WebAuthnAPI.FinishLogin(user, *sessionData, ctx.Request())
	if err != nil {
		return nil, err
	}
	return credential, nil
}

func (handler WebAuthnController) getDiscoverableCredential(ctx echo.Context, sessionData *webauthn.SessionData) (*webauthn.Credential, error) {
	credential, err := handler.WebAuthnAPI.FinishDiscoverableLogin(func(rawId []byte, userID []byte) (user webauthn.User, err error) {
		return handler.UserRepository.FindUserByID(ctx.Request().Context(), userID)
	}, *sessionData, ctx.Request())
	return credential, err
}
