package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/RogueTeam/8ball/decimal"
	"github.com/RogueTeam/8ball/gateway"
	"github.com/RogueTeam/8ball/internal/walletrpc/rpc"
	"github.com/RogueTeam/8ball/wallets/monero"
	"github.com/dgraph-io/badger/v4"
	"github.com/gabstv/httpdigest"
)

// Yaml configuration reference
type (
	Wallet struct {
		Filename    string  `yaml:"filename"`
		Password    string  `yaml:"password"`
		RpcUrl      string  `yaml:"rpc-url"`
		RpcUsername *string `yaml:"rpc-username,omitempty"`
		RpcPassword *string `yaml:"rpc-password,omitempty"`
	}
	Config struct {
		ProcessInterval    time.Duration   `yaml:"processInterval"`
		ListenAddress      string          `yaml:"listen-address"`
		DatabasePath       string          `yaml:"database-path"`
		MinAmount          decimal.Decimal `yaml:"min-amount"`
		MaxAmount          decimal.Decimal `yaml:"max-amount"`
		Timeout            time.Duration   `yaml:"receive-timeout"`
		FeePercentage      uint64          `yaml:"fee-percentage"`
		BeneficiaryAddress string          `yaml:"beneficiary-address"`
		Wallet             Wallet          `yaml:"wallet"`
	}
)

func (c *Config) Compile() (ctrl gateway.Controller, config gateway.Config, err error) {
	opt := badger.DefaultOptions(c.DatabasePath)

	var httpClient http.Client
	if c.Wallet.RpcUsername != nil && c.Wallet.RpcPassword != nil {
		httpClient.Transport = httpdigest.New(*c.Wallet.RpcUsername, *c.Wallet.RpcPassword)
	}

	moneroClient := rpc.New(rpc.Config{
		Url:    c.Wallet.RpcUrl,
		Client: &httpClient,
	})
	err = moneroClient.OpenWallet(context.TODO(), &rpc.OpenWalletRequest{
		Filename: c.Wallet.Filename,
		Password: c.Wallet.Password,
	})
	if err != nil {
		return ctrl, config, fmt.Errorf("failed to open wallet: %w", err)
	}

	config = gateway.Config{
		MinAmount:     c.MinAmount.ToUint64(),
		MaxAmount:     c.MaxAmount.ToUint64(),
		Timeout:       c.Timeout,
		FeePercentage: c.FeePercentage,
		Address:       c.BeneficiaryAddress,
		Wallet: monero.New(monero.Config{
			Accounts: true,
			Client:   moneroClient,
		}),
	}

	config.DB, err = badger.Open(opt)
	if err != nil {
		return ctrl, config, fmt.Errorf("failed to open database: %w", err)
	}

	ctrl = gateway.New(config)
	return ctrl, config, nil
}
