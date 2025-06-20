package payments

import (
	"fmt"
	"time"

	"anarchy.ttfm/8ball/utils"
	"anarchy.ttfm/8ball/wallets"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type Receive struct {
	Amount   uint64
	Priority wallets.Priority
}

// Creates a new payment address based on the passed crypto currency and amount expected to receive
// It uses the default timeout in order to prevent infinite entries
// id is the id to be used for future checks
// fee is the percentage to be discounted from the entire transaction
func (c *Controller) Receive(req Receive) (payment Payment, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	err = c.db.Update(func(txn *badger.Txn) (err error) {
		// Prepare new entry
		receiver, err := c.wallet.NewAddress(ctx, wallets.NewAddressRequest{Label: payment.Id.String()})
		if err != nil {
			return fmt.Errorf("failed to prepare receiver address: %w", err)
		}

		payment = Payment{
			Id:         uuid.New(),
			Priority:   req.Priority,
			Status:     StatusPending,
			Expiration: time.Now().Add(c.timeout),
			Amount:     req.Amount,
			Receiver: Receiver{
				Address: receiver.Address,
				Index:   receiver.Index,
			},
			Beneficiary: Beneficiary{
				Address: c.beneficiary,
			},
		}

		// Pending entry
		err = txn.Set(PendingKey(payment.Id), payment.Id[:])
		if err != nil {
			return fmt.Errorf("failed to add pending key: %w", err)
		}

		// Save entry
		paymentContents := payment.Bytes()
		err = txn.Set(PaymentKey(payment.Id), paymentContents)
		if err != nil {
			return fmt.Errorf("failed to set payment status: %w", err)
		}

		return nil
	})
	if err != nil {
		return payment, fmt.Errorf("failed to add entry to the database: %w", err)
	}
	return payment, nil
}
