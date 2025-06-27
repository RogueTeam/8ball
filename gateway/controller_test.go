package gateway_test

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"anarchy.ttfm/8ball/gateway/testsuite"
	"anarchy.ttfm/8ball/internal/walletrpc/rpc"
	"anarchy.ttfm/8ball/utils"
	"anarchy.ttfm/8ball/wallets/mock"
	"anarchy.ttfm/8ball/wallets/monero"
	testsuite2 "anarchy.ttfm/8ball/wallets/testsuite"
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

func newMoneroClient(t *testing.T) (client *rpc.Client) {
	assertions := assert.New(t)

	var config = rpc.Config{
		Url: "http://127.0.0.1:22222/json_rpc",

		Client: &http.Client{
			Transport: httpdigest.New("username", "password"), // Remove if no auth.
		},
	}
	client = rpc.New(config)
	assertions.NotNil(client, "failed to create client")

	return client
}

func Test_Integration(t *testing.T) {
	t.Parallel()
	t.Run("Mock", func(t *testing.T) {
		t.Parallel()

		var configs = []mock.Config{
			{FundsDelta: 5 * time.Second},
			// {FundsDelta: 5 * time.Second, ZeroOnTransfer: true},
		}
		for _, config := range configs {
			cc, _ := json.Marshal(config)
			t.Run(string(cc), func(t *testing.T) {
				t.Parallel()

				w := mock.New(config)
				testsuite.Test(t, 1, w, &testsuite2.MockGenerator{})
			})
		}
	})
	t.Run("Monero", func(t *testing.T) {
		t.Parallel()

		assertions := assert.New(t)

		client := newMoneroClient(t)

		ctx, cancel := utils.NewContext()
		defer cancel()
		err := client.OpenWallet(ctx, &rpc.OpenWalletRequest{
			Filename: walletFilename,
			Password: walletPassword,
		})
		assertions.Nil(err, "failed to open wallet")

		var config = monero.Config{
			Accounts: true,
			Client:   client,
		}
		wallet := monero.New(config)
		testsuite.Test(t, 5*time.Minute, wallet, &testsuite2.MoneroGenerator{})
	})
}
