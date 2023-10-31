package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

const (
	defaultKeyword = "gophkeeper"
	envKeywordName = "KEYWORD"
)

type ClientCfg struct {
	GKeeper      string `env:"GKS_ADDRESS" json:"gkeeper_address"`
	CertFilePath string `env:"CERTIFICATE" json:"agent_certificate"`
	KeyFilePath  string `env:"AGENT_KEY" json:"agent_key"`
	Keyword      []byte
}

func NewClientCfg() *ClientCfg {
	return &ClientCfg{}
}

func ReadEnvClientCfg(cfg *ClientCfg) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("an error occured when parse client config, err: %w", err)
	}

	key := os.Getenv(envKeywordName)
	if key == "" {
		key = defaultKeyword
	}
	cfg.Keyword = []byte(key)

	return nil
}
