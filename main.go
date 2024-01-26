package main

import (
	"github.com/shangsuru/passkey-demo/routes"
	"github.com/shangsuru/passkey-demo/webauthn"

	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Database Setup
	db, err := gorm.Open(sqlite.Open("db/test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Session
	sessionStore := gormsessions.NewStore(db, true, []byte("secret"))

	r := gin.Default()
	r.Use(sessions.Sessions("mySession", sessionStore))
 
	// Route Setup
	routes.SetupFrontendRoutes(r)
	webAuthnController := webauthn.NewWebAuthnController()
	webAuthnRouteController := routes.NewWebAuthnRouteController(webAuthnController)
	webAuthnRouteController.WebAuthnRoutes(r)
	_ = r.Run()
}
