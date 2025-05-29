package monero_test

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"anarchy.ttfm.onion/gateway/blockchains/monero"
	"anarchy.ttfm.onion/gateway/blockchains/testsuite"
	"github.com/dev-warrior777/go-monero/rpc"
	"github.com/gabstv/httpdigest"
	"github.com/stretchr/testify/assert"
)

const (
	defaultUsername = "gateway"
	defaultPassword = "gateway"
	defaultCreds    = defaultUsername + ":" + defaultPassword
	testingAccount  = "test"
	daemonIp        = "127.0.0.1"
	daemonPort      = "18081"
	daemonAddress   = daemonIp + ":" + daemonPort
	walletIp        = "127.0.0.1"
	walletPort      = "18082"
	walletAddress   = walletIp + ":" + walletPort
)

var (
	walletDir      string
	walletFilename string
	walletPassword string
)

func init() {
	walletDir = os.Getenv("WALLET_DIR")
	if walletDir == "" {
		log.Fatal("WALLET_DIR not set")
	}
	walletFilename = os.Getenv("WALLET_FILENAME")
	if walletFilename == "" {
		log.Fatal("WALLET_FILENAME not set")
	}
	walletPassword = os.Getenv("WALLET_PASSWORD")
	if walletPassword == "" {
		log.Fatal("WALLET_PASSWORD not set")
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

func prepareMonero(t *testing.T) (kill func()) {
	assertions := assert.New(t)

	var monerodOut bytes.Buffer
	monerod := exec.Command("monerod",
		"--non-interactive",
		"--testnet",
		"--rpc-login", defaultCreds,
		"--rpc-bind-ip", daemonIp,
		"--rpc-bind-port", daemonPort,
	)
	monerod.Stdout = &monerodOut
	monerod.Stderr = &monerodOut
	err := monerod.Start()
	assertions.Nil(err, "failed to start monerod")

	forceConnection(t, daemonAddress)

	var walletRpcOut bytes.Buffer
	walletRpc := exec.Command("monero-wallet-rpc",
		"--testnet",
		"--trusted-daemon",
		"--non-interactive",
		"--rpc-bind-ip", walletIp,
		"--rpc-bind-port", walletPort,
		"--daemon-address", daemonAddress,
		"--daemon-login", defaultCreds,
		"--rpc-login", defaultCreds,
		"--wallet-dir", walletDir,
	)
	walletRpc.Stdout = &walletRpcOut
	walletRpc.Stderr = &walletRpcOut
	err = walletRpc.Start()
	assertions.Nil(err, "failed to start wallet rpc")

	forceConnection(t, walletAddress)

	return func() {
		fmt.Println("=== Monerod RPC ===")
		fmt.Println(monerodOut.String())
		fmt.Println("=== Wallet RPC ===")
		fmt.Println(walletRpcOut.String())

		err := monerod.Process.Kill()
		assertions.Nil(err, "failed to kill monerod")
		err = walletRpc.Process.Kill()
		assertions.Nil(err, "failed to kill wallet-rpc")

		_, err = monerod.Process.Wait()
		assertions.Nil(err, "failed to wait monerod")
		_, err = walletRpc.Process.Wait()
		assertions.Nil(err, "failed to wait wallet-rpc")
	}
}

func newClient(t *testing.T) (client *rpc.Client) {
	assertions := assert.New(t)

	var config = rpc.Config{
		Address: "http://" + walletAddress + "/json_rpc",

		Client: &http.Client{
			Transport: httpdigest.New(defaultUsername, defaultPassword), // Remove if no auth.
		},
	}
	client = rpc.New(config)
	assertions.NotNil(client, "failed to create client")

	return client
}

func Test_Monero(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		defer prepareMonero(t)()

		assertions := assert.New(t)

		client := newClient(t)
		log.Println(walletFilename, walletPassword)
		var config = monero.Config{
			Client:   client,
			Account:  testingAccount,
			Filename: walletFilename,
			Password: walletPassword,
		}
		wallet, err := monero.New(config)
		assertions.Nil(err, "failed to create wallet manager")

		testsuite.Test(t, &wallet)
	})
	// TODO: Implement me
}
