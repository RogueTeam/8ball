package controller

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusExpired   Status = "expired"
)

type Payment struct {
	// Identifier of the transaction
	Id uuid.UUID
	// Status of the payment
	Status Status
	// Currency used in the transaction
	Currency Currency
	// Expiration time of the payment
	Expiration time.Time
	// Overall amount to expect from the transaction
	Amount uint64
	// Fee percentage to discount from the transaction
	Fee uint64
	// Gateway address for receiving the transaction
	ReceiverIndex uint64
	// Destination address to forward funds
	Destination string
}
