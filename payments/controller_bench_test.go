package payments_test

import (
	"testing"
	"time"

	"anarchy.ttfm/8ball/payments"
	"anarchy.ttfm/8ball/random"
	"anarchy.ttfm/8ball/utils"
	"anarchy.ttfm/8ball/wallets"
	"anarchy.ttfm/8ball/wallets/mock"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

// Benchmark_ProcessManyPayments measures the performance of processing a large number of payments.
// It sets up 10,000,000 payments and then benchmarks a single call to ctrl.Process().
func Benchmark_Insertion(b *testing.B) {
	b.StopTimer()

	assertions := assert.New(b)

	ctx, cancel := utils.NewContext()
	defer cancel()

	// --- Setup Phase (Executed once before benchmarking) ---

	// Initialize the mock wallet
	wallet := mock.New(mock.Config{FundsDelta: 0})

	// Create a beneficiary account for collecting fees
	// The label is randomized to ensure uniqueness if multiple benchmarks run in parallel,
	// though for a single benchmark run, a fixed label would also suffice.
	beneficiaryLabel := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
	beneficiary, err := wallet.NewAddress(ctx, wallets.NewAddressRequest{Label: beneficiaryLabel})
	assertions.Nil(err, "failed to create beneficiary account")

	// Configure and open an in-memory BadgerDB instance
	// WithInMemory(true) ensures the database is temporary and cleaned up automatically.
	options := badger.
		DefaultOptions("").
		WithLoggingLevel(5).
		WithLogger(nil).
		WithInMemory(true)
	db, err := badger.Open(options)
	assertions.Nil(err, "failed to open database")
	defer db.Close()

	// Configure the controller with the database, fee, timeout, beneficiary, and wallet
	var config = payments.Config{
		DB:          db,
		Timeout:     5 * time.Second,
		Beneficiary: beneficiary.Address,
		Wallet:      wallet,
	}
	// Create the controller instance
	ctrl := payments.New(config)

	// Define the number of payments to create for this benchmark scenario
	const numPayments = 1_000_000 // million payments

	assertions.Nil(err, "failed to create receiver account")

	b.ResetTimer()
	b.StartTimer()
	for range b.N {
		for i := 0; i < numPayments; i++ {
			// Create the payment intent in the controller.
			_, err := ctrl.Receive(payments.Receive{
				Amount:   10_000,
				Priority: wallets.PriorityHigh,
			})
			assertions.Nil(err, "failed to create payment")
		}
	}
	b.StopTimer()

}
