package main

import (
	"embed"
	"github.com/labstack/echo/v4"
	"github.com/shangsuru/passkey-demo/handler"
)

type Server struct {
	router             *echo.Echo
	webAuthnController handler.WebAuthnController
	passwordController handler.PasswordController
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
	s.router.POST("/register/begin", s.webAuthnController.BeginRegistration())
	s.router.POST("/register/finish", s.webAuthnController.FinishRegistration())
	s.router.POST("/login/begin", s.webAuthnController.BeginLogin())
	s.router.POST("/login/finish", s.webAuthnController.FinishLogin())
	s.router.POST("/discoverable_login/begin", s.webAuthnController.BeginDiscoverableLogin())
	s.router.POST("/discoverable_login/finish", s.webAuthnController.FinishDiscoverableLogin())
	s.router.POST("/register/password", s.passwordController.SignUp())
	s.router.POST("/login/password", s.passwordController.Login())
	s.router.POST("/logout", s.passwordController.Logout())

	s.router.FileFS("/", "index.html", distIndexHtml)
	s.router.FileFS("/sign-up", "index.html", distIndexHtml)
	s.router.FileFS("/home", "index.html", distIndexHtml)
	s.router.FileFS("/passkeys", "index.html", distIndexHtml)
	s.router.StaticFS("/", distDirFS)
}
