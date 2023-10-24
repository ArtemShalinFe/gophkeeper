package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

type ServerCfg struct {
	Addr             string `env:"GKS_ADDRESS" json:"gkeeper_address"`
	DSN              string `env:"DATABASE_DSN" json:"dsn"`
	PrivateCryptoKey string `env:"CRYPTO_KEY" json:"server_private_key"`
	CertFilePath     string `env:"CERTIFICATE" json:"server_certificate"`
}

func NewServerCfg() *ServerCfg {
	return &ServerCfg{}
}

func ReadEnvServerCfg(cfg *ServerCfg) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("an error occured when parse server config err: %w", err)
	}
	return nil
}
