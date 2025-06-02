package controller

import (
	"encoding/json"
	"fmt"
	"log"

	"anarchy.ttfm.onion/gateway/blockchains"
	badger "github.com/dgraph-io/badger/v4"
)

var pendingPrefix = []byte("/pending/")

// Process is a function that goes over all pending payments and checks if the payment was executed
func (c *Controller) Process() (err error) {
	err = c.db.Update(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.IteratorOptions{
			Prefix:         pendingPrefix,
			PrefetchSize:   1_000,
			PrefetchValues: true,
		})
		defer it.Close()

		for ; it.ValidForPrefix(pendingPrefix); it.Next() {
			var payment Payment

			item := it.Item()

			err = item.Value(func(val []byte) (err error) {
				err = json.Unmarshal(val, &payment)
				if err != nil {
					return fmt.Errorf("failed to unmarshal to payment: %w", err)
				}
				return nil
			})
			if err != nil {
				err = fmt.Errorf("failed to retrieve transaction value: %w", err)
				log.Println(err) // We can't return but even then we need to try the others
				continue
			}

			// TODO: Process payment
			wallet, found := c.wallets[payment.Currency]
			if !found {
				err = fmt.Errorf("%w: %s", ErrNoHandlerForCurrency, payment.Currency)
				log.Println(err)
				continue
			}

			balance, err := wallet.AddressBalance(blockchains.AddressBalanceRequest{Address: payment.ReceiverAddress})
			if err != nil {
				err = fmt.Errorf("failed to retrieve middle address balance: %w", err)
				log.Println(err)
				continue
			}

			if balance.Unlocked < payment.Amount {
				log.Println("balance is not ready yet")
				continue
			}

			// TODO: Do all the logic
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}
	return nil
}
