package router

import (
	"time"

	"anarchy.ttfm/8ball/decimal"
	"anarchy.ttfm/8ball/gateway"
	"anarchy.ttfm/8ball/wallets"
	"github.com/google/uuid"
)

const DefaultPriority = wallets.PriorityLow

type Receive struct {
	Address string          `json:"address,omitzero"`
	Amount  decimal.Decimal `json:"amount,omitzero"`
}

func ReceiveToGateway(src *Receive) (out gateway.Receive, err error) {
	out = gateway.Receive{
		Address:  src.Address,
		Amount:   src.Amount.ToUint64(),
		Priority: DefaultPriority,
	}
	return out, nil
}

type (
	Fee struct {
		// Status of the payment
		Status gateway.Status `json:"status"`
		// Percentage to be payed
		Percentage uint64 `json:"percentage"`
		// Error message
		Error string `json:"error,omitzero"`
		// Actual amount payed to the account
		Payed decimal.Decimal `json:"payed,omitzero"`
	}
	Beneficiary struct {
		// Status of the payment
		Status gateway.Status `json:"status"`
		// Error message
		Error string `json:"error,omitzero"`
		// Actual amount payed to the Beneficiary
		Payed decimal.Decimal `json:"payed,omitzero"`
	}
	Payment struct {
		// Identifier of the transaction
		Id uuid.UUID `json:"id"`
		// Overall amount to expect from the transaction
		Amount decimal.Decimal `json:"amount"`
		// Expiration time of the payment
		Expiration time.Time `json:"expiration"`
		// The receiver is the address used to receive the payment
		PaymentAddress string `json:"paymentAddress"`
		// Fee details
		Fee Fee `json:"fee"`
		// Beneficiary information. Stored in case wallet changes
		Beneficiary Beneficiary `json:"beneficiary"`
	}
)

// Convert from Gateway's Payment type to the internal Payment
// hiding sensitive values
func PaymentFromGateway(src *gateway.Payment) (payment Payment) {
	payment = Payment{
		Id:             src.Id,
		Expiration:     src.Expiration,
		PaymentAddress: src.Receiver.Address,
		Fee: Fee{
			Status:     src.Fee.Status,
			Error:      src.Fee.Error,
			Percentage: src.Fee.Percentage,
		},
		Beneficiary: Beneficiary{
			Status: src.Beneficiary.Status,
			Error:  src.Beneficiary.Error,
		},
	}
	payment.Amount.FromUint64(src.Amount)
	payment.Fee.Payed.FromUint64(src.Fee.Payed)
	payment.Beneficiary.Payed.FromUint64(src.Beneficiary.Payed)
	return payment
}
