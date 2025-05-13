package auth

import (
	"context"
	"fmt"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jairinthecloud1/FaaS/internal/config"
)

var (
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	providerOnce sync.Once
	initError    error
)

// InitializeOIDC initializes the OIDC provider and verifier
func InitializeOIDC(ctx context.Context) error {
	providerOnce.Do(func() {
		var err error
		provider, err = oidc.NewProvider(ctx, config.OIDC.Issuer)
		if err != nil {
			initError = fmt.Errorf("failed to initialize OIDC provider: %w", err)
			return
		}

		verifier = provider.Verifier(&oidc.Config{
			ClientID: config.OIDC.ClientID,
		})
	})

	return initError
}

// VerifyToken verifies an OIDC token and returns the parsed claims
func VerifyToken(ctx context.Context, tokenString string) (*oidc.IDToken, error) {
	if provider == nil || verifier == nil {
		return nil, fmt.Errorf("OIDC not initialized")
	}

	// Remove "Bearer " prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Verify the token
	return verifier.Verify(ctx, tokenString)
}
