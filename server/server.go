package main

import (
	"embed"

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

var (
	//go:embed all:dist
	dist embed.FS
	//go:embed dist/index.html
	indexHTML embed.FS

	distDirFS     = echo.MustSubFS(dist, "dist")
	distIndexHtml = echo.MustSubFS(indexHTML, "dist")
)

func (s *Server) registerEndpoints() {
	s.router.FileFS("/", "index.html", distIndexHtml)
	s.router.FileFS("/sign-up", "index.html", distIndexHtml)
	s.router.StaticFS("/", distDirFS)

	s.router.GET("/register/begin/:username", s.webAuthnController.BeginRegistration())
	s.router.POST("/register/finish/:username", s.webAuthnController.FinishRegistration())
	s.router.GET("/login/begin/:username", s.webAuthnController.BeginLogin())
	s.router.POST("/login/finish/:username", s.webAuthnController.FinishLogin())
}
