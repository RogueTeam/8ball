package controller

import (
	"errors"
	"time"

	"anarchy.ttfm.onion/gateway/blockchains"
	badger "github.com/dgraph-io/badger/v4"
)

var (
	ErrNoHandlerForCurrency = errors.New("no handler for currency")
	ErrPaymentNotFound      = errors.New("payment not found")
)

type Controller struct {
	db      *badger.DB
	fee     uint64
	timeout time.Duration
	wallets map[Currency]blockchains.Wallet
}

type Config struct {
	// Badger database to use
	DB *badger.DB
	// Percentage representing the CUT taken from the transaction
	Fee uint64
	// Default Timeout until payment in canceled
	Timeout time.Duration
	// Wallets to be used in the transactions
	Wallets map[Currency]blockchains.Wallet
}

func New(config *Config) (ctrl Controller) {
	ctrl.db = config.DB
	ctrl.fee = config.Fee
	ctrl.timeout = config.Timeout
	ctrl.wallets = config.Wallets

	return ctrl
}
