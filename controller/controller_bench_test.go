package controller_test

import (
	"testing"
	"time"

	"anarchy.ttfm/8ball/blockchains"
	"anarchy.ttfm/8ball/blockchains/mock"
	"anarchy.ttfm/8ball/controller"
	"anarchy.ttfm/8ball/random"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

// Benchmark_ProcessManyPayments measures the performance of processing a large number of payments.
// It sets up 10,000,000 payments and then benchmarks a single call to ctrl.Process().
func Benchmark_Insertion(b *testing.B) {
	b.StopTimer()

	assertions := assert.New(b)
	// --- Setup Phase (Executed once before benchmarking) ---

	// Initialize the mock wallet
	wallet := mock.New()

	// Create a beneficiary account for collecting fees
	// The label is randomized to ensure uniqueness if multiple benchmarks run in parallel,
	// though for a single benchmark run, a fixed label would also suffice.
	beneficiaryLabel := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
	beneficiary, err := wallet.NewAccount(blockchains.NewAccountRequest{Label: beneficiaryLabel})
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
	var config = controller.Config{
		DB:          db,
		Fee:         1, // A nominal fee for each payment
		Timeout:     5 * time.Second,
		Beneficiary: beneficiary.Address,
		Wallet:      wallet,
	}
	// Create the controller instance
	ctrl := controller.New(config)

	// Define the number of payments to create for this benchmark scenario
	const numPayments = 1_000_000 // million payments

	receiverLabel := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
	receiver, err := wallet.NewAccount(blockchains.NewAccountRequest{Label: receiverLabel})
	assertions.Nil(err, "failed to create receiver account")

	b.ResetTimer()
	b.StartTimer()
	for range b.N {
		for i := 0; i < numPayments; i++ {
			// Create the payment intent in the controller.
			_, err := ctrl.New(receiver.Address, 10_000, blockchains.PriorityHigh)
			assertions.Nil(err, "failed to create payment")
		}
	}
	b.StopTimer()

}
