package auth

import (
	"github.com/go-webauthn/webauthn/webauthn"
)

func NewWebAuthnAPI() (*webauthn.WebAuthn, error) {
	webAuthnAPI, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "PassKey Demo",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	})
	return webAuthnAPI, err
}
