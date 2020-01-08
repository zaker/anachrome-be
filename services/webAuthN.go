package services

import (
	authn "github.com/duo-labs/webauthn/webauthn"
	"github.com/zaker/anachrome-be/stores"
)

type WebAuthN struct {
	UserStore stores.UserStore
}

func (w *WebAuthN) GetUser() (authn.User, error) {
	return w.UserStore.GetUser()
}
