package monero_test

import (
	"log"
	"net/http"
	"os"
	"testing"

	"anarchy.ttfm.onion/gateway/blockchains/monero"
	"anarchy.ttfm.onion/gateway/blockchains/monero/walletrpc/rpc"
	"anarchy.ttfm.onion/gateway/blockchains/testsuite"
	"github.com/gabstv/httpdigest"
	"github.com/stretchr/testify/assert"
)

var (
	walletFilename string
	walletPassword string
	accountName    string
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
	accountName = os.Getenv("MONERO_ACCOUNT_NAME")
	if accountName == "" {
		log.Fatal("MONERO_ACCOUNT_NAME not set")
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

func (g *dataGenerator) Destination() (addr string) {
	addr = os.Getenv("MONERO_DESTINATION")
	if addr == "" {
		log.Fatal("MONERO_DESTINATION not set")
	}
	return addr
}

func (g *dataGenerator) TransferAmount() (amount uint64) {
	return 1000000000
}

func Test_Monero(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		assertions := assert.New(t)

		client := newClient(t)
		var config = monero.Config{
			Client:   client,
			Account:  accountName,
			Filename: walletFilename,
			Password: walletPassword,
		}
		wallet, err := monero.New(config)
		assertions.Nil(err, "failed to create wallet manager")

		testsuite.Test(t, &wallet, &dataGenerator{})
	})
}
