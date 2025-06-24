package gateway

import (
	"fmt"
	"log"
	"sync"

	"anarchy.ttfm/8ball/utils"
	"anarchy.ttfm/8ball/wallets"
)

func (c *Controller) processFee(p Payment) (err error) {
	// cc, _ := json.MarshalIndent(p, "", "\t")
	// log.Println("Processing fee:", string(cc))

	ctx, cancel := utils.NewContext()
	defer cancel()

	address, err := c.getReceiverAddress(ctx, p.Receiver)
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}

	if address.Balance == 0 {
		return nil
	}

	// We can wait for the rest of the money to arrive
	if address.Balance > address.UnlockedBalance {
		return nil
	}

	sweep, err := c.wallet.SweepAll(ctx, wallets.SweepRequest{
		SourceIndex: p.Receiver.Index,
		Destination: p.Fee.Address,
		Priority:    p.Priority,
		UnlockTime:  0,
	})
	if err != nil {
		err = fmt.Errorf("failed to transfer funds: %w", err)
		p.Fee.SetError(err)

		err = c.savePaymentState(p)
		if err != nil {
			return fmt.Errorf("failed to set save payment: %w", err)
		}
		return err
	}

	p.Fee.Payed = sweep.Amount
	p.Fee.Transaction = sweep.Address
	p.Fee.Status = StatusCompleted

	err = c.savePaymentState(p)
	if err != nil {
		return fmt.Errorf("failed to set save payment: %w", err)
	}
	err = c.deleteKey(FeeKey(p.Id))
	if err != nil {
		return fmt.Errorf("failed to delete pending payment entry: %w", err)
	}
	return nil
}

func (c *Controller) ProcessPendingFees() (processed uint64, err error) {
	payments, errChan := c.streamPayments(feePrefixBytes)
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

			err := c.processFee(payment)
			if err != nil {
				log.Println("failed to process payment:", payment.Id, err)
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
