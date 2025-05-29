package testsuite

import (
	"testing"

	"anarchy.ttfm.onion/gateway/blockchains"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T, w blockchains.Wallet) {
	t.Run("Balance", func(t *testing.T) {
		assertions := assert.New(t)

		balance, err := w.Balance()
		assertions.Nil(err, "failed to retrieve wallet balance")

		t.Log(balance)
	})
}
