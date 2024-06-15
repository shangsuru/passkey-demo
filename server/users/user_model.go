package users

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// https://github.com/go-webauthn/webauthn/blob/master/webauthn/credential.go
type WebauthnCredentials struct {
	ID              uuid.UUID                         `json:"id" bun:"id,pk"`
	UserID          uuid.UUID                         `json:"user_id" bun:"user_id"`
	CredentialID    []byte                            `json:"credential_id" bun:"credential_id"`
	PublicKey       []byte                            `json:"public_key" bun:"public_key"`
	AttestationType string                            `json:"attestation_type" bun:"attestation_type"`
	Transport       []protocol.AuthenticatorTransport `json:"transport" bun:"transport,array"`
	Flags           webauthn.CredentialFlags          `json:"flags" bun:"flags"`
	Authenticator   webauthn.Authenticator            `json:"authenticator" bun:"authenticator"`
}

type User struct {
	ID                  uuid.UUID             `json:"id" bun:"id,pk"`
	Name                string                `json:"name" bun:"name"`
	WebauthnCredentials []WebauthnCredentials `json:"webauthn_credentials" bun:"rel:has-many,join:id=user_id"`
	CreatedAt           time.Time             `json:"created_at" bun:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at" bun:"updated_at"`
}

func (u *User) WebAuthnID() []byte {
	bytes, _ := u.ID.MarshalBinary()
	return bytes
}

func (u *User) WebAuthnName() string {
	return u.Name
}

func (u *User) WebAuthnDisplayName() string {
	return u.Name
}

func (u *User) WebAuthnIcon() string {
	return ""
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	credentials := make([]webauthn.Credential, len(u.WebauthnCredentials))

	for i, v := range u.WebauthnCredentials {
		credentials[i] = webauthn.Credential{
			ID:              v.CredentialID,
			PublicKey:       v.PublicKey,
			AttestationType: v.AttestationType,
			Transport:       v.Transport,
			Flags:           v.Flags,
			Authenticator:   v.Authenticator,
		}
	}

	return credentials
}

// Returns authenticators already registered to the user
// to prevent multiple registrations of the same authenticator
func (u *User) CredentialExcludeList() []protocol.CredentialDescriptor {
	credentialExcludeList := []protocol.CredentialDescriptor{}
	for _, cred := range u.WebauthnCredentials {
		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.CredentialID,
		}
		credentialExcludeList = append(credentialExcludeList, descriptor)
	}

	return credentialExcludeList
}
