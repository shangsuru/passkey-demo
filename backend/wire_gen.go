// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/labstack/echo/v4"
	"github.com/shangsuru/passkey-demo/auth"
	"github.com/shangsuru/passkey-demo/users"
)

// Injectors from wire.go:

func NewServer() (*Server, error) {
	echoEcho := echo.New()
	userRepository := users.NewUserRepository()
	webAuthn, err := auth.NewWebAuthnAPI()
	if err != nil {
		return nil, err
	}
	webAuthnController := auth.WebAuthnController{
		UserStore:   userRepository,
		WebAuthnAPI: webAuthn,
	}
	server := &Server{
		router:             echoEcho,
		webAuthnController: webAuthnController,
	}
	return server, nil
}
