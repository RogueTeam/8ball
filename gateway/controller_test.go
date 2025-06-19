package gateway_test

import (
	"encoding/json"
	"testing"
	"time"

	"anarchy.ttfm/8ball/gateway/testsuite"
	"anarchy.ttfm/8ball/wallets/mock"
	testsuite2 "anarchy.ttfm/8ball/wallets/testsuite"
)

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
				testsuite.Test(t, w, &testsuite2.MockGenerator{})
			})
		}
	})
	t.Run("Monero", func(t *testing.T) {
		t.Skip()
		return
		//t.Parallel()
		//assertions := assert.New(t)
		//
		//walletFilename := os.Getenv("MONERO_WALLET_FILENAME")
		//assertions.NotEmpty(walletFilename, "MONERO_WALLET_FILENAME")
		//walletPassword := os.Getenv("MONERO_WALLET_PASSWORD")
		//assertions.NotEmpty(walletPassword, "MONERO_WALLET_PASSWORD")
		//
		//var config = rpc.Config{
		//	Address: "http://127.0.0.1:22222/json_rpc",
		//
		//	Client: &http.Client{
		//		Transport: httpdigest.New("username", "password"), // Remove if no auth.
		//	},
		//}
		//
		//client := rpc.New(config)
		//
		//w := monero.New(monero.Config{Client: client})
		//testsuite.Test(t, w, &testsuite2.MoneroGenerator{})
	})
}
