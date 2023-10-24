package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

type AgentCfg struct {
	GKeeper      string `env:"GKS_ADDRESS" json:"gkeeper_address"`
	Addr         string `env:"GKA_ADDRESS" json:"gagent_address"`
	CertFilePath string `env:"CERTIFICATE" json:"agent_certificate"`
}

func NewAgentCfg() *AgentCfg {
	return &AgentCfg{}
}

func ReadEnvAgentCfg(cfg *AgentCfg) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("an error occured when parse agent config err: %w", err)
	}
	return nil
}
