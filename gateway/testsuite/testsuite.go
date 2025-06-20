package testsuite

import (
	"encoding/json"
	"testing"
	"time"

	_ "embed"

	"anarchy.ttfm/8ball/gateway"
	"anarchy.ttfm/8ball/random"
	"anarchy.ttfm/8ball/utils"
	"anarchy.ttfm/8ball/wallets"
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
			BeneficiaryStatus gateway.Status `yaml:"beneficiary-status"`
			FeeStatus         gateway.Status `yaml:"fee-status"`
		}
		type Test struct {
			Fee                 uint64        `yaml:"fee"`
			Parts               uint64        `yaml:"parts"`
			FullFillParts       uint64        `yaml:"fullfill-parts"`
			Timeout             time.Duration `yaml:"timeout"`
			TransferDelay       time.Duration `yaml:"transfer-delay"`
			ProcessPendingDelay time.Duration `yaml:"process-pending-delay"`
			ProcessFeeDelay     time.Duration `yaml:"process-fee-delay"`
			Expect              Expect        `yaml:"expect"`
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
				assertions.Nil(err, "failed to create gateway address")

				options := badger.
					DefaultOptions("").
					WithInMemory(true)
				db, err := badger.Open(options)
				assertions.Nil(err, "failed to open database")
				var config = gateway.Config{
					DB:            db,
					Timeout:       timeoutExtra + test.Timeout,
					FeePercentage: test.Fee,
					Address:       gatewayAddress.Address,
					Wallet:        wallet,
				}
				ctrl := gateway.New(config)
				// t.Logf("Create controller: %+v", ctrl)

				businessAddress, err := wallet.NewAddress(ctx, wallets.NewAddressRequest{Label: label1})
				assertions.Nil(err, "failed to create business address")

				payment, err := ctrl.Receive(gateway.Receive{
					Address:  businessAddress.Address,
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
				t.Log("[*] Processing pending payments delay...", test.ProcessPendingDelay)
				time.Sleep(test.ProcessPendingDelay)
				t.Log("[*] Processing pending payments")
				var paymentLatest gateway.Payment
				for try := range 3_600 {
					t.Log("\t[*] Try processing payments: ", try+1)

					err = ctrl.ProcessPendingPayments()
					assertions.Nil(err, "failed to process payments")

					// Verify payment
					paymentLatest, err = ctrl.Query(payment.Id)
					assertions.Nil(err, "failed to query payment")

					if paymentLatest.Beneficiary.Status != gateway.StatusPending {
						break
					}
					time.Sleep(time.Second)
				}

				assertions.Equal(test.Expect.BeneficiaryStatus, paymentLatest.Beneficiary.Status, "invalid benefiary status")

				if test.Expect.BeneficiaryStatus == gateway.StatusExpired {
					t.Log("[*] Early return expired payment doesn't have funds")
					return
				}

				// Process
				t.Log("[*] Processing fees delay...", test.ProcessPendingDelay)
				time.Sleep(test.ProcessFeeDelay)
				t.Log("[*] Processing fees")
				for try := range 3_600 {
					t.Log("\t[*] Try processing fees: ", try+1)

					err = ctrl.ProcessPendingFees()
					assertions.Nil(err, "failed to process fee")

					// Verify payment
					paymentLatest, err = ctrl.Query(payment.Id)
					assertions.Nil(err, "failed to query fee")

					if paymentLatest.Fee.Status != gateway.StatusPending {
						break
					}
					time.Sleep(time.Second)
				}

				assertions.Equal(test.Expect.FeeStatus, paymentLatest.Fee.Status, "invalid fee status")

			})
		}
	})
}
