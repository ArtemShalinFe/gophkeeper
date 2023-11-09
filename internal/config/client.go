package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env"
)

// ClientCfg - An object that implements the application configuration.
type ClientCfg struct {
	// GKeeper - The address of the server running the gophkeeper service.
	GKeeper string `env:"GKS_ADDRESS" json:"gkeeper_address"`
	// CertFilePath - The path to the certificate file.
	CertFilePath string `env:"CERTIFICATE" json:"agent_certificate"`
}

// NewClientCfg - Object Constructor.
func NewClientCfg() *ClientCfg {
	return &ClientCfg{}
}

// ReadEnvClientCfg - Reads environment variables and
// stores them in an object that implements the application configuration.
func ReadEnvClientCfg(cfg *ClientCfg) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("an error occured when parse client config, err: %w", err)
	}
	if strings.TrimSpace(cfg.GKeeper) == "" {
		cfg.GKeeper = defaultAddr
	}
	return nil
}
