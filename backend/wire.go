//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4"

	auth "github.com/shangsuru/passkey-demo/auth"
	"github.com/shangsuru/passkey-demo/users"
)

func NewServer() (*Server, error) {
	panic(wire.Build(
		wire.Struct(new(Server), "*"),
		echo.New,
		wire.Struct(new(auth.WebAuthnController), "*"),
		users.NewUserRepository,
		auth.NewWebAuthnAPI,
	))
}
