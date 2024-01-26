package main

import (
	"net/http"

	"passkeys/controllers"

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

	// Routes
	r.Static("/static", "./views")
	r.LoadHTMLGlob("views/html/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})
	r.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})

	webAuthnController := controllers.NewWebAuthnController()
	r.GET("/register/begin/:username", webAuthnController.BeginRegistration)
	r.POST("/register/finish/:username", webAuthnController.FinishRegistration)
	r.GET("/login/begin/:username", webAuthnController.BeginLogin)
	r.POST("/login/finish/:username", webAuthnController.FinishLogin)

	_ = r.Run()
}
