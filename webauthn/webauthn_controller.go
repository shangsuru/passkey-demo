package webauthn

import (
	"log"
	"net/http"

	"github.com/shangsuru/passkey-demo/users"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
)

type WebAuthnController interface {
	BeginRegistration(c *gin.Context)
	FinishRegistration(c *gin.Context)
	BeginLogin(c *gin.Context)
	FinishLogin(c *gin.Context)
}

type webAuthnController struct {
	userStore users.UserRepository
	webAuthn *webauthn.WebAuthn
}

func NewWebAuthnController() WebAuthnController {
	webAuthn, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "PassKey Demo",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	})

	if err != nil {
		log.Fatal("failed to create WebAuthn from config:", err)
	}
	return webAuthnController {
		userStore: users.NewUserRepository(),
		webAuthn: webAuthn,
	}
}

func (wc webAuthnController) BeginRegistration(c *gin.Context) {
	username := c.Param("username")

	// get user
	user, err := wc.userStore.GetUser(username)
	// user doesn't exist, create new user
	if err != nil {
		wc.userStore.PutUser(username)
	}

	// generate PublicKeyCredentialCreationOptions, session data
	options, sessionData, err := wc.webAuthn.BeginRegistration(user)
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

func (wc webAuthnController) FinishRegistration(c *gin.Context) {
	username := c.Param("username")
	// get user
	user, err := wc.userStore.GetUser(username)
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

	credential, err := wc.webAuthn.FinishRegistration(user, *sessionData, c.Request)
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

func (wc webAuthnController) BeginLogin(c *gin.Context) {
	username := c.Param("username")
	user, err := wc.userStore.GetUser(username)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// generate PublicKeyCredentialRequestOptions, session data
	options, sessionData, err := wc.webAuthn.BeginLogin(user)
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

func (wc webAuthnController) FinishLogin(c *gin.Context) {
	username := c.Param("username")
	user, err := wc.userStore.GetUser(username)
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
	_, err = wc.webAuthn.FinishLogin(user, *sessionData, c.Request)
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