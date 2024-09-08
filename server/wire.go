//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"github.com/shangsuru/passkey-demo/repository"

	"github.com/shangsuru/passkey-demo/db"
	"github.com/shangsuru/passkey-demo/handler"
)

// After updating this, run `go generate` to update wire_gen.go
func NewServer() (*Server, error) {
	panic(wire.Build(
		wire.Struct(new(Server), "*"),
		echo.New,
		wire.Struct(new(handler.WebAuthnController), "*"),
		wire.Struct(new(handler.PasswordController), "*"),
		wire.Struct(new(repository.UserRepository), "*"),
		db.GetDB,
		handler.NewSessionRepository,
		handler.NewWebAuthnAPI,
	))
}
