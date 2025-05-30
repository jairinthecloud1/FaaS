package authenticator

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Authenticator is used to authenticate our users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

func envSanitized(envVar string) string {
	return strings.ReplaceAll(os.Getenv(envVar), "\n", "")
}

// New instantiates the *Authenticator.
func New() (*Authenticator, error) {

	log.Printf("Initializing Authenticator with AUTH0_DOMAIN=%s, AUTH0_CLIENT_ID=%s, AUTH0_CLIENT_SECRET=%s, AUTH0_CALLBACK_URL=%s",
		envSanitized("AUTH0_DOMAIN"),
		envSanitized("AUTH0_CLIENT_ID"),
		envSanitized("AUTH0_CLIENT_SECRET"),
		envSanitized("AUTH0_CALLBACK_URL"),
	)

	provider, err := oidc.NewProvider(
		context.Background(),

		"https://"+envSanitized("AUTH0_DOMAIN")+"/")

	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     envSanitized("AUTH0_CLIENT_ID"),
		ClientSecret: envSanitized("AUTH0_CLIENT_SECRET"),
		RedirectURL:  envSanitized("AUTH0_CALLBACK_URL"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}
