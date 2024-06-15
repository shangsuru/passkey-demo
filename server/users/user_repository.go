package users

import (
	"context"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserRepository struct {
	DB *bun.DB
}

func (ur *UserRepository) FindUserByName(ctx context.Context, name string) (*User, error) {
	var user User
	err := ur.DB.NewSelect().
		Model(&user).
		Relation("WebauthnCredentials").
		Column("*").
		Where("name = ?", name).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) CreateUser(ctx context.Context, name string) (*User, error) {
	user := &User{
		Name: name,
	}

	_, err := ur.DB.NewInsert().Model(user).Column("name").Returning("*").Exec(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) AddWebauthnCredential(ctx context.Context, userID uuid.UUID, credential *webauthn.Credential) error {
	newWebautnCredential := &WebauthnCredentials{
		UserID:          userID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       credential.Transport,
		Flags:           credential.Flags,
		Authenticator:   credential.Authenticator,
	}

	_, err := ur.DB.NewInsert().
		Model(newWebautnCredential).
		Column("user_id", "credential_id", "public_key", "attestation_type", "transport", "flags", "authenticator").
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
