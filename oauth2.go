package main

import (
	"golang.org/x/oauth2"
)

type Config interface {
	setClientId(id string) error
	setClientSecret(secret string) error
	setEndpoint(endPoint oauth2.Endpoint) error
	setRedirectURL(redirectUrl string) error
	setScope(scopes []string) error
}
