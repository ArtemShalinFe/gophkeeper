package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env"
)

const (
	defaultAddr = "localhost:6085"
	defaultDSN  = "postgres://gkeeper:gkeeper@localhost:5432/gophkeeper?sslmode=disable"
)

// ServerCfg - An object that implements the server configuration.
type ServerCfg struct {
	// Addr - The address in the format "host:port" on which the server will be started. Example: localhost:9080.
	Addr string `env:"GKS_ADDRESS" json:"gkeeper_address"`
	// DSN - The name of the data source, the data structures used to describe the connection to the data source.
	DSN string `env:"DATABASE_DSN" json:"dsn"`
	// PrivateCryptoKey - The path to the private key to ensure the operation of TLS.
	PrivateCryptoKey string `env:"CRYPTO_KEY" json:"server_private_key"`
	// CertFilePath - The path to the certificate to ensure TLS operation.
	CertFilePath string `env:"CERTIFICATE" json:"server_certificate"`
}

// NewServerCfg - Object Constructor.
func NewServerCfg() *ServerCfg {
	return &ServerCfg{}
}

// ReadEnvServerCfg - Reads environment variables and
// stores them in an object that implements the server configuration.
func ReadEnvServerCfg(cfg *ServerCfg) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("an error occured when parse server config err: %w", err)
	}
	if strings.TrimSpace(cfg.Addr) == "" {
		cfg.Addr = defaultAddr
	}
	if strings.TrimSpace(cfg.DSN) == "" {
		cfg.DSN = defaultDSN
	}
	return nil
}
