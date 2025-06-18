package gateway_test

import (
	"net/http"
	"os"
	"testing"

	"anarchy.ttfm/8ball/blockchains/mock"
	"anarchy.ttfm/8ball/blockchains/monero"
	testsuite2 "anarchy.ttfm/8ball/blockchains/testsuite"
	"anarchy.ttfm/8ball/gateway/testsuite"
	"anarchy.ttfm/8ball/internal/walletrpc/rpc"
	"github.com/gabstv/httpdigest"
	"github.com/stretchr/testify/assert"
)

func Test_Integration(t *testing.T) {
	t.Run("Mock", func(t *testing.T) {
		w := mock.New()
		testsuite.Test(t, w, &testsuite2.MockGenerator{})
	})
	t.Run("Monero", func(t *testing.T) {
		assertions := assert.New(t)

		walletFilename := os.Getenv("MONERO_WALLET_FILENAME")
		assertions.NotEmpty(walletFilename, "MONERO_WALLET_FILENAME")
		walletPassword := os.Getenv("MONERO_WALLET_PASSWORD")
		assertions.NotEmpty(walletPassword, "MONERO_WALLET_PASSWORD")

		var config = rpc.Config{
			Address: "http://127.0.0.1:22222/json_rpc",

			Client: &http.Client{
				Transport: httpdigest.New("username", "password"), // Remove if no auth.
			},
		}

		client := rpc.New(config)

		w := monero.New(monero.Config{Client: client})
		testsuite.Test(t, w, &testsuite2.MoneroGenerator{})
	})
}
