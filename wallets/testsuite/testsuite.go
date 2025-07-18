package testsuite

import (
	"testing"
	"time"

	"github.com/RogueTeam/8ball/random"
	"github.com/RogueTeam/8ball/utils"
	wallets "github.com/RogueTeam/8ball/wallets"
	"github.com/stretchr/testify/assert"
)

// DataGenerator defines an interface for test data generation.
type DataGenerator interface {
	// TransferAmount returns the amount to send for a transfer.
	TransferAmount() (funds uint64)
}

// Test runs a comprehensive suite of tests for any Wallet implementation.
func Test(t *testing.T, w wallets.Wallet, gen DataGenerator) {

	t.Run("Initial Address 0 State", func(t *testing.T) {
		t.Parallel()

		assertions := assert.New(t)

		ctx, cancel := utils.NewContextWithTimeout(time.Hour)
		defer cancel()

		err := w.Sync(ctx, true)
		assertions.Nil(err, "failed to sync")

		// Verify Address 0 exists and has zero balance initially
		address0, err := w.Address(ctx, wallets.AddressRequest{Index: 0})
		assertions.Nil(err, "failed to retrieve initial address 0 balance")
		assertions.Equal(uint64(0), address0.Index, "Address 0 should have index 0")

		t.Logf("Initial Address 0: %+v", address0)
	})

	t.Run("NewAddress", func(t *testing.T) {
		t.Parallel()

		assertions := assert.New(t)

		ctx, cancel := utils.NewContextWithTimeout(time.Hour)
		defer cancel()

		err := w.Sync(ctx, true)
		assertions.Nil(err, "failed to sync")

		label := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		address, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: label})
		assertions.Nil(err, "failed to create new address")
		assertions.NotNil(address.Address, "new address should have an address")
		assertions.Greater(address.Index, uint64(0), "new address index should be greater than 0")
		assertions.Equal(uint64(0), address.Balance, "new address should have zero balance initially")
		assertions.Equal(uint64(0), address.UnlockedBalance, "new address should have zero unlocked balance initially")

		t.Logf("Created new address: %+v", address)

		// Verify the new address can be retrieved
		retrievedAddress, err := w.Address(ctx, wallets.AddressRequest{Index: address.Index})
		assertions.Nil(err, "failed to get newly created address")
		assertions.Equal(address, retrievedAddress, "retrieved address should match created address")
		t.Logf("Retrieved new address: %+v", retrievedAddress)

		// Test creating another address to check index increment
		label2 := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		address2, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: label2})
		assertions.Nil(err, "failed to create second new address")
		assertions.Less(address.Index, address2.Index, "second address index should be incremented")
	})

	t.Run("ValidateAddress", func(t *testing.T) {
		t.Parallel()

		assertions := assert.New(t)

		ctx, cancel := utils.NewContextWithTimeout(time.Hour)
		defer cancel()

		err := w.Sync(ctx, true)
		assertions.Nil(err, "failed to sync")

		label := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		address, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: label})
		assertions.Nil(err, "failed to create new address")

		// Test with a valid address
		err = w.ValidateAddress(ctx, wallets.ValidateAddressRequest{Address: address.Address})
		assertions.Nil(err, "failed to validate address")
	})

	t.Run("Transfer", func(t *testing.T) {
		t.Parallel()

		t.Run("To Internal Address", func(t *testing.T) {
			t.Parallel()

			assertions := assert.New(t)

			ctx, cancel := utils.NewContextWithTimeout(time.Hour)
			defer cancel()

			err := w.Sync(ctx, true)
			assertions.Nil(err, "failed to sync")

			firstAddress0, err := w.Address(ctx, wallets.AddressRequest{Index: 0})
			assertions.Nil(err, "failed to get current address 0 balance")
			t.Logf("Address 0 balance before transfers: Balance:%d ; UnlockedBalance:%d", firstAddress0.Balance, firstAddress0.UnlockedBalance)

			dst, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new address for internal transfer")

			transfer, err := w.Transfer(ctx, wallets.TransferRequest{
				SourceIndex: 0,
				Destination: dst.Address,
				Amount:      gen.TransferAmount(),
				Priority:    wallets.PriorityHigh,
				UnlockTime:  0,
			})
			if !assertions.Nil(err, "failed to transfer funds to internal address") {
				return
			}
			assertions.NotEmpty(transfer.Address, "transfer should have a transaction address")
			assertions.Equal(uint64(0), transfer.SourceIndex, "source index should be 0")
			assertions.Equal(dst.Address, transfer.Destination, "destination address should match")
			t.Logf("Transfer to internal address: %+v", transfer)

			// Verify balances after internal transfer
			address0After, err := w.Address(ctx, wallets.AddressRequest{Index: 0})
			assertions.Nil(err, "failed to get address 0 balance after internal transfer")
			assertions.Less(address0After.Balance, firstAddress0.Balance, "Address 0 balance should be reduced by transfer amount")

			var tx wallets.Transaction
			var found bool
			for try := range 3_600 {
				t.Log("[*] Syncing")
				err = w.Sync(ctx, true)
				assertions.Nil(err, "failed to sync")
				t.Log("[*] Synced")

				t.Log("[*] Checking Transaction: Attempt ", try+1)

				tx, err = w.Transaction(ctx, wallets.TransactionRequest{SourceIndex: transfer.SourceIndex, TransactionId: transfer.Address})
				if !assertions.Nil(err, "failed to get destination address balance after internal transfer") {
					return
				}

				if tx.Status != wallets.TransactionStatusPending {
					found = true
					break
				}
				time.Sleep(time.Second)
			}
			assertions.True(found, "Destination address balance should increase by net transfer amount")
			assertions.Equal(wallets.TransactionStatusCompleted, tx.Status, "status doesn't match")
			t.Log("[+] Transaction found")
		})

		t.Run("Insufficient Funds", func(t *testing.T) {
			t.Parallel()

			assertions := assert.New(t)

			ctx, cancel := utils.NewContextWithTimeout(time.Hour)
			defer cancel()

			err := w.Sync(ctx, true)
			assertions.Nil(err, "failed to sync")

			currentAddress0, err := w.Address(ctx, wallets.AddressRequest{Index: 0})
			assertions.Nil(err, "failed to get current address 0 balance")
			t.Logf("Address 0 balance before transfers: %d", currentAddress0.Balance)

			dst, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new address for internal transfer")

			// Attempt to transfer more than available in Address 0
			insufficientAmount := currentAddress0.UnlockedBalance + 1000 // More than current balance

			_, err = w.Transfer(ctx, wallets.TransferRequest{
				SourceIndex: 0,
				Destination: dst.Address,
				Amount:      insufficientAmount,
				Priority:    wallets.PriorityHigh,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "transfer should fail due to insufficient funds")
			t.Logf("Attempted transfer with insufficient funds, got expected error: %v", err)

		})

		t.Run("Zero Amount Transfer", func(t *testing.T) {
			t.Parallel()

			assertions := assert.New(t)

			ctx, cancel := utils.NewContextWithTimeout(time.Hour)
			defer cancel()

			err := w.Sync(ctx, true)
			assertions.Nil(err, "failed to sync")

			currentAddress0, err := w.Address(ctx, wallets.AddressRequest{Index: 0})
			assertions.Nil(err, "failed to get current address 0 balance")
			t.Logf("Address 0 balance before transfers: %d", currentAddress0.Balance)

			dst, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new address for internal transfer")

			_, err = w.Transfer(ctx, wallets.TransferRequest{
				SourceIndex: 0,
				Destination: dst.Address,
				Amount:      0, // Zero amount
				Priority:    wallets.PriorityHigh,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "transfer should fail for zero amount")
			t.Logf("Attempted transfer with zero amount, got expected error: %v", err)
		})

		t.Run("Transfer from Non-Existent Address", func(t *testing.T) {
			t.Parallel()

			assertions := assert.New(t)

			ctx, cancel := utils.NewContextWithTimeout(time.Hour)
			defer cancel()

			err := w.Sync(ctx, true)
			assertions.Nil(err, "failed to sync")

			dst, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new address for internal transfer")

			_, err = w.Transfer(ctx, wallets.TransferRequest{
				SourceIndex: ^uint64(0),
				Destination: dst.Address,
				Amount:      gen.TransferAmount(),
				Priority:    wallets.PriorityHigh,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "transfer from non-existent address should fail")
			t.Logf("Attempted transfer from non-existent address, got expected error: %v", err)
		})
	})

	t.Run("SweepAll", func(t *testing.T) {
		t.Parallel()

		t.Run("Successful Sweep", func(t *testing.T) {
			t.Parallel()

			assertions := assert.New(t)

			ctx, cancel := utils.NewContextWithTimeout(time.Hour)
			defer cancel()

			err := w.Sync(ctx, true)
			assertions.Nil(err, "failed to sync")

			sourceLabel := "sweep_source" + random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
			// Create a new address and fund it specifically for this sweep test
			sweepSourceAddr, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: sourceLabel})
			if !assertions.Nil(err, "failed to create sweep source address") {
				return
			}

			// Transfer to Source address
			firstTransfer, err := w.Transfer(ctx, wallets.TransferRequest{
				SourceIndex: 0,
				Destination: sweepSourceAddr.Address,
				Amount:      gen.TransferAmount(),
				Priority:    wallets.PriorityHigh,
				UnlockTime:  0,
			})
			if !assertions.Nil(err, "failed to transfer testing amount") {
				return
			}

			t.Log("[*] Waiting for transfer be available")

			var transferTx wallets.Transaction
			var validSourceAddressBalance bool
			for try := range 3_600 {
				t.Log("[*] Syncing")
				err = w.Sync(ctx, true)
				assertions.Nil(err, "failed to sync")
				t.Log("[*] Synced")

				t.Log("[*] Checking Transaction: Attempt ", try+1)
				transferTx, err = w.Transaction(ctx, wallets.TransactionRequest{SourceIndex: firstTransfer.SourceIndex, TransactionId: firstTransfer.Address})
				if !assertions.Nil(err, "failed to get destination address balance after internal transfer") {
					return
				}

				if transferTx.Status != wallets.TransactionStatusPending {
					validSourceAddressBalance = true
					break
				}

				time.Sleep(time.Second)
			}
			assertions.True(validSourceAddressBalance, "source address never received balance")
			if !assertions.Equal(wallets.TransactionStatusCompleted, transferTx.Status, "invalid status") {
				return
			}
			t.Log("[+] Transfer received at address", sweepSourceAddr.Index)

			t.Log("[*] Syncing")
			err = w.Sync(ctx, true)
			assertions.Nil(err, "failed to sync")
			t.Log("[*] Synced")

			dstLabel := "sweep_destination" + random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
			sweepDstAddr, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: dstLabel})
			assertions.Nil(err, "failed to create sweep destination address")

			// Sweep the entire source to dst
			var sweepSucceed bool
			var sweep wallets.Sweep

			t.Log("[*] Waiting for successful sweep")
			for range 3_600 {
				t.Log("[*] Syncing")
				err = w.Sync(ctx, true)
				assertions.Nil(err, "failed to sync")
				t.Log("[*] Synced")

				sourceAddrLast, err := w.Address(ctx, wallets.AddressRequest{Index: sweepSourceAddr.Index})
				assertions.Nil(err, "failed to retrieve source address")
				t.Log("[*] Balance:", sourceAddrLast.Balance)
				t.Log("[*] Unlocked Balance:", sourceAddrLast.UnlockedBalance)

				if sourceAddrLast.UnlockedBalance == 0 {
					time.Sleep(time.Second)
					continue
				}
				t.Log("[*] Sweep Attempt ")
				sweep, err = w.SweepAll(ctx, wallets.SweepRequest{
					SourceIndex: sweepSourceAddr.Index,
					Destination: sweepDstAddr.Address,
					Priority:    wallets.PriorityHigh,
					UnlockTime:  0,
				})
				sweepSucceed = assertions.Nil(err, "failed to sweep funds")
				break
			}

			if !assertions.True(sweepSucceed, "failed to sweep all funds") {
				return
			}
			t.Log("[+] Sweep succeed")

			assertions.NotEmpty(sweep.Address, "sweep should return one transaction hash")
			assertions.NotEmpty(sweep.Amount, "sweep should return one amount")
			assertions.NotEmpty(sweep.Fee, "sweep should return one fee")
			assertions.Equal(sweepSourceAddr.Index, sweep.SourceIndex, "sweep source index should match")
			assertions.Equal(sweepDstAddr.Address, sweep.Destination, "sweep destination address should match")
			t.Logf("Successful sweep: %+v", sweep)

			// Verify balances after sweep
			sourceAddrAfter, err := w.Address(ctx, wallets.AddressRequest{Index: sweepSourceAddr.Index})
			assertions.Nil(err, "failed to get source address balance after sweep")
			assertions.Equal(uint64(0), sourceAddrAfter.Balance, "source address balance should be zero after sweep")
			assertions.Equal(uint64(0), sourceAddrAfter.UnlockedBalance, "source address unlocked balance should be zero after sweep")

			var sweepTx wallets.Transaction
			var dstAddressBalanceValid bool
			for try := range 3_600 {
				t.Log("[*] Checking Transaction: Attempt ", try+1)

				sweepTx, err = w.Transaction(ctx, wallets.TransactionRequest{SourceIndex: sweep.SourceIndex, TransactionId: sweep.Address})
				assertions.Nil(err, "failed to get destination address balance after internal transfer")

				if sweepTx.Status != wallets.TransactionStatusPending {
					dstAddressBalanceValid = true
					break
				}
				time.Sleep(time.Second)
			}
			assertions.True(dstAddressBalanceValid, "destination address balance should increase by net swept amount")
			assertions.Equal(wallets.TransactionStatusCompleted, sweepTx.Status, "invalid status")
			t.Log("[+] Transaction found")
		})

		t.Run("Sweep Empty Address", func(t *testing.T) {
			t.Parallel()

			assertions := assert.New(t)

			ctx, cancel := utils.NewContextWithTimeout(time.Hour)
			defer cancel()

			err := w.Sync(ctx, true)
			assertions.Nil(err, "failed to sync")

			// Create a new address with zero balance
			emptyAddr, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: "empty_address"})
			assertions.Nil(err, "failed to create empty address")

			sweepDstAddr, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: "sweep_destination"})
			assertions.Nil(err, "failed to create sweep destination address")

			_, err = w.SweepAll(ctx, wallets.SweepRequest{
				SourceIndex: emptyAddr.Index,
				Destination: sweepDstAddr.Address,
				Priority:    wallets.PriorityHigh,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "sweeping an empty address should fail")
			t.Logf("Attempted sweep of empty address, got expected error: %v", err)
		})

		t.Run("Sweep from Non-Existent Address", func(t *testing.T) {
			t.Parallel()

			assertions := assert.New(t)

			ctx, cancel := utils.NewContextWithTimeout(time.Hour)
			defer cancel()

			err := w.Sync(ctx, true)
			assertions.Nil(err, "failed to sync")

			sweepDstAddr, err := w.NewAddress(ctx, wallets.NewAddressRequest{Label: "sweep_destination"})
			assertions.Nil(err, "failed to create sweep destination address")

			_, err = w.SweepAll(ctx, wallets.SweepRequest{
				SourceIndex: ^uint64(0),
				Destination: sweepDstAddr.Address,
				Priority:    wallets.PriorityHigh,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "sweep from non-existent address should fail")
			t.Logf("Attempted sweep from non-existent address, got expected error: %v", err)
		})
	})
}
