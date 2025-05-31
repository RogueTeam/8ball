package monero_test

import (
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"anarchy.ttfm.onion/gateway/blockchains/monero"
	"anarchy.ttfm.onion/gateway/blockchains/testsuite"
	"github.com/dev-warrior777/go-monero/rpc"
	"github.com/gabstv/httpdigest"
	"github.com/stretchr/testify/assert"
)

var (
	walletDir      string
	walletFilename string
	accountName    string
)

func init() {
	walletFilename = os.Getenv("WALLET_FILENAME")
	if walletFilename == "" {
		log.Fatal("WALLET_FILENAME not set")
	}
	accountName = os.Getenv("ACCOUNT_NAME")
	if accountName == "" {
		log.Fatal("ACCOUNT_NAME not set")
	}
}

func forceConnection(t *testing.T, addr string) {
	assertions := assert.New(t)
	var found bool
	for range 10 {
		time.Sleep(time.Second)
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			found = true
			break
		}
	}

	assertions.True(found, "could not connect to "+addr)
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

func Test_Monero(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		assertions := assert.New(t)

		client := newClient(t)
		var config = monero.Config{
			Client:   client,
			Account:  accountName,
			Filename: walletFilename,
		}
		wallet, err := monero.New(config)
		assertions.Nil(err, "failed to create wallet manager")

		testsuite.Test(t, &wallet)
	})
	// TODO: Implement me
}
