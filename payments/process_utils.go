package payments

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

var pendingPrefix = []byte("/pending/")

// Streams pending payments into a channel. Its intended be used in parallel while querying wallets
// payments channel should must be consumed at all
func (c *Controller) processStreamPendingPayments() (payments chan Payment, err chan error) {
	payments = make(chan Payment, 1_000)
	err = make(chan error, 1)
	go func() {
		defer close(payments)
		defer close(err)

		err <- c.db.View(func(txn *badger.Txn) (err error) {
			options := badger.DefaultIteratorOptions
			options.Prefix = pendingPrefix
			it := txn.NewIterator(options)
			defer it.Close()

			for it.Rewind(); it.ValidForPrefix(pendingPrefix); it.Next() {
				var payment Payment

				pendingItem := it.Item()

				var id uuid.UUID
				err = pendingItem.Value(func(val []byte) (err error) {
					copy(id[:], val)
					return nil
				})
				if err != nil {
					err = fmt.Errorf("failed to retrieve payment id: %w", err)
					log.Println(err) // We can't return but even then we need to try the others
					continue
				}

				paymentItem, err := txn.Get(PaymentKey(id))
				if err != nil {
					err = fmt.Errorf("failed to retrieve payment: %w", err)
					log.Println(err)
					continue
				}

				err = paymentItem.Value(func(val []byte) (err error) {
					err = payment.FromBytes(val)
					if err != nil {
						return fmt.Errorf("failed to unmarshal to payment: %w", err)
					}
					return nil
				})
				if err != nil {
					err = fmt.Errorf("failed to retrieve payment: %w", err)
					log.Println(err) // We can't return but even then we need to try the others
					continue
				}

				payments <- payment

			}

			return nil
		})
	}()
	return payments, err
}
