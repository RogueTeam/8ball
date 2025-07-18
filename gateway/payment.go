package gateway

import (
	"encoding/json"
	"time"

	"github.com/RogueTeam/8ball/wallets"
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

const (
	feePrefix      = "/fee/"
	pendingPrefix  = "/pending/"
	paymentsPrefix = "/payment/"
)

var (
	pendingPrefixBytes = []byte(pendingPrefix)
	feePrefixBytes     = []byte(feePrefix)
)

func FeeKey(id uuid.UUID) (key []byte) {
	return []byte(feePrefix + id.String())
}

func PendingKey(id uuid.UUID) (key []byte) {
	return []byte(pendingPrefix + id.String())
}

func PaymentKey(id uuid.UUID) (key []byte) {
	return []byte(paymentsPrefix + id.String())
}

type (
	Receiver struct {
		// Address that will receive the funds
		Address string
		// Index of the address
		Index uint64
	}
	Beneficiary struct {
		// Status of the payment
		Status Status
		// Error message
		Error string
		// Address of the beneficiary during this transaction
		Address string
		// Actual amount payed to the Beneficiary
		Payed uint64
		// Address of the transactio that was used to pay the beneficiary
		Transaction string
	}
	Fee struct {
		// Status of the payment
		Status Status
		// Error message
		Error string
		// Percentage to be payed
		Percentage uint64
		// Address of the account that will the fee profit
		Address string
		// Actual amount payed to the account
		Payed uint64
		// Transaction that was used to pay the fee
		Transaction string
	}
	Payment struct {
		// Identifier of the transaction
		Id uuid.UUID
		// Priority to forward funds to beneficiary
		Priority wallets.Priority
		// Overall amount to expect from the transaction
		Amount uint64
		// Expiration time of the payment
		Expiration time.Time
		// The receiver is the address used to receive the payment
		Receiver Receiver
		// Fee details
		Fee Fee
		// Beneficiary information. Stored in case wallet changes
		Beneficiary Beneficiary
	}
)

func (b *Beneficiary) SetError(err error) {
	if err == nil {
		return
	}

	b.Status = StatusError
	b.Error = err.Error()
}

func (f *Fee) SetError(err error) {
	if err == nil {
		return
	}

	f.Status = StatusError
	f.Error = err.Error()
}

func (p *Payment) Bytes() (bytes []byte) {
	bytes, _ = json.Marshal(p)
	return bytes
}

func (p *Payment) FromBytes(b []byte) (err error) {
	return json.Unmarshal(b, p)
}
