package monero_test

import (
	"log"
	"net/http"
	"os"
	"testing"

	"anarchy.ttfm/8ball/blockchains/monero"
	"anarchy.ttfm/8ball/blockchains/testsuite"
	"anarchy.ttfm/8ball/internal/walletrpc/rpc"
	"anarchy.ttfm/8ball/utils"
	"github.com/gabstv/httpdigest"
	"github.com/stretchr/testify/assert"
)

var (
	walletFilename string
	walletPassword string
)

func init() {
	walletFilename = os.Getenv("MONERO_WALLET_FILENAME")
	if walletFilename == "" {
		log.Fatal("MONERO_WALLET_FILENAME not set")
	}
	walletPassword = os.Getenv("MONERO_WALLET_PASSWORD")
	if walletPassword == "" {
		log.Fatal("MONERO_WALLET_PASSWORD not set")
	}
}

func newClient(t *testing.T) (client *rpc.Client) {
	assertions := assert.New(t)

	var config = rpc.Config{
		Address: "http://127.0.0.1:22222/json_rpc",

		Client: &http.Client{
			Transport: httpdigest.New("username", "password"), // Remove if no auth.
		},
	}
	client = rpc.New(config)
	assertions.NotNil(client, "failed to create client")

	return client
}

type dataGenerator struct {
}

func (g *dataGenerator) TransferAmount() (amount uint64) {
	return 100000000
}

func Test_Monero(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		assertions := assert.New(t)

		client := newClient(t)

		ctx, cancel := utils.NewContext()
		defer cancel()
		err := client.OpenWallet(ctx, &rpc.OpenWalletRequest{
			Filename: walletFilename,
			Password: walletPassword,
		})
		assertions.Nil(err, "failed to open wallet")

		var config = monero.Config{
			Client: client,
		}
		wallet := monero.New(config)

		testsuite.Test(t, &wallet, &dataGenerator{})
	})
}
