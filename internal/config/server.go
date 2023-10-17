package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

type ServerCfg struct {
	Addr             string `env:"ADDRESS" json:"address"`
	PrivateCryptoKey string `env:"CRYPTO_KEY" json:"private_key"`
	CertFilePath     string `env:"CERTIFICATE" json:"certificate"`
}

func NewConfig() *ServerCfg {
	return &ServerCfg{}
}

func ReadEnvConfig(cfg *ServerCfg) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("an occured error when parse server config err: %w", err)
	}
	return nil
}
