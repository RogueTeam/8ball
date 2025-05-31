package testsuite

import (
	"testing"

	"anarchy.ttfm.onion/gateway/blockchains"
	"anarchy.ttfm.onion/gateway/random"
	"github.com/stretchr/testify/assert"
)

type DataGeneratpr interface {
	// Generates a destination address to test sending funds
	Destination() (address string)

	// Returns the amount to send
	TransferAmount() (funds uint64)
}

func Test(t *testing.T, w blockchains.Wallet, gen DataGeneratpr) {
	t.Run("Balance", func(t *testing.T) {
		assertions := assert.New(t)

		balance, err := w.Balance()
		assertions.Nil(err, "failed to retrieve wallet balance")

		t.Log(balance)
	})
	t.Run("NewAddress", func(t *testing.T) {
		assertions := assert.New(t)

		addr, err := w.NewAddress(blockchains.NewAddressRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
		assertions.Nil(err, "failed to create new address")

		t.Log(addr)
	})
	t.Run("Transfer", func(t *testing.T) {
		t.Run("To subaddress", func(t *testing.T) {
			assertions := assert.New(t)

			dst, err := w.NewAddress(blockchains.NewAddressRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new address")

			transfer, err := w.Transfer(blockchains.TransferRequest{
				Source:      dst.AccountAddress,
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
