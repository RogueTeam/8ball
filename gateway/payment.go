package gateway

import (
	"encoding/json"
	"fmt"
	"time"

	"anarchy.ttfm/8ball/blockchains"
	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusExpired   Status = "expired"
	StatusError     Status = "error"
)

func PendingKey(id uuid.UUID) (key string) {
	return fmt.Sprintf("/pending/%s", id)
}

func PaymentKey(id uuid.UUID) (key string) {
	return fmt.Sprintf("/payments/%s", id)
}

type Payment struct {
	// Identifier of the transaction
	Id uuid.UUID
	// Status of the payment
	Status Status
	// Expiration time of the payment
	Expiration time.Time
	// Overall amount to expect from the transaction
	Amount uint64
	// Priority of the transaction
	Priority blockchains.Priority
	// Fee percentage to discount from the transaction
	Fee uint64
	// Address that should receive the funds
	Receiver string
	// Gateway address for receiving the transaction
	ReceiverIndex uint64
	// Beneficiary address to forward funds
	Beneficiary string
	// Error message
	Error string
	// Confirmation if Fee could be discounted and payed to beneficiary
	FeePayed bool
	// Confirmation if destination could receive its money
	DestinationPayed bool
}

func (p *Payment) Bytes() (bytes []byte) {
	bytes, _ = json.Marshal(p)
	return bytes
}

func (p *Payment) FromBytes(b []byte) (err error) {
	return json.Unmarshal(b, p)
}
