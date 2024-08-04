// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/labstack/echo/v4"
	"github.com/shangsuru/passkey-demo/db"
	"github.com/shangsuru/passkey-demo/handler"
	"github.com/shangsuru/passkey-demo/repository"
)

// Injectors from wire.go:

func NewServer() (*Server, error) {
	echoEcho := echo.New()
	bunDB := db.GetDB()
	userRepository := repository.UserRepository{
		DB: bunDB,
	}
	webAuthn, err := handler.NewWebAuthnAPI()
	if err != nil {
		return nil, err
	}
	webAuthnController := handler.WebAuthnController{
		UserRepository: userRepository,
		WebAuthnAPI:    webAuthn,
	}
	passwordController := handler.PasswordController{
		UserRepository: userRepository,
	}
	server := &Server{
		router:             echoEcho,
		webAuthnController: webAuthnController,
		passwordController: passwordController,
	}
	return server, nil
}
