package gateway

import (
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

	address, err := c.wallet.Address(ctx, blockchains.AddressRequest{Index: p.ReceiverIndex})
	if err != nil {
		return fmt.Errorf("failed to retrieve payment address: %w", err)
	}

	switch {
	case address.UnlockedBalance > 0:
		defer func() {
			if err != nil {
				p.Status = StatusError
				p.Error = err.Error()
			}
			c.savePaymentState(p)
		}()

		if !p.FeePayed {
			fee := calculateFee(address.UnlockedBalance, p.Fee)
			// Discount fee and transfer it to the beneficiary
			// Pay to the one running the gateway
			var feeTransfer blockchains.Transfer
			feeTransfer, err = c.wallet.Transfer(ctx, blockchains.TransferRequest{
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
			p.FeePayed = true
			p.FeeTransaction = feeTransfer.Address
		}

		if !p.BeneficiaryPayed {
			// Transfer remaining balance to business
			var beneficiarySweep blockchains.Sweep
			beneficiarySweep, err = c.wallet.SweepAll(ctx, blockchains.SweepRequest{
				SourceIndex: p.ReceiverIndex,
				Destination: p.Beneficiary,
				Priority:    p.Priority,
				UnlockTime:  0,
			})
			if err != nil {
				return fmt.Errorf("failed to sweep remaining contents to destination: %w", err)
			}

			// Confirm destination was payed
			p.BeneficiaryPayed = true
			p.BeneficiaryTransaction = beneficiarySweep.Address
		}

		// If the payment was made successfully
		if address.UnlockedBalance >= p.Amount {
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

// If it expired
// - Delete pending entry
// - Try to transfer the received amount to beneficiary
func (c *Controller) processLivePayment(p Payment) (err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	account, err := c.wallet.Address(ctx, blockchains.AddressRequest{Index: p.ReceiverIndex})
	if err != nil {
		return fmt.Errorf("failed to retrieve payment account: %w", err)
	}
	// Ignore since we don't have received the payment
	if account.UnlockedBalance < p.Amount {
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

	if !p.FeePayed {
		fee := calculateFee(account.UnlockedBalance, p.Fee)
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
		p.FeePayed = true
		p.FeeTransaction = feeTransfer.Address
	}

	if !p.BeneficiaryPayed {
		// Transfer remaining balance to destination
		beneficiarySweep, err := c.wallet.SweepAll(ctx, blockchains.SweepRequest{
			SourceIndex: p.ReceiverIndex,
			Destination: p.Beneficiary,
			Priority:    p.Priority,
			UnlockTime:  0,
		})
		if err != nil {
			return fmt.Errorf("failed to sweep remaining contents to destination: %w", err)
		}

		// Confirm destination was payed
		p.BeneficiaryPayed = true
		p.BeneficiaryTransaction = beneficiarySweep.Address
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
