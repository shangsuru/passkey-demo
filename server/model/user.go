package model

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID             `json:"id" bun:"id,pk"`
	Username            string                `json:"username" bun:"username"`
	WebauthnCredentials []WebauthnCredentials `json:"webauthn_credentials" bun:"rel:has-many,join:id=user_id"`
	PasswordHash        string                `json:"-" bun:"password_hash,notnull"`
	CreatedAt           time.Time             `json:"created_at" bun:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at" bun:"updated_at"`
}

func (u *User) WebAuthnID() []byte {
	bytes, _ := u.ID.MarshalBinary()
	return bytes
}

func (u *User) WebAuthnName() string {
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
	return u.Username
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

func (u *User) CredentialExcludeList() []protocol.CredentialDescriptor {
	var credentialExcludeList []protocol.CredentialDescriptor
	for _, cred := range u.WebauthnCredentials {
		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.CredentialID,
		}
		credentialExcludeList = append(credentialExcludeList, descriptor)
	}

	// Returns authenticators already registered to the user to prevent multiple registrations of the same authenticator
	return credentialExcludeList
}
