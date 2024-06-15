package main

import (
	"github.com/labstack/echo/v4"

	"github.com/shangsuru/passkey-demo/auth"
)

type Server struct {
	router             *echo.Echo
	webAuthnController auth.WebAuthnController
}

func (s *Server) Start() {
	s.registerEndpoints()
	s.router.Logger.Fatal(s.router.Start(":9044"))
}

func (s *Server) registerEndpoints() {
	s.router.Static("/static", "web")
	s.router.File("/", "web/register.html")
	s.router.File("/login", "web/login.html")

	s.router.GET("/register/begin/:username", s.webAuthnController.BeginRegistration())
	s.router.POST("/register/finish/:username", s.webAuthnController.FinishRegistration())
	s.router.GET("/login/begin/:username", s.webAuthnController.BeginLogin())
	s.router.POST("/login/finish/:username", s.webAuthnController.FinishLogin())
}
