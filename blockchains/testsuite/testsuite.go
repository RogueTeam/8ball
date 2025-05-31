package testsuite

import (
	"testing"

	"anarchy.ttfm.onion/gateway/blockchains"
	"anarchy.ttfm.onion/gateway/blockchains/random"
	"github.com/stretchr/testify/assert"
)

type DataGeneratpr interface {
	// Generates a destination address to test sending funds
	Destination() (address string)

	// Returns the amount to send
	Send() (funds uint64)
}

func Test(t *testing.T, w blockchains.Wallet) {
	t.Run("Balance", func(t *testing.T) {
		assertions := assert.New(t)

		balance, err := w.Balance()
		assertions.Nil(err, "failed to retrieve wallet balance")

		t.Log(balance)
	})
	t.Run("NewAddress", func(t *testing.T) {
		assertions := assert.New(t)

		addr, err := w.NewAddress(blockchains.NewAddressRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
		assertions.Nil(err, "failed to retrieve wallet balance")

		t.Log(addr)
	})
}
