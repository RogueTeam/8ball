package blockchains

import (
	"context"
	"errors"
)

var ErrInvalidPriority = errors.New("invalid priority")

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "meidum"
	PriorityHigh   Priority = "high"
)

type Priority string

type (
	AccountRequest struct {
		// Index of the account
		Index uint64
	}
	NewAccountRequest struct {
		// Label for the new account
		Label string
	}
	Account struct {
		// Address of the account
		Address string
		// Index of the account
		Index uint64
		// Total balance of the address
		Balance uint64
		// Balannce ready to use
		UnlockedBalance uint64
	}
	SweepRequest struct {
		// Source account index
		SourceIndex uint64
		// Destination Address
		Destination string
		// Priority of the transaction
		Priority Priority
		// Unlock time (blocks)
		UnlockTime uint64
	}
	Sweep struct {
		// Address of the transaction
		Address []string
		// Source account index
		SourceIndex uint64
		// Destination Address
		Destination string
		// Amount transfered
		Amount []uint64
		// Fee applied to the transaction
		Fee []uint64
	}
	TransferRequest struct {
		// Source account index
		SourceIndex uint64
		// Destination Address
		Destination string
		// Amount transfered
		Amount uint64
		// Priority of the transaction
		Priority Priority
		// Unlock time (blocks)
		UnlockTime uint64
	}
	Transfer struct {
		// Address of the transaction
		Address string
		// Source account index
		SourceIndex uint64
		// Destination Address
		Destination string
		// Amount transfered
		Amount uint64
		// Fee applied to the transaction
		Fee uint64
	}
	ValidateAddressRequest struct {
		Address string
	}
	ValidateAddress struct {
		Valid bool
	}
	TransactionRequest struct {
		TransactionId string
	}
	Transaction struct {
		Address     string
		Amount      uint64
		Destination string
		Status      TransactionStatus
	}
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

type Wallet interface {
	// Create a new address associate with the account
	NewAccount(ctx context.Context, req NewAccountRequest) (account Account, err error)

	// Transfers the entire balance of an address to destination
	SweepAll(ctx context.Context, req SweepRequest) (sweep Sweep, err error)

	// Transfers to a destination address
	Transfer(ctx context.Context, req TransferRequest) (transfer Transfer, err error)

	// Returns the balance of the opened wallet
	Account(ctx context.Context, req AccountRequest) (account Account, err error)

	// Validate if a monero is valid or not
	ValidateAddress(ctx context.Context, req ValidateAddressRequest) (valid ValidateAddress, err error)

	// Query the status of a transaction
	Transaction(ctx context.Context, req TransactionRequest) (tx Transaction, err error)
}
