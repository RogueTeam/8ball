package payments

import (
	"errors"
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

// Queries a payment by its id or updates it in case it already expired
func (c *Controller) Query(id uuid.UUID) (payment Payment, err error) {
	err = c.db.View(func(txn *badger.Txn) (err error) {
		entry, err := txn.Get(PaymentKey(id))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrPaymentNotFound
			}
			return fmt.Errorf("failed to query existing payment: %w", err)
		}

		err = entry.Value(func(val []byte) (err error) {
			err = payment.FromBytes(val)
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

	return payment, nil
}
