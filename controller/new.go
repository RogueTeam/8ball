package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"anarchy.ttfm.onion/gateway/blockchains"
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
		payment = Payment{
			Id:          uuid.New(),
			Status:      StatusPending,
			Expiration:  time.Now().Add(c.timeout),
			Amount:      amount,
			Priority:    priority,
			Fee:         c.fee,
			Destination: dst,
		}

		// Pending entry
		err = txn.Set([]byte(fmt.Sprintf("/pending/%s", payment.Id)), payment.Id[:])
		if err != nil {
			return fmt.Errorf("failed to add pending key: %w", err)
		}

		// Prepare new entry
		receiver, err := c.wallet.NewAccount(blockchains.NewAccountRequest{Label: payment.Id.String()})
		if err != nil {
			return fmt.Errorf("failed to prepare receiver address: %w", err)
		}

		// Add address to output struct
		payment.ReceiverIndex = receiver.Index

		// Save entry
		paymentKey := fmt.Sprintf("/payments/%s", payment.Id)
		paymentContents, err := json.Marshal(&payment)
		if err != nil {
			return fmt.Errorf("failed to marshal payment status: %w", err)
		}
		err = txn.Set([]byte(paymentKey), paymentContents)
		if err != nil {
			return fmt.Errorf("failed to set payment status: %w", err)
		}

		return nil
	})
	if err != nil {
		return payment, fmt.Errorf("faied to add entry to the database: %w", err)
	}
	return payment, nil
}
