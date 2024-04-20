//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	auth "github.com/shangsuru/passkey-demo/auth"
	"github.com/shangsuru/passkey-demo/users"
)

func NewServer() (*Server, error) {
	panic(wire.Build(
		wire.Struct(new(Server), "*"),
		gin.Default,
		wire.Struct(new(auth.WebAuthnController), "*"),
		users.NewUserRepository,
		auth.NewWebAuthnAPI,
	))
}
