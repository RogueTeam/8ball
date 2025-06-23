package wallets

import (
	"context"
	"encoding/json"
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
	AddressRequest struct {
		// Index of the address
		Index uint64
	}
	NewAddressRequest struct {
		// Label for the new address
		Label string
	}
	Address struct {
		// Address of the address
		Address string
		// Index of the address
		Index uint64
		// Total balance of the address
		Balance uint64
		// Balannce ready to use
		UnlockedBalance uint64
	}
	SweepRequest struct {
		// Source address index
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
		Address string
		// Source address index
		SourceIndex uint64
		// Destination Address
		Destination string
		// Amount transfered
		Amount uint64
		// Fee applied to the transaction
		Fee uint64
	}
	TransferRequest struct {
		// Source address index
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
		// Source address index
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
		SourceIndex   uint64
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
	// Sync the wallet
	Sync(ctx context.Context, full bool) (err error)

	// Create a new address associate with the address
	NewAddress(ctx context.Context, req NewAddressRequest) (address Address, err error)

	// Transfers the entire balance of an address to destination
	SweepAll(ctx context.Context, req SweepRequest) (sweep Sweep, err error)

	// Transfers to a destination address
	Transfer(ctx context.Context, req TransferRequest) (transfer Transfer, err error)

	// Returns the balance of the opened wallet
	Address(ctx context.Context, req AddressRequest) (address Address, err error)

	// Validate if a monero is valid or not
	ValidateAddress(ctx context.Context, req ValidateAddressRequest) (valid ValidateAddress, err error)

	// Query the status of a transaction
	Transaction(ctx context.Context, req TransactionRequest) (tx Transaction, err error)
}

func (a *Address) String() (s string) {
	contents, _ := json.Marshal(a)
	return string(contents)
}
