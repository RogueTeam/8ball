package mock

import (
	"errors"
	"fmt"
	"sync"

	"anarchy.ttfm/8ball/blockchains"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
)

// Mock implements the blockchains.Wallet interface for testing purposes.
type Mock struct {
	mu           sync.Mutex
	accounts     map[uint64]blockchains.Account // index -> Account
	nextIndex    uint64
	transactions map[string]interface{} // txHash -> transaction details (for tracking)
}

var _ blockchains.Wallet = (*Mock)(nil)

// New creates a new Mock wallet.
func New() *Mock {
	m := &Mock{
		accounts:     make(map[uint64]blockchains.Account),
		nextIndex:    0, // Start nextIndex at 0
		transactions: make(map[string]interface{}),
	}

	// Initialize with a zero-index account
	zeroAccount := blockchains.Account{
		Address:         "mock_address_0", // A default address for the initial account
		Index:           0,
		Balance:         1_000_000_000,
		UnlockedBalance: 1_000_000_000,
	}
	m.accounts[0] = zeroAccount
	m.nextIndex++ // Increment nextIndex after setting up the initial account

	return m
}

// NewAccount creates a new mock account.
func (m *Mock) NewAccount(req blockchains.NewAccountRequest) (account blockchains.Account, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// In a mock, the label isn't strictly used for uniqueness,
	// but we can simulate creating a new address.
	newAddress := fmt.Sprintf("mock_address_%d", m.nextIndex)

	account = blockchains.Account{
		Address:         newAddress,
		Index:           m.nextIndex,
		Balance:         0,
		UnlockedBalance: 0,
	}
	m.accounts[m.nextIndex] = account
	m.nextIndex++
	return account, nil
}

// SweepAll transfers the entire balance of an account to a destination.
func (m *Mock) SweepAll(req blockchains.SweepRequest) (sweep blockchains.Sweep, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sourceAccount, ok := m.accounts[req.SourceIndex]
	if !ok {
		return sweep, ErrAccountNotFound
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
	m.accounts[req.SourceIndex] = sourceAccount

	mockTxHash := fmt.Sprintf("mock_sweep_tx_%d_%s", req.SourceIndex, req.Destination)

	sweep = blockchains.Sweep{
		Address:     []string{mockTxHash},
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      []uint64{transferredAmount}, // Simulate fee deduction
		Fee:         []uint64{appliedFee},
	}
	m.transactions[mockTxHash] = sweep // Track the transaction

	for index, account := range m.accounts {
		if account.Address != req.Destination {
			continue
		}
		account.Balance += transferredAmount
		account.UnlockedBalance += transferredAmount
		m.accounts[index] = account
		break
	}
	return sweep, nil
}

const DefaultFee = 50

// Transfer transfers a specified amount to a destination address.
func (m *Mock) Transfer(req blockchains.TransferRequest) (transfer blockchains.Transfer, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.Amount == 0 {
		return transfer, ErrInvalidAmount
	}

	sourceAccount, ok := m.accounts[req.SourceIndex]
	if !ok {
		return transfer, ErrAccountNotFound
	}

	if sourceAccount.UnlockedBalance < DefaultFee+req.Amount {
		return transfer, ErrInsufficientBalance
	}

	// For a mock, we just deduct the balance and generate a mock transaction hash.
	sourceAccount.Balance -= req.Amount + DefaultFee
	sourceAccount.UnlockedBalance -= req.Amount + DefaultFee // Assuming transferred amount was unlocked
	m.accounts[req.SourceIndex] = sourceAccount

	mockTxHash := fmt.Sprintf("mock_transfer_tx_%d_%s_%d", req.SourceIndex, req.Destination, req.Amount)

	transfer = blockchains.Transfer{
		Address:     mockTxHash,
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      req.Amount, // Simulate fee deduction
		Fee:         DefaultFee,
	}
	m.transactions[mockTxHash] = transfer // Track the transaction

	for index, account := range m.accounts {
		if account.Address != req.Destination {
			continue
		}
		account.Balance += req.Amount
		account.UnlockedBalance += req.Amount
		m.accounts[index] = account
		break
	}
	return transfer, nil
}

// Account returns the balance of the specified account.
func (m *Mock) Account(req blockchains.AccountRequest) (account blockchains.Account, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc, ok := m.accounts[req.Index]
	if !ok {
		return account, ErrAccountNotFound
	}
	return acc, nil
}

// ValidateAddress always returns true for any address in the mock.
func (m *Mock) ValidateAddress(req blockchains.ValidateAddressRequest) (valid blockchains.ValidateAddress, err error) {
	// For testing, all addresses are valid.
	return blockchains.ValidateAddress{Valid: true}, nil
}
