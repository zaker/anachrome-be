package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/zaker/anachrome-be/services"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/labstack/echo/v4"
)

type Auth struct {
	RPID     string
	RPOrigin string
	Service  *services.WebAuthN
	web      *webauthn.WebAuthn
}

func (a *Auth) BeginRegistration(c echo.Context) error {
	user, err := a.Service.GetUser() // Find or create the new user
	if err != nil {
		return err
	}
	options, sessionData, err := a.web.BeginRegistration(user)
	if err != nil {
		return err
	}

	// handle errors if present
	// store the sessionData values
	cValue, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name: "registration-session",

		Value:   string(cValue),
		Expires: time.Now().Add(1 * time.Hour),
	}
	c.SetCookie(cookie)
	c.JSON(http.StatusOK, options) // return the options generated
	// options.publicKey contain our registration options
	return nil
}

func (a *Auth) FinishRegistration(c echo.Context) error {
	user, err := a.Service.GetUser() // Get the user
	if err != nil {
		return err
	}
	// Get the session data stored from the function above
	// using gorilla/sessions it could look like this
	cookie, err := c.Cookie("registration-session")
	if err != nil {
		return err
	}
	var sessionData webauthn.SessionData
	err = json.Unmarshal([]byte(cookie.Value), &sessionData)
	if err != nil {
		return err
	}

	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(c.Request().Body)
	_, err = a.web.CreateCredential(user, sessionData, parsedResponse)
	// Handle validation or input errors
	if err != nil {
		return err
	}
	// If creation was successful, store the credential object

	c.JSON(http.StatusOK, "Registration Success") // Handle next steps
	return nil
}

func (a *Auth) BeginLogin(c echo.Context) error {
	user, err := a.Service.GetUser() // Get the user
	if err != nil {
		return err
	}
	options, sessionData, err := a.web.BeginLogin(user)
	// handle errors if present
	if err != nil {
		return err
	}

	// store the sessionData values
	cValue, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name: "login-session",

		Value:   string(cValue),
		Expires: time.Now().Add(1 * time.Hour),
	}
	c.SetCookie(cookie)
	c.JSON(http.StatusOK, options) // return the options generated
	// options.publicKey contain our registration options
	return nil
}

func (a *Auth) FinishLogin(c echo.Context) error {
	user, err := a.Service.GetUser() // Get the user
	if err != nil {
		return err
	}
	// Get the session data stored from the function above
	// using gorilla/sessions it could look like this

	cookie, err := c.Cookie("login-session")
	if err != nil {
		return err
	}
	var sessionData webauthn.SessionData
	err = json.Unmarshal([]byte(cookie.Value), &sessionData)
	if err != nil {
		return err
	}
	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(c.Request().Body)
	_, err = a.web.ValidateLogin(user, sessionData, parsedResponse)
	// Handle validation or input errors
	if err != nil {
		return err
	}
	// If login was successful, handle next steps
	c.JSON(http.StatusOK, "Login Success")
	return nil
}
