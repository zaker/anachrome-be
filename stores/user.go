package stores

import (
	authn "github.com/duo-labs/webauthn/webauthn"
)

type User struct {
	id []byte
}

type UserStore interface {
	GetUser() (*User, error)
}

type UserFileStore struct {
}

func (user *User) WebAuthnID() []byte {
	return user.id
}

func (user *User) WebAuthnName() string {
	return "newUser"
}

func (user *User) WebAuthnDisplayName() string {
	return "New User"
}

func (user *User) WebAuthnIcon() string {
	return "https://pics.com/avatar.png"
}

func (user *User) WebAuthnCredentials() []authn.Credential {
	return []authn.Credential{}
}
