package testsuite

import (
	"encoding/json"
	"testing"
	"time"

	_ "embed"

	"github.com/RogueTeam/8ball/payments"
	"github.com/RogueTeam/8ball/random"
	"github.com/RogueTeam/8ball/utils"
	"github.com/RogueTeam/8ball/wallets"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// DataGenerator defines an interface for test data generation.
type DataGenerator interface {
	// TransferAmount returns the amount to send for a transfer.
	TransferAmount() (funds uint64)
}

//go:embed tests/succeed.yaml
var succeedTests []byte

// Test runs a comprehensive suite of tests for any Wallet implementation.
func Test(t *testing.T, timeoutExtra time.Duration, wallet wallets.Wallet, gen DataGenerator) {
	t.Run("Succeed", func(t *testing.T) {
		assertions := assert.New(t)

		type Expect struct {
			Status payments.Status `yaml:"status"`
		}
		type Test struct {
			Parts         uint64        `yaml:"parts"`
			FullFillParts uint64        `yaml:"fullfill-parts"`
			Timeout       time.Duration `yaml:"timeout"`
			TransferDelay time.Duration `yaml:"transfer-delay"`
			ProcessDelay  time.Duration `yaml:"process-delay"`
			Expect        Expect        `yaml:"expect"`
		}

		var tests []Test
		err := yaml.Unmarshal(succeedTests, &tests)
		assertions.Nil(err, "failed to load tests")

		for _, test := range tests {
			name, _ := json.Marshal(test)
			t.Run(string(name), func(t *testing.T) {
				t.Parallel()
				assertions := assert.New(t)

				ctx, cancel := utils.NewContext()
				defer cancel()

				label1 := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
				gatewayAddress, err := wallet.NewAddress(ctx, wallets.NewAddressRequest{Label: label1})
				assertions.Nil(err, "failed to create beneficiary account")

				options := badger.
					DefaultOptions("").
					WithInMemory(true)
				db, err := badger.Open(options)
				assertions.Nil(err, "failed to open database")
				var config = payments.Config{
					DB:          db,
					Timeout:     timeoutExtra + test.Timeout,
					Beneficiary: gatewayAddress.Address,
					Wallet:      wallet,
				}
				ctrl := payments.New(config)
				// t.Logf("Create controller: %+v", ctrl)

				payment, err := ctrl.Receive(payments.Receive{
					Amount:   gen.TransferAmount(),
					Priority: wallets.PriorityHigh,
				})
				assertions.Nil(err, "failed to create payment")
				// t.Logf("Create payment: %+v", payment)

				// Query first
				firstQuery, err := ctrl.Query(payment.Id)
				assertions.Nil(err, "failed to query first payment")

				assertions.Equal(payment.Id, firstQuery.Id, "Don't equal")

				// Pay the dst
				t.Log("[*] Transfering funds")
				go func() {
					for range test.FullFillParts {
						t.Log("[*] Transfer delay...", test.TransferDelay)
						time.Sleep(test.TransferDelay)
						transfer, err := wallet.Transfer(ctx, wallets.TransferRequest{
							SourceIndex: 0,
							Destination: payment.Receiver.Address,
							Amount:      gen.TransferAmount() / test.Parts,
							Priority:    wallets.PriorityHigh,
							UnlockTime:  0,
						})
						assertions.Nil(err, "failed to transfer to destination")

						cc, _ := json.Marshal(transfer)
						t.Log("[+] Transfered:", string(cc))
					}
				}()

				// Process
				t.Log("[*] Process delay...", test.ProcessDelay)
				time.Sleep(test.ProcessDelay)
				t.Log("[*] Processing payments")
				var paymentLatest payments.Payment
				for try := range 3_600 {
					t.Log("\t[*] Try processing payments: ", try+1)

					processed, err := ctrl.Process()
					assertions.Nil(err, "failed to process payments")

					// Verify payment
					paymentLatest, err = ctrl.Query(payment.Id)
					assertions.Nil(err, "failed to query payment")

					if processed == 0 || paymentLatest.Status != payments.StatusPending {
						break
					}
					time.Sleep(time.Second)
				}

				assertions.Equal(test.Expect.Status, paymentLatest.Status, "invalid status")

				if test.Expect.Status == payments.StatusExpired {
					t.Log("[*] Early return expired payment doesn't have funds")
					return
				}

			})
		}
	})
}
