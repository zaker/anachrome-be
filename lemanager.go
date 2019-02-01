package main

import (
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func ConfigManager(man *autocert.Manager) error {
	man.Client.DirectoryURL = "https://acme-staging-v02.api.letsencrypt.org/directory"

	man.Prompt = acme.AcceptTOS
	return nil
}
