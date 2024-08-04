package repository

import (
	"context"
	"github.com/shangsuru/passkey-demo/model"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserRepository struct {
	DB *bun.DB
}

func (ur *UserRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
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

func (ur *UserRepository) FindUserByID(ctx context.Context, rawUserID []byte) (*model.User, error) {
	userID, err := uuid.FromBytes(rawUserID)
	if err != nil {
		return nil, err
	}

	var user model.User
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

func (ur *UserRepository) CreateUser(ctx context.Context, email string, passwordHash string) (*model.User, error) {
	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
	}

	_, err := ur.DB.NewInsert().Model(user).Column("id", "email", "password_hash").Returning("*").Exec(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) DeleteUser(ctx context.Context, user *model.User) error {
	_, err := ur.DB.NewDelete().Model(user).WherePK().Exec(ctx)
	return err
}

func (ur *UserRepository) AddWebauthnCredential(ctx context.Context, userID uuid.UUID, credential *webauthn.Credential) error {
	newWebauthnCredential := &model.WebauthnCredentials{
		ID:              uuid.New(),
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
		Column("id", "user_id", "credential_id", "public_key", "attestation_type", "transport", "flags", "authenticator").
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (ur *UserRepository) FindUserIDByCredentialID(ctx context.Context, id []byte) (*uuid.UUID, error) {
	var credential model.WebauthnCredentials
	err := ur.DB.NewSelect().
		Model(&credential).
		Column("user_id").
		Where("credential_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &credential.UserID, nil
}
