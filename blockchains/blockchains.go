package blockchains

import "errors"

var ErrInvalidPriority = errors.New("invalid priority")

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "meidum"
	PriorityHigh   Priority = "high"
)

type Priority string

type (
	NewAddressRequest struct {
		Label string
	}
	Address struct {
		AccountAddress string
		AccountIndex   uint64
		Address        string
		Index          uint64
	}
	SweepRequest struct {
		// Source address
		Source string
		// Destination Address
		Destination string
		// Amount transfered
		Amount uint64
		// Priority of the transaction
		Priority Priority
		// Unlock time (blocks)
		UnlockTime uint64
	}
	Sweep struct {
		// Address of the transaction
		Address []string
		// Source address
		Source string
		// Destination Address
		Destination string
		// Amount transfered
		Amount []uint64
		// Fee applied to the transaction
		Fee []uint64
	}
	TransferRequest struct {
		// Source address
		Source string
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
		// Source address
		Source string
		// Destination Address
		Destination string
		// Amount transfered
		Amount uint64
		// Fee applied to the transaction
		Fee uint64
	}
	Balance struct {
		// Address of account or the subaddress itself
		Address string
		// Total balance of the address
		Amount uint64
		// Balannce ready to use
		Unlocked uint64
	}
	AddressBalanceRequest struct {
		Address string
	}
	ValidateAddressRequest struct {
		Address string
	}
	ValidateAddress struct {
		Valid bool
	}
)

type Wallet interface {
	// Create a new address associate with the account
	NewAddress(req NewAddressRequest) (address Address, err error)

	// Transfers the entire balance of an address to destination
	SweepAll(req SweepRequest) (sweep Sweep, err error)

	// Transfers to a destination address
	Transfer(req TransferRequest) (transfer Transfer, err error)

	// Returns the balance of the opened wallet account
	Balance() (balance Balance, err error)

	// Returns the balance of an specific address
	AddressBalance(address AddressBalanceRequest) (balance Balance, err error)

	// Validate if a monero is valid or not
	ValidateAddress(req ValidateAddressRequest) (valid ValidateAddress, err error)
}
