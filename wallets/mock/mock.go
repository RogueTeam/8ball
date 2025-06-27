package mock

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	wallets "anarchy.ttfm/8ball/wallets"
)

var (
	ErrAddressNotFound     = errors.New("address not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrTransactionNotFound = errors.New("transaction not found")
)

type Transaction struct {
	Status   wallets.TransactionStatus
	Sweep    *wallets.Sweep
	Transfer *wallets.Transfer
}

// Mock implements the wallets.Wallet interface for testing purposes.
type Mock struct {
	mu             sync.Mutex
	addresses      map[uint64]wallets.Address // index -> Account
	nextIndex      uint64
	transactions   map[string]Transaction // txHash -> transaction details (for tracking)
	fundsDelta     time.Duration
	zeroOnTransfer bool
}

var _ wallets.Wallet = (*Mock)(nil)

type Config struct {
	FundsDelta     time.Duration
	ZeroOnTransfer bool
}

// New creates a new Mock wallet.
func New(config Config) *Mock {
	m := &Mock{
		addresses:    make(map[uint64]wallets.Address),
		nextIndex:    0, // Start nextIndex at 0
		transactions: make(map[string]Transaction),
		fundsDelta:   config.FundsDelta,
	}

	// Initialize with a zero-index account
	zeroAccount := wallets.Address{
		Address:         "mock_address_0", // A default address for the initial account
		Index:           0,
		Balance:         1_000_000_000_000,
		UnlockedBalance: 1_000_000_000_000,
	}
	m.addresses[0] = zeroAccount
	m.nextIndex++ // Increment nextIndex after setting up the initial account

	return m
}

func (m *Mock) Sync(ctx context.Context, _ bool) (err error) { return nil }

// NewAddress creates a new mock account.
func (m *Mock) NewAddress(ctx context.Context, req wallets.NewAddressRequest) (address wallets.Address, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// In a mock, the label isn't strictly used for uniqueness,
	// but we can simulate creating a new address.
	newAddress := fmt.Sprintf("mock_address_%d", m.nextIndex)

	address = wallets.Address{
		Address:         newAddress,
		Index:           m.nextIndex,
		Balance:         0,
		UnlockedBalance: 0,
	}
	m.addresses[m.nextIndex] = address
	m.nextIndex++
	return address, nil
}

// SweepAll transfers the entire balance of an account to a destination.
func (m *Mock) SweepAll(ctx context.Context, req wallets.SweepRequest) (sweep wallets.Sweep, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sourceAccount, ok := m.addresses[req.SourceIndex]
	if !ok {
		return sweep, ErrAddressNotFound
	}

	if sourceAccount.UnlockedBalance == 0 {
		return sweep, fmt.Errorf("source account %d has no balance to sweep", req.SourceIndex)
	}

	// For a mock, we just move the balance and generate a mock transaction hash.
	var transferredAmount uint64
	var appliedFee uint64
	if sourceAccount.UnlockedBalance > DefaultFee {
		transferredAmount = sourceAccount.UnlockedBalance - transferredAmount
		appliedFee = transferredAmount
	} else {
		transferredAmount = sourceAccount.UnlockedBalance
	}

	sourceAccount.Balance = 0
	sourceAccount.UnlockedBalance = 0
	m.addresses[req.SourceIndex] = sourceAccount

	mockTxHash := fmt.Sprintf("mock_sweep_tx_%d_%s", req.SourceIndex, req.Destination)

	sweep = wallets.Sweep{
		Address:     mockTxHash,
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      transferredAmount, // Simulate fee deduction
		Fee:         appliedFee,
	}
	m.transactions[mockTxHash] = Transaction{Status: wallets.TransactionStatusPending, Sweep: &sweep} // Track the transaction

	for index, account := range m.addresses {
		if account.Address != req.Destination {
			continue
		}

		account.Balance += transferredAmount
		m.addresses[index] = account

		go func() {
			time.Sleep(m.fundsDelta)

			m.mu.Lock()
			defer m.mu.Unlock()

			tx := m.transactions[mockTxHash]
			tx.Status = wallets.TransactionStatusCompleted
			m.transactions[mockTxHash] = tx

			account := m.addresses[index]
			account.UnlockedBalance = account.Balance
			m.addresses[index] = account
		}()
		break
	}
	return sweep, nil
}

const DefaultFee = 50

// Transfer transfers a specified amount to a destination address.
func (m *Mock) Transfer(ctx context.Context, req wallets.TransferRequest) (transfer wallets.Transfer, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.Amount == 0 {
		return transfer, ErrInvalidAmount
	}

	sourceAccount, ok := m.addresses[req.SourceIndex]
	if !ok {
		return transfer, ErrAddressNotFound
	}

	if sourceAccount.UnlockedBalance < DefaultFee+req.Amount {
		return transfer, ErrInsufficientBalance
	}

	// For a mock, we just deduct the balance and generate a mock transaction hash.
	sourceAccount.Balance -= req.Amount + DefaultFee

	if m.zeroOnTransfer {
		sourceAccount.UnlockedBalance = 0
		go func() {
			time.Sleep(m.fundsDelta)

			m.mu.Lock()
			defer m.mu.Unlock()

			sourceAccount := m.addresses[req.SourceIndex]
			sourceAccount.UnlockedBalance = sourceAccount.Balance
			m.addresses[req.SourceIndex] = sourceAccount
		}()
	} else {
		sourceAccount.UnlockedBalance = sourceAccount.Balance // Assuming transferred amount was unlocked
	}
	m.addresses[req.SourceIndex] = sourceAccount

	mockTxHash := fmt.Sprintf("mock_transfer_tx_%d_%s_%d", req.SourceIndex, req.Destination, req.Amount)

	transfer = wallets.Transfer{
		Address:     mockTxHash,
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      req.Amount, // Simulate fee deduction
		Fee:         DefaultFee,
	}
	m.transactions[mockTxHash] = Transaction{Status: wallets.TransactionStatusPending, Transfer: &transfer} // Track the transaction

	for index, account := range m.addresses {
		if account.Address != req.Destination {
			continue
		}

		account.Balance += req.Amount
		m.addresses[index] = account

		go func() {
			time.Sleep(m.fundsDelta)

			m.mu.Lock()
			defer m.mu.Unlock()

			tx := m.transactions[mockTxHash]
			tx.Status = wallets.TransactionStatusCompleted
			m.transactions[mockTxHash] = tx

			account := m.addresses[index]
			account.UnlockedBalance = account.Balance
			m.addresses[index] = account
		}()
		break
	}
	return transfer, nil
}

// Address returns the balance of the specified account.
func (m *Mock) Address(ctx context.Context, req wallets.AddressRequest) (address wallets.Address, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc, ok := m.addresses[req.Index]
	if !ok {
		return address, ErrAddressNotFound
	}
	return acc, nil
}

// ValidateAddress always returns true for any address in the mock.
func (m *Mock) ValidateAddress(ctx context.Context, req wallets.ValidateAddressRequest) (err error) {
	// For testing, all addresses are valid.
	return nil
}

func (m *Mock) Transaction(ctx context.Context, req wallets.TransactionRequest) (tx wallets.Transaction, err error) {
	transaction, found := m.transactions[req.TransactionId]
	if !found {
		return tx, ErrTransactionNotFound
	}

	tx = wallets.Transaction{
		Address: req.TransactionId,
	}

	switch {
	case transaction.Sweep != nil:
		tx.Amount = transaction.Sweep.Amount
		tx.Destination = transaction.Sweep.Destination
		tx.Status = transaction.Status
	case transaction.Transfer != nil:
		tx.Amount = transaction.Transfer.Amount
		tx.Destination = transaction.Transfer.Destination
		tx.Status = transaction.Status
	}

	return tx, nil
}
