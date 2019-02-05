package vault

import (
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/command/token"
	"github.com/readytalk/stim/pkg/log"

	"errors"
	"time"
)

type Vault struct {
	client      *api.Client
	tokenHelper token.InternalTokenHelper
	config      *Config
}

type Config struct {
	Noprompt bool
	Address  string
	Username string
	Timeout  time.Duration
	Log      log.Logger
}

func New(config *Config) (*Vault, error) {
	// Ensure that the Vault address is set
	if config.Address == "" {
		return nil, errors.New("Vault address not set")
	}

	v := &Vault{config: config}
	log.SetLogger(config.Log)

	if v.config.Timeout == 0 {
		v.config.Timeout = time.Second * 10 // No need to wait over a minite from default
	}

	// Configure new Vault Client
	apiConfig := api.DefaultConfig()
	apiConfig.Address = v.config.Address // Since we read the env we can override
	// apiConfig.HttpClient.Timeout = v.config.Timeout

	// Create our new API client
	var err error
	v.client, err = api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	// Ensure Vault is up and Healthy
	_, err = v.isVaultHealthy()
	if err != nil {
		return nil, err
	}

	// Run Login logic
	err = v.Login()
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (v *Vault) GetUser() string {
	return v.config.Username
}
