package main

import (
	"encoding/json"
	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
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

	// store session data as marshaled JSON
	session := sessions.Default(c)
	bytes, err := json.Marshal(sessionData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	session.Set("registration", bytes)
	err = session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
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

	// load the session data
	session := sessions.Default(c)
	bytes := session.Get("registration").([]byte)
	var sessionData webauthn.SessionData
	err = json.Unmarshal(bytes, &sessionData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	credential, err := webAuthn.FinishRegistration(user, sessionData, c.Request)
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
	r.LoadHTMLGlob("views/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.GET("/register/begin/:username", BeginRegistration)
	r.POST("/register/finish/:username", FinishRegistration)

	_ = r.Run() // listen and serve on 0.0.0.0:8080
}
