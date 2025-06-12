package gateway

import (
	"errors"
	"fmt"
	"time"

	"anarchy.ttfm/8ball/blockchains"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

// Creates a new payment address based on the passed crypto currency and amount expected to receive
// It uses the default timeout in order to prevent infinite entries
// id is the id to be used for future checks
// fee is the percentage to be discounted from the entire transaction
func (c *Controller) New(dst string, amount uint64, priority blockchains.Priority) (payment Payment, err error) {
	valid, err := c.wallet.ValidateAddress(blockchains.ValidateAddressRequest{Address: dst})
	if err != nil {
		return payment, fmt.Errorf("failed to validate dst address: %w", err)
	}

	if !valid.Valid {
		return payment, errors.New("invalid address")
	}

	err = c.db.Update(func(txn *badger.Txn) (err error) {
		// Prepare new entry
		receiver, err := c.wallet.NewAccount(blockchains.NewAccountRequest{Label: payment.Id.String()})
		if err != nil {
			return fmt.Errorf("failed to prepare receiver address: %w", err)
		}

		payment = Payment{
			Id:            uuid.New(),
			Status:        StatusPending,
			Expiration:    time.Now().Add(c.timeout),
			Amount:        amount,
			Priority:      priority,
			Fee:           c.fee,
			Beneficiary:   dst,
			Receiver:      receiver.Address,
			ReceiverIndex: receiver.Index,
		}

		// Save entry
		paymentContents := payment.Bytes()

		// Pending entry
		pendingKey := PendingKey(payment.Id)
		err = txn.Set([]byte(pendingKey), paymentContents)
		if err != nil {
			return fmt.Errorf("failed to add pending key: %w", err)
		}

		paymentKey := PaymentKey(payment.Id)
		err = txn.Set([]byte(paymentKey), paymentContents)
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
