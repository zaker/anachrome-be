package stores

import (
	authn "github.com/duo-labs/webauthn/webauthn"
)

type User struct {
	id          []byte
	name        string
	DisplayName string
	Icon        string
	Creds       []authn.Credential
}

type UserStore interface {
	GetUser() (*User, error)
}

type UserFileStore struct {
}

func (us *UserFileStore) GetUser() (*User, error) {
	return &User{
		id:          []byte("1234"),
		name:        "zaker",
		DisplayName: "Zaker",
		Creds:       []authn.Credential{},
	}, nil
}

func (user *User) WebAuthnID() []byte {
	return user.id
}

func (user *User) WebAuthnName() string {
	return user.name
}

func (user *User) WebAuthnDisplayName() string {
	return user.DisplayName
}

func (user *User) WebAuthnIcon() string {
	return user.Icon
}

func (user *User) WebAuthnCredentials() []authn.Credential {
	return user.Creds
}
