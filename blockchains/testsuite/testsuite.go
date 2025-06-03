package testsuite

import (
	"testing"

	"anarchy.ttfm.onion/gateway/blockchains"
	"anarchy.ttfm.onion/gateway/random"
	"github.com/stretchr/testify/assert"
)

type DataGeneratpr interface {
	// Returns the amount to send
	TransferAmount() (funds uint64)
}

func Test(t *testing.T, w blockchains.Wallet, gen DataGeneratpr) {
	t.Run("Balance", func(t *testing.T) {
		assertions := assert.New(t)

		balance, err := w.Account(blockchains.AccountRequest{Index: 0})
		assertions.Nil(err, "failed to retrieve wallet balance")

		t.Log(balance)
	})
	t.Run("NewAccount", func(t *testing.T) {
		assertions := assert.New(t)

		account, err := w.NewAccount(blockchains.NewAccountRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
		assertions.Nil(err, "failed to create new acciybt")
		t.Log(account)

		balance, err := w.Account(blockchains.AccountRequest{Index: account.Index})
		assertions.Nil(err, "failed to get account")
		t.Log(balance)
	})
	t.Run("Transfer", func(t *testing.T) {
		t.Run("To Internal Account", func(t *testing.T) {
			assertions := assert.New(t)

			dst, err := w.NewAccount(blockchains.NewAccountRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new account")

			transfer, err := w.Transfer(blockchains.TransferRequest{
				SourceIndex: 0,
				Destination: dst.Address,
				Amount:      gen.TransferAmount(),
				Priority:    blockchains.PriorityLow,
				UnlockTime:  0,
			})
			assertions.Nil(err, "failed to transfer funds")
			t.Log(transfer)
		})
	})
}
