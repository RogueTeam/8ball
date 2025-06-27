package gateway

import (
	"fmt"
	"net/http"
	"time"

	"anarchy.ttfm/8ball/decimal"
	"anarchy.ttfm/8ball/gateway"
	"anarchy.ttfm/8ball/internal/walletrpc/rpc"
	"anarchy.ttfm/8ball/wallets/monero"
	"github.com/dgraph-io/badger/v4"
	"github.com/gabstv/httpdigest"
)

// Yaml configuration reference
type (
	WalletRPC struct {
		Url      string  `yaml:"url"`
		Username *string `yaml:"username,omitempty"`
		Password *string `yaml:"password,omitempty"`
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
		WalletRpc          WalletRPC       `yaml:"wallet-rpc"`
	}
)

func (c *Config) Compile() (ctrl gateway.Controller, config gateway.Config, err error) {
	opt := badger.DefaultOptions(c.DatabasePath)

	var httpClient http.Client
	if c.WalletRpc.Username != nil && c.WalletRpc.Password != nil {
		httpClient.Transport = httpdigest.New(*c.WalletRpc.Username, *c.WalletRpc.Password)
	}

	config = gateway.Config{
		MinAmount:     c.MinAmount.ToUint64(),
		MaxAmount:     c.MaxAmount.ToUint64(),
		Timeout:       c.Timeout,
		FeePercentage: c.FeePercentage,
		Address:       c.BeneficiaryAddress,
		Wallet: monero.New(monero.Config{
			Accounts: true,
			Client: rpc.New(rpc.Config{
				Url:    c.WalletRpc.Url,
				Client: &httpClient,
			}),
		}),
	}

	config.DB, err = badger.Open(opt)
	if err != nil {
		return ctrl, config, fmt.Errorf("failed to open database: %w", err)
	}

	ctrl = gateway.New(config)
	return ctrl, config, nil
}
