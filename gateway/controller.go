package gateway

import (
	"errors"
	"time"

	"anarchy.ttfm/8ball/wallets"
	badger "github.com/dgraph-io/badger/v4"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

type Controller struct {
	db          *badger.DB
	timeout     time.Duration
	beneficiary string
	wallet      wallets.Wallet
}

type Config struct {
	// Badger database to use
	DB *badger.DB
	// Default Timeout until payment in canceled
	Timeout time.Duration
	// Beneficiaries address. To these address the money is going to be payed
	// This is the address of the one running the gateway
	Beneficiary string
	// Wallets to be used for managing transactions
	Wallet wallets.Wallet
}

func New(config Config) (ctrl Controller) {
	ctrl.db = config.DB
	ctrl.timeout = config.Timeout
	ctrl.beneficiary = config.Beneficiary
	ctrl.wallet = config.Wallet

	return ctrl
}
