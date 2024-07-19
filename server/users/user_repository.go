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

func (ur *UserRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := ur.DB.NewSelect().
		Model(&user).
		Relation("WebauthnCredentials").
		Column("*").
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) FindUserByID(ctx context.Context, rawUserID []byte) (*User, error) {
	userID, err := uuid.FromBytes(rawUserID)
	if err != nil {
		return nil, err
	}

	var user User
	err = ur.DB.NewSelect().
		Model(&user).
		Relation("WebauthnCredentials").
		Column("*").
		Where("id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) CreateUser(ctx context.Context, email string, passwordHash string) (*User, error) {
	user := &User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	_, err := ur.DB.NewInsert().Model(user).Column("email", "password_hash").Returning("*").Exec(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) DeleteUser(ctx context.Context, user *User) error {
	_, err := ur.DB.NewDelete().Model(user).WherePK().Exec(ctx)
	return err
}

func (ur *UserRepository) AddWebauthnCredential(ctx context.Context, userID uuid.UUID, credential *webauthn.Credential) error {
	newWebauthnCredential := &WebauthnCredentials{
		UserID:          userID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       credential.Transport,
		Flags:           credential.Flags,
		Authenticator:   credential.Authenticator,
	}

	_, err := ur.DB.NewInsert().
		Model(newWebauthnCredential).
		Column("user_id", "credential_id", "public_key", "attestation_type", "transport", "flags", "authenticator").
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
