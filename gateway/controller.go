package gateway

import (
	"errors"
	"time"

	"github.com/RogueTeam/8ball/wallets"
	badger "github.com/dgraph-io/badger/v4"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

type Controller struct {
	minAmount     uint64
	maxAmount     uint64
	db            *badger.DB
	timeout       time.Duration
	feePercentage uint64
	address       string
	wallet        wallets.Wallet
}

type Config struct {
	// Minimum amount accepted
	MinAmount uint64
	// Maximum amount accepted
	MaxAmount uint64
	// Badger database to use
	DB *badger.DB
	// Default Timeout until payment in canceled
	Timeout time.Duration
	// Percentage from 0 to 100 to be discounted from the payments and payed the gateway
	// manager
	FeePercentage uint64
	// To these address the fees are going to be payed
	// This is the address of the one running the gateway
	Address string
	// Wallets to be used for managing transactions
	Wallet wallets.Wallet
}

func New(config Config) (ctrl Controller) {
	ctrl.minAmount = config.MinAmount
	ctrl.maxAmount = config.MaxAmount
	ctrl.db = config.DB
	ctrl.timeout = config.Timeout
	ctrl.feePercentage = config.FeePercentage
	ctrl.address = config.Address
	ctrl.wallet = config.Wallet

	return ctrl
}
