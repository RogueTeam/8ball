package gateway

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"anarchy.ttfm/8ball/blockchains"
	"anarchy.ttfm/8ball/utils"
	badger "github.com/dgraph-io/badger/v4"
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

				item := it.Item()

				err = item.Value(func(val []byte) (err error) {
					err = payment.FromBytes(val)
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

				payments <- payment

			}

			return nil
		})
	}()
	return payments, err
}

// This is a utility function that should be called just in case something goes wrong while processing a pending payment
func (c *Controller) deletePendingPayment(p Payment) {
	err := c.db.Update(func(txn *badger.Txn) (err error) {
		pendingKey := PendingKey(p.Id)
		err = txn.Delete([]byte(pendingKey))
		if err != nil {
			return fmt.Errorf("failed to delete pending key: %w", err)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// This utility function is used for those scenarios in which the payment has changed state
func (c *Controller) savePaymentState(p Payment) {
	err := c.db.Update(func(txn *badger.Txn) (err error) {
		contents := p.Bytes()

		paymentKey := PaymentKey(p.Id)

		err = txn.Set([]byte(paymentKey), contents)
		if err != nil {
			return fmt.Errorf("failed to set new payment at key:m %w", err)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func calculateFee(amount, feePercentage uint64) (fee uint64) {
	return amount * feePercentage / 100
}

// If it expired
// - Delete pending entry
// - Try to transfer the received amount to beneficiary
func (c *Controller) processExpiredPayment(p Payment) (err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	err = c.wallet.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync wallet: %w", err)
	}

	address, err := c.wallet.Address(ctx, blockchains.AddressRequest{Index: p.ReceiverIndex})
	if err != nil {
		return fmt.Errorf("failed to retrieve payment address: %w", err)
	}

	switch {
	case address.UnlockedBalance > 0:
		if address.UnlockedBalance < address.Balance {
			// Address has the more funds but they are not ready yet
			return nil
		}

		defer func() {
			if err != nil {
				p.Status = StatusError
				p.Error = err.Error()
			}
			c.savePaymentState(p)
		}()

		if p.FeeTransaction != "" {
			feePayed, err := c.transactionCompleted(ctx, p.FeeTransaction)
			if err != nil {
				return fmt.Errorf("failed to check if fee was paid: %w", err)
			}

			if !feePayed {
				return nil
			}
			p.IsFeePayed = true
		}

		err = c.payFee(ctx, &p)
		if err != nil {
			return fmt.Errorf("failed to pay fee: %w", err)
		}

		err = c.payBeneficiary(ctx, &p)
		if err != nil {
			return fmt.Errorf("failed to pay beneficiary: %w", err)
		}

		beneficiaryPayed, err := c.transactionCompleted(ctx, p.BeneficiaryTransaction)
		if err != nil {
			return fmt.Errorf("failed to check if beneficiary was paid: %w", err)
		}

		if !beneficiaryPayed {
			return nil
		}
		p.IsBeneficiaryPayed = true

		// If the payment was made successfully
		if p.PayedFee == calculateFee(p.Amount, p.Fee) {
			p.Status = StatusCompleted
		} else {
			// There is money but not enought
			// Payment expired, we can make a profit from the unlocked balance
			p.Status = StatusPartiallyCompleted
		}

		c.deletePendingPayment(p)
		return nil
	case address.Balance > 0:
		// Money is there but not available yet
		return nil
	default:
		defer func() {
			c.deletePendingPayment(p)

			if err != nil {
				p.Status = StatusError
				p.Error = err.Error()
			}
			c.savePaymentState(p)
		}()

		// Expired and no money was found
		p.Status = StatusExpired
		return nil
	}
}

func (c *Controller) payFee(ctx context.Context, p *Payment) (err error) {
	if p.IsFeePayed || p.Fee == 0 || p.FeeTransaction != "" {
		return
	}

	err = c.wallet.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync wallet: %w", err)
	}

	address, err := c.wallet.Address(ctx, blockchains.AddressRequest{Index: p.ReceiverIndex})
	if err != nil {
		return fmt.Errorf("failed to retrieve payment address: %w", err)
	}

	fee := calculateFee(address.UnlockedBalance, p.Fee)

	// Discount fee and transfer it to the beneficiary
	feeTransfer, err := c.wallet.Transfer(ctx, blockchains.TransferRequest{
		SourceIndex: p.ReceiverIndex,
		Destination: c.beneficiary,
		Amount:      fee,
		Priority:    blockchains.PriorityHigh,
		UnlockTime:  0,
	})
	if err != nil {
		return fmt.Errorf("failed to transfer to beneficiary: %w", err)
	}

	// Confirm fee was payed
	p.PayedFee = fee
	p.FeeTransaction = feeTransfer.Address

	return
}

func (c *Controller) payBeneficiary(ctx context.Context, p *Payment) (err error) {
	if p.IsBeneficiaryPayed || p.BeneficiaryTransaction != "" {
		return
	}

	err = c.wallet.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync wallet: %w", err)
	}

	address, err := c.wallet.Address(ctx, blockchains.AddressRequest{Index: p.ReceiverIndex})
	if err != nil {
		return fmt.Errorf("failed to retrieve payment address: %w", err)
	}

	// Transfer remaining balance to destination
	beneficiarySweep, err := c.wallet.SweepAll(ctx, blockchains.SweepRequest{
		SourceIndex: p.ReceiverIndex,
		Destination: p.Beneficiary,
		Priority:    p.Priority,
		UnlockTime:  0,
	})
	if err != nil {
		return fmt.Errorf("failed to sweep remaining contents to destination: %w -> %s", err, address.String())
	}

	// Confirm destination was payed
	p.PayedBeneficiary = beneficiarySweep.Amount
	p.BeneficiaryTransaction = beneficiarySweep.Address

	return nil
}

func (c *Controller) transactionCompleted(ctx context.Context, address string) (payed bool, err error) {
	err = c.wallet.Sync(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to sync wallet: %w", err)
	}

	tx, err := c.wallet.Transaction(ctx, blockchains.TransactionRequest{TransactionId: address})
	if err != nil {
		return false, fmt.Errorf("failed to query transaction: %w", err)
	}

	switch tx.Status {
	case blockchains.TransactionStatusCompleted:
		return true, nil
	case blockchains.TransactionStatusPending:
		return false, nil
	default: // blockchains.TransactionStatusFailed
		return false, errors.New("transaction failed")
	}
}

// If it expired
// - Delete pending entry
// - Try to transfer the received amount to beneficiary
func (c *Controller) processLivePayment(p Payment) (err error) {
	log.Println("Processing live payment")
	ctx, cancel := utils.NewContext()
	defer cancel()

	err = c.wallet.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync wallet: %w", err)
	}

	address, err := c.wallet.Address(ctx, blockchains.AddressRequest{Index: p.ReceiverIndex})
	if err != nil {
		return fmt.Errorf("failed to retrieve payment address: %w", err)
	}

	log.Println(address.String())
	// Ignore since we don't have received the payment
	if !p.IsFeePayed && address.UnlockedBalance < p.Amount {
		return nil
	}

	if address.UnlockedBalance < address.Balance {
		// Address has the more funds but they are not ready yet
		return nil
	}

	// We have received the payment. Lets distribute it between participants
	defer func() {
		if err != nil {
			p.Status = StatusError
			p.Error = err.Error()
		}
		c.savePaymentState(p)
	}()

	err = c.payFee(ctx, &p)
	if err != nil {
		return fmt.Errorf("failed to pay fee: %w", err)
	}

	if p.FeeTransaction != "" {
		feePayed, err := c.transactionCompleted(ctx, p.FeeTransaction)
		if err != nil {
			return fmt.Errorf("failed to check if fee was paid: %w", err)
		}

		if !feePayed {
			return nil
		}

		p.IsFeePayed = true
	}

	err = c.payBeneficiary(ctx, &p)
	if err != nil {
		return fmt.Errorf("failed to pay beneficiary: %w", err)
	}

	beneficiaryPayed, err := c.transactionCompleted(ctx, p.BeneficiaryTransaction)
	if err != nil {
		return fmt.Errorf("failed to check if beneficiary was paid: %w", err)
	}
	p.IsBeneficiaryPayed = true

	if !beneficiaryPayed {
		return nil
	}

	c.deletePendingPayment(p)
	p.Status = StatusCompleted
	return nil
}

func (c *Controller) processPayment(now time.Time, p Payment) (err error) {
	if p.Expiration.Before(now) {
		err = c.processExpiredPayment(p)
		if err != nil {
			return fmt.Errorf("failed to process expired payment: %w", err)
		}
		return nil
	}

	err = c.processLivePayment(p)
	if err != nil {
		return fmt.Errorf("failed to process live payment: %w", err)
	}
	return nil
}

const MaxConcurrentJobs = 1_000

// Process is a function that goes over all pending payments and checks if the payment was executed
func (c *Controller) Process() (err error) {
	now := time.Now()

	payments, errChan := c.processStreamPendingPayments()
	defer utils.ConsumeChannel(payments)
	defer utils.ConsumeChannel(errChan)

	var jobs = utils.NewJobPool(MaxConcurrentJobs)
	var wg sync.WaitGroup
	for payment := range payments {
		jobs.Get()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer jobs.Put()

			err := c.processPayment(now, payment)
			if err != nil {
				log.Println("failed to process payment")
			}
		}()
	}

	wg.Wait()

	err = <-errChan
	if err != nil {
		return fmt.Errorf("failed to retrieve jobs: %w", err)
	}
	return nil
}
