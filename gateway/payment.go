package gateway

import (
	"encoding/json"
	"fmt"
	"time"

	"anarchy.ttfm/8ball/wallets"
	"github.com/google/uuid"
)

type Status string

const (
	StatusPending            Status = "pending"
	StatusCompleted          Status = "completed"
	StatusPartiallyCompleted Status = "partially-completed"
	StatusExpired            Status = "expired"
	StatusError              Status = "error"
)

func PendingKey(id uuid.UUID) (key []byte) {
	return []byte(fmt.Sprintf("/pending/%s", id))
}

func PaymentKey(id uuid.UUID) (key []byte) {
	return []byte(fmt.Sprintf("/payments/%s", id))
}

type (
	Receiver struct {
		// Address that will receive the funds
		Address string
		// Index of the address
		Index uint64
	}
	Beneficiary struct {
		// Address of the beneficiary during this transaction
		Address string
		// Actual amount payed to the Beneficiary
		Payed uint64
		// Address of the transactio that was used to pay the beneficiary
		Transaction string
	}
	Payment struct {
		// Identifier of the transaction
		Id uuid.UUID
		// Priority to forward funds to beneficiary
		Priority wallets.Priority
		// Status of the payment
		Status Status
		// Expiration time of the payment
		Expiration time.Time
		// Overall amount to expect from the transaction
		Amount uint64
		// The receiver is the address used to receive the payment
		Receiver Receiver
		// Beneficiary information. Stored in case wallet changes
		Beneficiary Beneficiary
		// Error message
		Error string
	}
)

func (p *Payment) SetError(err error) {
	if err == nil {
		return
	}

	p.Status = StatusError
	p.Error = err.Error()
}

func (p *Payment) Bytes() (bytes []byte) {
	bytes, _ = json.Marshal(p)
	return bytes
}

func (p *Payment) FromBytes(b []byte) (err error) {
	return json.Unmarshal(b, p)
}
