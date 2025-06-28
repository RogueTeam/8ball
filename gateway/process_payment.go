package gateway

import (
	"fmt"
	"log"
	"sync"
	"time"

	"anarchy.ttfm/8ball/utils"
	"anarchy.ttfm/8ball/wallets"
)

func (c *Controller) processPayment(p Payment) (err error) {
	now := time.Now()

	// cc, _ := json.MarshalIndent(p, "", "\t")
	// log.Println("Processing payment:", string(cc))

	ctx, cancel := utils.NewContext()
	defer cancel()

	address, err := c.getReceiverAddress(ctx, p.Receiver)
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}

	// We can wait for the rest of the money to arrive
	if address.Balance > address.UnlockedBalance {
		return nil
	}

	// Since the it has not been expired and there is not the totallity of the money
	// wec an wait more
	if now.Before(p.Expiration) && address.UnlockedBalance < p.Amount {
		return nil
	}

	// At this point the transaction could be in any of these states:
	// - have already expired with funds or empty
	// - still live with funds

	// The account was found with funds so:
	// - If it is live. Funds are complete
	// - If expired it may have incomplete funds
	if address.UnlockedBalance > 0 {
		transfer, err := c.wallet.Transfer(ctx, wallets.TransferRequest{
			SourceIndex: p.Receiver.Index,
			Destination: p.Beneficiary.Address,
			Amount:      address.UnlockedBalance - calculateFee(address.UnlockedBalance, p.Fee.Percentage),
			Priority:    p.Priority,
			UnlockTime:  0,
		})
		if err != nil {
			err = fmt.Errorf("failed to transfer funds: %w", err)
			p.Beneficiary.SetError(err)

			err = c.savePaymentState(p)
			if err != nil {
				return fmt.Errorf("failed to set save payment: %w", err)
			}
			return err
		}

		p.Beneficiary.Payed = transfer.Amount
		p.Beneficiary.Transaction = transfer.Address

		if address.UnlockedBalance >= p.Amount {
			p.Beneficiary.Status = StatusCompleted
		} else {
			p.Beneficiary.Status = StatusPartiallyCompleted
		}
		err = c.savePendingFee(p)
		if err != nil {
			return fmt.Errorf("failed to save pending fee: %w", err)
		}
	} else {
		p.Beneficiary.Status = StatusExpired
		p.Fee.Status = StatusExpired
	}

	err = c.savePaymentState(p)
	if err != nil {
		return fmt.Errorf("failed to set save payment: %w", err)
	}
	err = c.deleteKey(PendingKey(p.Id))
	if err != nil {
		return fmt.Errorf("failed to delete pending payment entry: %w", err)
	}
	return nil
}

const MaxConcurrentJobs = 1_000

// ProcessPendingPayments is a function that goes over all pending payments and checks if the payment was executed
func (c *Controller) ProcessPendingPayments() (processed uint64, err error) {
	payments, errChan := c.streamPayments(pendingPrefixBytes)
	defer utils.ConsumeChannel(payments)
	defer utils.ConsumeChannel(errChan)

	var jobs = utils.NewJobPool(MaxConcurrentJobs)
	var wg sync.WaitGroup
	for payment := range payments {
		processed++
		jobs.Get()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer jobs.Put()

			err := c.processPayment(payment)
			if err != nil {
				log.Printf("failed to process beneficiary payment: %v: %v", payment.Id, err)
			}
		}()
	}

	wg.Wait()

	err = <-errChan
	if err != nil {
		return processed, fmt.Errorf("failed to retrieve jobs: %w", err)
	}
	return processed, nil
}
