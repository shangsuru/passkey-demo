package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/shangsuru/passkey-demo/users"
)

type WebAuthnController struct {
	UserStore   users.UserRepository
	WebAuthnAPI *webauthn.WebAuthn
}

func (wc WebAuthnController) BeginRegistration(c *gin.Context) {
	username := c.Param("username")

	// get user
	user, err := wc.UserStore.GetUser(username)
	// user doesn't exist, create new user
	if err != nil {
		wc.UserStore.PutUser(username)
	}

	// generate PublicKeyCredentialCreationOptions, session data
	options, sessionData, err := wc.WebAuthnAPI.BeginRegistration(user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := storeSessionData(c, username, sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, options)
}

func (wc WebAuthnController) FinishRegistration(c *gin.Context) {
	username := c.Param("username")
	// get user
	user, err := wc.UserStore.GetUser(username)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	sessionData, err := loadSessionData(c, username)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	credential, err := wc.WebAuthnAPI.FinishRegistration(user, *sessionData, c.Request)
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

func (wc WebAuthnController) BeginLogin(c *gin.Context) {
	username := c.Param("username")
	user, err := wc.UserStore.GetUser(username)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// generate PublicKeyCredentialRequestOptions, session data
	options, sessionData, err := wc.WebAuthnAPI.BeginLogin(user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := storeSessionData(c, username, sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, options)
}

func (wc WebAuthnController) FinishLogin(c *gin.Context) {
	username := c.Param("username")
	user, err := wc.UserStore.GetUser(username)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	sessionData, err := loadSessionData(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	// in an actual implementation we should perform additional
	// checks on the returned 'credential'
	_, err = wc.WebAuthnAPI.FinishLogin(user, *sessionData, c.Request)
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
