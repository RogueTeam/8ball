package gateway_test

import (
	"testing"
	"time"

	"anarchy.ttfm/8ball/blockchains"
	"anarchy.ttfm/8ball/blockchains/mock"
	"anarchy.ttfm/8ball/gateway"
	"anarchy.ttfm/8ball/random"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func Test_Integration(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		assertions := assert.New(t)

		wallet := mock.New()

		label1 := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		beneficiary, err := wallet.NewAccount(blockchains.NewAccountRequest{Label: label1})
		assertions.Nil(err, "failed to create beneficiary account")

		options := badger.
			DefaultOptions("").
			WithInMemory(true)
		db, err := badger.Open(options)
		assertions.Nil(err, "failed to open database")
		var config = gateway.Config{
			DB:          db,
			Fee:         1,
			Timeout:     5 * time.Second,
			Beneficiary: beneficiary.Address,
			Wallet:      wallet,
		}
		ctrl := gateway.New(config)
		// t.Logf("Create controller: %+v", ctrl)

		label2 := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		receiver, err := wallet.NewAccount(blockchains.NewAccountRequest{Label: label2})
		assertions.Nil(err, "failed to create dst account")

		payment, err := ctrl.New(receiver.Address, 10_000, blockchains.PriorityHigh)
		assertions.Nil(err, "failed to create payment")
		// t.Logf("Create payment: %+v", payment)

		// Query first
		firstQuery, err := ctrl.Query(payment.Id)
		assertions.Nil(err, "failed to query first payment")

		assertions.Equal(payment.Id, firstQuery.Id, "Don't equal")

		// Pay the dst
		_, err = wallet.Transfer(blockchains.TransferRequest{
			SourceIndex: 0,
			Destination: payment.Receiver,
			Amount:      10_000,
			Priority:    blockchains.PriorityHigh,
			UnlockTime:  0,
		})
		assertions.Nil(err, "failed to transfer to destination")

		// Process
		err = ctrl.Process()
		assertions.Nil(err, "failed to process payments")

		// Verify payment
		secondQuery, err := ctrl.Query(payment.Id)
		assertions.Nil(err, "failed to query payment")

		assertions.Equal(gateway.StatusCompleted, secondQuery.Status, "status don't match")

		// Verify beneficiary received the fee
		beneficiaryAccount, err := wallet.Account(blockchains.AccountRequest{Index: beneficiary.Index})
		assertions.Nil(err, "failed to query beneficiary account")
		assertions.Equal(uint64(100), beneficiaryAccount.UnlockedBalance, "invalid beneficiary balance")
		// Verify Destination received the rest of the money
		receiverAccount, err := wallet.Account(blockchains.AccountRequest{Index: receiver.Index})
		assertions.Nil(err, "failed to query receiver account")
		assertions.NotZero(receiverAccount.UnlockedBalance, "invalid receiver balance")
	})
}
