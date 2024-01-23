package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var webAuthn *webauthn.WebAuthn

var sessionStore gormsessions.Store
var userStore *UserDB

func BeginRegistration(c *gin.Context) {
	username := c.Param("username")

	// get user
	user, err := userStore.GetUser(username)
	// user doesn't exist, create new user
	if err != nil {
		displayName := strings.Split(username, "@")[0]
		user = NewUser(username, displayName)
		userStore.PutUser(user)
	}

	// generate PublicKeyCredentialCreationOptions, session data
	options, sessionData, err := webAuthn.BeginRegistration(user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := StoreSessionData(c, username, sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, options)
}

func FinishRegistration(c *gin.Context) {
	username := c.Param("username")
	// get user
	user, err := userStore.GetUser(username)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	sessionData, err := LoadSessionData(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	credential, err := webAuthn.FinishRegistration(user, *sessionData, c.Request)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user.AddCredential(*credential)

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration Success",
	})
}

func BeginLogin(c *gin.Context) {
	username := c.Param("username")
	user, err := userStore.GetUser(username)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// generate PublicKeyCredentialRequestOptions, session data
	options, sessionData, err := webAuthn.BeginLogin(user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := StoreSessionData(c, username, sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, options)
}

func FinishLogin(c *gin.Context) {
	username := c.Param("username")
	user, err := userStore.GetUser(username)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	sessionData, err := LoadSessionData(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	// in an actual implementation we should perform additional
	// checks on the returned 'credential'
	_, err = webAuthn.FinishLogin(user, *sessionData, c.Request)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// handle successful login
	c.JSON(http.StatusOK, gin.H{
		"message": "Login Success",
	})
}

func main() {
	var err error
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "PassKey Demo",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	})

	if err != nil {
		log.Fatal("failed to create WebAuthn from config:", err)
	}

	// Database
	db, err := gorm.Open(sqlite.Open("db/test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Users
	// TODO: use database
	userStore = DB()

	// Session
	sessionStore = gormsessions.NewStore(db, true, []byte("secret"))

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

	r.GET("/register/begin/:username", BeginRegistration)
	r.POST("/register/finish/:username", FinishRegistration)
	r.GET("/login/begin/:username", BeginLogin)
	r.POST("/login/finish/:username", FinishLogin)

	_ = r.Run() // listen and serve on 0.0.0.0:8080
}
