package gateway

import (
	"context"
	"fmt"
	"time"

	"anarchy.ttfm/8ball/wallets"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

type Receive struct {
	Address  string
	Amount   uint64
	Priority wallets.Priority
}

func (c *Controller) validateReceive(ctx context.Context, r *Receive) (err error) {
	if r.Amount < c.minAmount {
		return fmt.Errorf("amount should be greater or equal than: %d", c.minAmount)
	}
	if r.Amount > c.maxAmount {
		return fmt.Errorf("amount should be less or equal than: %d", c.maxAmount)
	}

	err = r.Priority.Validate()
	if err != nil {
		return fmt.Errorf("invalid priority: %w", err)
	}

	err = c.wallet.ValidateAddress(ctx, wallets.ValidateAddressRequest{Address: r.Address})
	if err != nil {
		return fmt.Errorf("failed to validate address: %w", err)
	}
	return nil
}

// Creates a new payment address based on the passed crypto currency and amount expected to receive
// It uses the default timeout in order to prevent infinite entries
// id is the id to be used for future checks
// fee is the percentage to be discounted from the entire transaction
func (c *Controller) Receive(ctx context.Context, req *Receive) (payment Payment, err error) {
	err = c.validateReceive(ctx, req)
	if err != nil {
		return payment, fmt.Errorf("failed to validate request: %w", err)
	}

	err = c.db.Update(func(txn *badger.Txn) (err error) {
		payment = Payment{
			Id:         uuid.New(),
			Priority:   req.Priority,
			Amount:     req.Amount,
			Expiration: time.Now().Add(c.timeout),
			Fee: Fee{
				Status:     StatusPending,
				Percentage: c.feePercentage,
				Address:    c.address,
			},
			Beneficiary: Beneficiary{
				Status:  StatusPending,
				Address: req.Address,
			},
		}

		// Prepare new entry
		receiver, err := c.wallet.NewAddress(ctx, wallets.NewAddressRequest{Label: payment.Id.String()})
		if err != nil {
			return fmt.Errorf("failed to prepare receiver address: %w", err)
		}

		payment.Receiver = Receiver{
			Address: receiver.Address,
			Index:   receiver.Index,
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
