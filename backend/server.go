package main

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/shangsuru/passkey-demo/auth"
)

type Server struct {
	router             *gin.Engine
	webAuthnController auth.WebAuthnController
}

func (s *Server) Start() {
	// Database Setup
	db, err := gorm.Open(sqlite.Open("db/test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Session
	sessionStore := gormsessions.NewStore(db, true, []byte("secret"))

	s.router.Use(sessions.Sessions("mySession", sessionStore))

	// Route Setup
	s.registerEndpoints()
	_ = s.router.Run()
}

func (s *Server) registerEndpoints() {
	s.router.Static("/static", "../frontend")
	s.router.LoadHTMLGlob("../frontend/html/*")
	s.router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})
	s.router.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})

	s.router.GET("/register/begin/:username", s.webAuthnController.BeginRegistration)
	s.router.POST("/register/finish/:username", s.webAuthnController.FinishRegistration)
	s.router.GET("/login/begin/:username", s.webAuthnController.BeginLogin)
	s.router.POST("/login/finish/:username", s.webAuthnController.FinishLogin)
}
