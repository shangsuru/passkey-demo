package webauthn

import (
	"encoding/json"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
)

func loadSessionData(c *gin.Context, username string) (*webauthn.SessionData, error) {
	session := sessions.Default(c)

	// Get session data bytes from the session
	bytes := session.Get(username).([]byte)

	// Unmarshal session data from JSON
	var sessionData webauthn.SessionData
	if err := json.Unmarshal(bytes, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %v", err)
	}

	// Return session data and nil error if everything is successful
	return &sessionData, nil
}

func storeSessionData(c *gin.Context, username string, sessionData *webauthn.SessionData) error {
	session := sessions.Default(c)

	// Marhsall session data to JSON
	bytes, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marhsall session data: %v", err)	
	}

	// Set session data for the user
	session.Set(username, bytes)

	// Save the session
	if err = session.Save(); err != nil {
		return fmt.Errorf("failed to save session: %v", err)
	}

	// Return nil if everything is successful
	return nil
}
