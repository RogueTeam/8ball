package controller

import (
	"errors"
	"time"

	"anarchy.ttfm.onion/gateway/blockchains"
	badger "github.com/dgraph-io/badger/v4"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

type Controller struct {
	db          *badger.DB
	fee         uint64
	timeout     time.Duration
	beneficiary string
	wallet      blockchains.Wallet
}

type Config struct {
	// Badger database to use
	DB *badger.DB
	// Percentage representing the CUT taken from the transaction
	Fee uint64
	// Default Timeout until payment in canceled
	Timeout time.Duration
	// Beneficiaries address per currency. To these address the money is going to be payed
	Beneficiary string
	// Wallets to be used in the transactions
	Wallet blockchains.Wallet
}

func New(config Config) (ctrl Controller) {
	ctrl.db = config.DB
	ctrl.fee = config.Fee
	ctrl.timeout = config.Timeout
	ctrl.beneficiary = config.Beneficiary
	ctrl.wallet = config.Wallet

	return ctrl
}
