package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"anarchy.ttfm/8ball/wallets"
	badger "github.com/dgraph-io/badger/v4"
)

func (c *Controller) getReceiverAddress(ctx context.Context, r Receiver) (address wallets.Address, err error) {
	log.Println("Querying address:", r.Index)

	err = c.wallet.Sync(ctx)
	if err != nil {
		return address, fmt.Errorf("failed to sync wallet: %w", err)
	}

	address, err = c.wallet.Address(ctx, wallets.AddressRequest{Index: r.Index})
	if err != nil {
		return address, fmt.Errorf("failed to retrieve address: %w", err)
	}

	if address.Address != r.Address {
		return address, fmt.Errorf("expecting a different address: %s != %s; did server wallet changed?", r.Address, address.Address)
	}
	return address, nil
}

// This utility function is used for those scenarios in which the payment has changed state
func (c *Controller) savePaymentState(p Payment) (err error) {
	cc, _ := json.MarshalIndent(p, "", "\t")
	log.Println("Saving payment:", string(cc))

	return c.db.Update(func(txn *badger.Txn) (err error) {
		contents := p.Bytes()

		err = txn.Set(PaymentKey(p.Id), contents)
		if err != nil {
			return fmt.Errorf("failed to set new payment at key:m %w", err)
		}
		return nil
	})
}

// This is a utility function that should be called just in case something goes wrong while processing a pending payment
func (c *Controller) deletePendingPayment(p Payment) (err error) {
	log.Println("Deleting pending entry")
	return c.db.Update(func(txn *badger.Txn) (err error) {
		err = txn.Delete(PendingKey(p.Id))
		if err != nil {
			return fmt.Errorf("failed to delete pending key: %w", err)
		}
		return nil
	})
}
