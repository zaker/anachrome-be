package services

import (
	"github.com/duo-labs/webauthn/protocol"
	authn "github.com/duo-labs/webauthn/webauthn"
	"github.com/zaker/anachrome-be/stores"
)

type WebAuthN struct {
	userStore stores.UserStore
	web       *authn.WebAuthn
}

type AuthNConfig struct {
	RPDisplayName string
	RPID          string
	RPOrigin      string
	RPIcon        string
}

func NewAuthN(store stores.UserStore, cfg *AuthNConfig) (*WebAuthN, error) {

	return &WebAuthN{userStore: store,
		web: &authn.WebAuthn{
			Config: &authn.Config{
				RPDisplayName: "Anachrome",
				RPID:          "localhost",
				RPOrigin:      "http://localhost:8080",
			}}}, nil
}
func (w *WebAuthN) GetUser() (authn.User, error) {
	return w.userStore.GetUser()
}

func (w *WebAuthN) BeginRegistration(user authn.User) (*protocol.CredentialCreation, *authn.SessionData, error) {
	return w.web.BeginRegistration(user)
}

func (w *WebAuthN) CreateCredential(user authn.User,
	sessionData authn.SessionData,
	parsedResponse *protocol.ParsedCredentialCreationData) (*authn.Credential, error) {

	return w.web.CreateCredential(user, sessionData, parsedResponse)
}

func (w *WebAuthN) BeginLogin(user authn.User) (*protocol.CredentialAssertion, *authn.SessionData, error) {
	return w.web.BeginLogin(user)
}

func (w *WebAuthN) ValidateLogin(user authn.User,
	sessionData authn.SessionData,
	parsedResponse *protocol.ParsedCredentialAssertionData) (*authn.Credential, error) {

	return w.web.ValidateLogin(user, sessionData, parsedResponse)
}
