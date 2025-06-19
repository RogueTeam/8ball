package testsuite

import (
	"encoding/json"
	"testing"
	"time"

	"anarchy.ttfm/8ball/blockchains"
	"anarchy.ttfm/8ball/gateway"
	"anarchy.ttfm/8ball/random"
	"anarchy.ttfm/8ball/utils"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

// DataGenerator defines an interface for test data generation.
type DataGenerator interface {
	// TransferAmount returns the amount to send for a transfer.
	TransferAmount() (funds uint64)
}

// Test runs a comprehensive suite of tests for any Wallet implementation.
func Test(t *testing.T, wallet blockchains.Wallet, gen DataGenerator) {
	t.Run("Succeed", func(t *testing.T) {
		assertions := assert.New(t)

		ctx, cancel := utils.NewContext()
		defer cancel()

		label1 := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		gatewayAddress, err := wallet.NewAddress(ctx, blockchains.NewAddressRequest{Label: label1})
		assertions.Nil(err, "failed to create beneficiary account")

		options := badger.
			DefaultOptions("").
			WithInMemory(true)
		db, err := badger.Open(options)
		assertions.Nil(err, "failed to open database")
		var config = gateway.Config{
			DB:          db,
			Fee:         10,
			Timeout:     24 * time.Hour,
			Beneficiary: gatewayAddress.Address,
			Wallet:      wallet,
		}
		ctrl := gateway.New(config)
		// t.Logf("Create controller: %+v", ctrl)

		label2 := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		businessAddress, err := wallet.NewAddress(ctx, blockchains.NewAddressRequest{Label: label2})
		assertions.Nil(err, "failed to create dst account")

		payment, err := ctrl.Receive(gateway.Receive{
			Destination: businessAddress.Address, Amount: gen.TransferAmount(), Priority: blockchains.PriorityHigh,
		})
		assertions.Nil(err, "failed to create payment")
		// t.Logf("Create payment: %+v", payment)

		// Query first
		firstQuery, err := ctrl.Query(payment.Id)
		assertions.Nil(err, "failed to query first payment")

		assertions.Equal(payment.Id, firstQuery.Id, "Don't equal")

		// Pay the dst
		clientTransfer, err := wallet.Transfer(ctx, blockchains.TransferRequest{
			SourceIndex: 0,
			Destination: payment.Receiver,
			Amount:      gen.TransferAmount(),
			Priority:    blockchains.PriorityHigh,
			UnlockTime:  0,
		})
		assertions.Nil(err, "failed to transfer to destination")

		t.Log("[*] Waiting for payment be effective")
		var paymentEffective bool
		for try := range 3_600 {
			t.Log("[*] Try: ", try+1)
			tx, err := wallet.Transaction(ctx, blockchains.TransactionRequest{TransactionId: clientTransfer.Address})
			assertions.Nil(err, "failed to retrieve transaction")

			if tx.Status == blockchains.TransactionStatusCompleted {
				paymentEffective = true
				break
			}
			time.Sleep(time.Second)
		}
		assertions.True(paymentEffective, "payment not made effective")
		t.Log("[+] Transaction made effective")

		// Process
		t.Log("[*] Processing payments")
		var paymentProcessed bool
		for try := range 3_600 {
			t.Log("[*] Try: ", try+1)

			err = ctrl.Process()
			assertions.Nil(err, "failed to process payments")

			// Verify payment
			paymentLatest, err := ctrl.Query(payment.Id)
			assertions.Nil(err, "failed to query payment")

			contents, _ := json.MarshalIndent(paymentLatest, "", "\t")
			t.Log(string(contents))

			if paymentLatest.Status == gateway.StatusCompleted {
				paymentProcessed = true
				break
			}
			time.Sleep(time.Second)
		}

		assertions.True(paymentProcessed, "payment not processed")

		t.Log("[*] Waiting for beneficiary be credited")
		var beneficiaryPayed bool
		for try := range 3_600 {
			t.Log("[*] Try: ", try+1)
			beneficiaryLatestAddress, err := wallet.
				Address(ctx, blockchains.AddressRequest{Index: gatewayAddress.Index})
			assertions.Nil(err, "failed to query beneficiary address")

			if beneficiaryLatestAddress.UnlockedBalance > 0 {
				beneficiaryPayed = true
				break
			}
			time.Sleep(time.Second)
		}
		assertions.True(beneficiaryPayed, "beneficiary not payed")
		t.Log("[+] Beneficiary payed")

		t.Log("[*] Waiting for bussines be credited")
		var bussinessPayed bool
		for try := range 3_600 {
			t.Log("[*] Try: ", try+1)

			businessLatestAddress, err := wallet.Address(ctx, blockchains.AddressRequest{Index: businessAddress.Index})
			assertions.Nil(err, "failed to query bussiness addrss")

			if businessLatestAddress.UnlockedBalance > 0 {
				bussinessPayed = true
				break
			}
			time.Sleep(time.Second)
		}
		assertions.True(bussinessPayed, "business not payed")
		t.Log("[+] Bussiness payed")
	})
}
