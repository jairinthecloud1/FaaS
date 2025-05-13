package config

type OIDCConfig struct {
	Issuer   string
	ClientID string
}

var OIDC = OIDCConfig{
	Issuer:   "https://your-identity-provider.com", // Replace with your actual OIDC provider URL
	ClientID: "your-client-id",                     // Replace with your actual client ID
}
