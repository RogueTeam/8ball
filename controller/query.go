package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

// Queries a payment by its id or updates it in case it already expired
func (c *Controller) QueryOrUpdate(id uuid.UUID) (payment Payment, err error) {
	now := time.Now()
	paymentKey := []byte(fmt.Sprintf("/payments/%s", id))

	err = c.db.View(func(txn *badger.Txn) (err error) {
		entry, err := txn.Get(paymentKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrPaymentNotFound
			}
			return fmt.Errorf("failed to query existing payment: %w", err)
		}

		err = entry.Value(func(val []byte) (err error) {
			err = json.Unmarshal(val, &payment)
			if err != nil {
				return fmt.Errorf("failed to unmarshal payment: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to retrieve value: %w", err)
		}
		return nil
	})
	if err != nil {
		return payment, fmt.Errorf("faied to query entry from the database: %w", err)
	}

	if payment.Expiration.After(now) {
		return payment, nil
	}

	// In case it already expired
	payment.Status = StatusExpired
	err = c.db.Update(func(txn *badger.Txn) (err error) {
		paymentContents, err := json.Marshal(&payment)
		if err != nil {
			return fmt.Errorf("failed to marshal payment: %w", err)
		}

		err = txn.Set(paymentKey, paymentContents)
		if err != nil {
			return fmt.Errorf("failed to set payment: %w", err)
		}
		return nil
	})
	if err != nil {
		return payment, fmt.Errorf("failed to update payment status: %w", err)
	}

	// TODO: Refund if necessary

	return payment, nil
}
