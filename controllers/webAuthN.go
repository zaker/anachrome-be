package controllers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zaker/anachrome-be/services"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/labstack/echo/v4"
)

type Auth struct {
	Service *services.WebAuthN
}

func (a *Auth) BeginRegistration(c echo.Context) error {
	user, err := a.Service.GetUser() // Find or create the new user
	if err != nil {
		return err
	}
	options, sessionData, err := a.Service.BeginRegistration(user)
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
		// Domain:   c.Echo().Server.Addr,
		Domain:   "localhost",
		Path:     "/auth",
		Value:    base64.StdEncoding.EncodeToString(cValue),
		Expires:  time.Now().Add(12 * time.Hour),
		MaxAge:   int((1 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
	// return the options generated
	// options.publicKey contain our registration options
	return c.JSON(http.StatusOK, options)
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
	cValue, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return err
	}
	var sessionData webauthn.SessionData
	err = json.Unmarshal(cValue, &sessionData)
	if err != nil {
		return err
	}

	parsedResponse, err := protocol.ParseCredentialCreationResponse(c.Request())
	if err != nil {
		return err
	}
	cred, err := a.Service.CreateCredential(user, sessionData, parsedResponse)
	// Handle validation or input errors
	if err != nil {
		return err
	} else {
		c.Logger().Print("New credential", cred)
	}
	// If creation was successful, store the credential object

	return c.JSON(http.StatusOK, "ok") // Handle next steps

}

func (a *Auth) BeginLogin(c echo.Context) error {
	user, err := a.Service.GetUser() // Get the user
	if err != nil {
		return err
	}
	options, sessionData, err := a.Service.BeginLogin(user)
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
	return c.JSON(http.StatusOK, options) // return the options generated
	// options.publicKey contain our registration options

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
	if err != nil {
		return err
	}
	_, err = a.Service.ValidateLogin(user, sessionData, parsedResponse)
	// Handle validation or input errors
	if err != nil {
		return err
	}
	// If login was successful, handle next steps
	return c.JSON(http.StatusOK, "ok")

}
