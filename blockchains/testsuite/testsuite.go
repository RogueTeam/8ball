package testsuite

import (
	"testing"

	"anarchy.ttfm.onion/gateway/blockchains"
	"anarchy.ttfm.onion/gateway/blockchains/mock" // Import the mock wallet
	"anarchy.ttfm.onion/gateway/random"
	"github.com/stretchr/testify/assert"
)

// DataGenerator defines an interface for test data generation.
type DataGenerator interface {
	// TransferAmount returns the amount to send for a transfer.
	TransferAmount() (funds uint64)
}

// Test runs a comprehensive suite of tests for any Wallet implementation.
func Test(t *testing.T, w blockchains.Wallet, gen DataGenerator) {
	t.Run("Initial Account 0 State", func(t *testing.T) {
		assertions := assert.New(t)

		// Verify Account 0 exists and has zero balance initially
		account0, err := w.Account(blockchains.AccountRequest{Index: 0})
		assertions.Nil(err, "failed to retrieve initial account 0 balance")
		assertions.Equal(uint64(0), account0.Index, "Account 0 should have index 0")

		t.Logf("Initial Account 0: %+v", account0)
	})

	t.Run("NewAccount", func(t *testing.T) {
		assertions := assert.New(t)

		label := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		account, err := w.NewAccount(blockchains.NewAccountRequest{Label: label})
		assertions.Nil(err, "failed to create new account")
		assertions.NotNil(account.Address, "new account should have an address")
		assertions.Greater(account.Index, uint64(0), "new account index should be greater than 0")
		assertions.Equal(uint64(0), account.Balance, "new account should have zero balance initially")
		assertions.Equal(uint64(0), account.UnlockedBalance, "new account should have zero unlocked balance initially")

		t.Logf("Created new account: %+v", account)

		// Verify the new account can be retrieved
		retrievedAccount, err := w.Account(blockchains.AccountRequest{Index: account.Index})
		assertions.Nil(err, "failed to get newly created account")
		assertions.Equal(account, retrievedAccount, "retrieved account should match created account")
		t.Logf("Retrieved new account: %+v", retrievedAccount)

		// Test creating another account to check index increment
		label2 := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		account2, err := w.NewAccount(blockchains.NewAccountRequest{Label: label2})
		assertions.Nil(err, "failed to create second new account")
		assertions.Equal(account.Index+1, account2.Index, "second account index should be incremented")
	})

	t.Run("ValidateAddress", func(t *testing.T) {
		assertions := assert.New(t)

		label := random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)
		account, err := w.NewAccount(blockchains.NewAccountRequest{Label: label})
		assertions.Nil(err, "failed to create new account")

		// Test with a valid mock address (should always return true for the mock)
		validRes, err := w.ValidateAddress(blockchains.ValidateAddressRequest{Address: account.Address})
		assertions.Nil(err, "failed to validate address")
		assertions.True(validRes.Valid, "mock wallet should validate any address as true")
	})

	t.Run("Transfer", func(t *testing.T) {
		t.Run("To Internal Account", func(t *testing.T) {
			assertions := assert.New(t)

			currentAccount0, err := w.Account(blockchains.AccountRequest{Index: 0})
			assertions.Nil(err, "failed to get current account 0 balance")
			t.Logf("Account 0 balance before transfers: %d", currentAccount0.Balance)

			dst, err := w.NewAccount(blockchains.NewAccountRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new account for internal transfer")

			transfer, err := w.Transfer(blockchains.TransferRequest{
				SourceIndex: 0,
				Destination: dst.Address,
				Amount:      gen.TransferAmount(),
				Priority:    blockchains.PriorityLow,
				UnlockTime:  0,
			})
			assertions.Nil(err, "failed to transfer funds to internal account")
			assertions.NotEmpty(transfer.Address, "transfer should have a transaction address")
			assertions.Equal(uint64(0), transfer.SourceIndex, "source index should be 0")
			assertions.Equal(dst.Address, transfer.Destination, "destination address should match")
			t.Logf("Transfer to internal account: %+v", transfer)

			// Verify balances after internal transfer
			account0After, err := w.Account(blockchains.AccountRequest{Index: 0})
			assertions.Nil(err, "failed to get account 0 balance after internal transfer")
			assertions.Less(account0After.Balance, currentAccount0.Balance, "Account 0 balance should be reduced by transfer amount")

			dstAfter, err := w.Account(blockchains.AccountRequest{Index: dst.Index})
			t.Log(dstAfter)
			assertions.Nil(err, "failed to get destination account balance after internal transfer")
			assertions.Equal(transfer.Amount, dstAfter.Balance, "Destination account balance should increase by net transfer amount")
		})

		t.Run("Insufficient Funds", func(t *testing.T) {
			assertions := assert.New(t)

			currentAccount0, err := w.Account(blockchains.AccountRequest{Index: 0})
			assertions.Nil(err, "failed to get current account 0 balance")
			t.Logf("Account 0 balance before transfers: %d", currentAccount0.Balance)

			dst, err := w.NewAccount(blockchains.NewAccountRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new account for internal transfer")

			// Attempt to transfer more than available in Account 0
			insufficientAmount := currentAccount0.Balance + 1000 // More than current balance

			_, err = w.Transfer(blockchains.TransferRequest{
				SourceIndex: 0,
				Destination: dst.Address,
				Amount:      insufficientAmount,
				Priority:    blockchains.PriorityHigh,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "transfer should fail due to insufficient funds")
			t.Logf("Attempted transfer with insufficient funds, got expected error: %v", err)

			// Verify balance of Account 0 remains unchanged
			account0After, err := w.Account(blockchains.AccountRequest{Index: 0})
			assertions.Nil(err, "failed to get account 0 balance after failed transfer")
			assertions.Equal(currentAccount0.Balance, account0After.Balance, "Account 0 balance should not change after failed transfer")
		})

		t.Run("Zero Amount Transfer", func(t *testing.T) {
			assertions := assert.New(t)

			currentAccount0, err := w.Account(blockchains.AccountRequest{Index: 0})
			assertions.Nil(err, "failed to get current account 0 balance")
			t.Logf("Account 0 balance before transfers: %d", currentAccount0.Balance)

			dst, err := w.NewAccount(blockchains.NewAccountRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new account for internal transfer")

			_, err = w.Transfer(blockchains.TransferRequest{
				SourceIndex: 0,
				Destination: dst.Address,
				Amount:      0, // Zero amount
				Priority:    blockchains.PriorityLow,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "transfer should fail for zero amount")
			t.Logf("Attempted transfer with zero amount, got expected error: %v", err)
		})

		t.Run("Transfer from Non-Existent Account", func(t *testing.T) {
			assertions := assert.New(t)

			dst, err := w.NewAccount(blockchains.NewAccountRequest{Label: random.String(random.PseudoRand, random.CharsetAlphaNumeric, 10)})
			assertions.Nil(err, "failed to create new account for internal transfer")

			_, err = w.Transfer(blockchains.TransferRequest{
				SourceIndex: ^uint64(0),
				Destination: dst.Address,
				Amount:      gen.TransferAmount(),
				Priority:    blockchains.PriorityLow,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "transfer from non-existent account should fail")
			t.Logf("Attempted transfer from non-existent account, got expected error: %v", err)
		})
	})

	t.Run("SweepAll", func(t *testing.T) {
		t.Run("Successful Sweep", func(t *testing.T) {
			assertions := assert.New(t)

			// Create a new account and fund it specifically for this sweep test
			sweepSourceAcc, err := w.NewAccount(blockchains.NewAccountRequest{Label: "sweep_source"})
			assertions.Nil(err, "failed to create sweep source account")

			// Transfer to Source account
			_, err = w.Transfer(blockchains.TransferRequest{
				SourceIndex: 0,
				Destination: sweepSourceAcc.Address,
				Amount:      gen.TransferAmount(),
				Priority:    blockchains.PriorityLow,
				UnlockTime:  0,
			})
			assertions.Nil(err, "failed to transfer testing amount")

			sweepDstAcc, err := w.NewAccount(blockchains.NewAccountRequest{Label: "sweep_destination"})
			assertions.Nil(err, "failed to create sweep destination account")

			// Sweep the entire source to dst
			sweep, err := w.SweepAll(blockchains.SweepRequest{
				SourceIndex: sweepSourceAcc.Index,
				Destination: sweepDstAcc.Address,
				Priority:    blockchains.PriorityHigh,
				UnlockTime:  0,
			})
			assertions.Nil(err, "failed to sweep all funds")
			assertions.NotEmpty(sweep.Address, "sweep should return one transaction hash")
			assertions.NotEmpty(sweep.Amount, "sweep should return one amount")
			assertions.NotEmpty(sweep.Fee, "sweep should return one fee")
			assertions.Equal(sweepSourceAcc.Index, sweep.SourceIndex, "sweep source index should match")
			assertions.Equal(sweepDstAcc.Address, sweep.Destination, "sweep destination address should match")
			t.Logf("Successful sweep: %+v", sweep)

			// Verify balances after sweep
			sourceAccAfter, err := w.Account(blockchains.AccountRequest{Index: sweepSourceAcc.Index})
			assertions.Nil(err, "failed to get source account balance after sweep")
			assertions.Equal(uint64(0), sourceAccAfter.Balance, "source account balance should be zero after sweep")
			assertions.Equal(uint64(0), sourceAccAfter.UnlockedBalance, "source account unlocked balance should be zero after sweep")

			dstAccAfter, err := w.Account(blockchains.AccountRequest{Index: sweepDstAcc.Index})
			assertions.Nil(err, "failed to get destination account balance after sweep")
			assertions.NotZero(dstAccAfter.Balance, "destination account balance should increase by net swept amount")
		})

		t.Run("Sweep Empty Account", func(t *testing.T) {
			assertions := assert.New(t)

			// Create a new account with zero balance
			emptyAcc, err := w.NewAccount(blockchains.NewAccountRequest{Label: "empty_account"})
			assertions.Nil(err, "failed to create empty account")

			sweepDstAcc, err := w.NewAccount(blockchains.NewAccountRequest{Label: "sweep_destination"})
			assertions.Nil(err, "failed to create sweep destination account")

			_, err = w.SweepAll(blockchains.SweepRequest{
				SourceIndex: emptyAcc.Index,
				Destination: sweepDstAcc.Address,
				Priority:    blockchains.PriorityLow,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "sweeping an empty account should fail")
			assertions.Contains(err.Error(), "no balance to sweep", "error message should indicate no balance")
			t.Logf("Attempted sweep of empty account, got expected error: %v", err)
		})

		t.Run("Sweep from Non-Existent Account", func(t *testing.T) {
			assertions := assert.New(t)

			sweepDstAcc, err := w.NewAccount(blockchains.NewAccountRequest{Label: "sweep_destination"})
			assertions.Nil(err, "failed to create sweep destination account")

			_, err = w.SweepAll(blockchains.SweepRequest{
				SourceIndex: ^uint64(0),
				Destination: sweepDstAcc.Address,
				Priority:    blockchains.PriorityLow,
				UnlockTime:  0,
			})
			assertions.NotNil(err, "sweep from non-existent account should fail")
			assertions.Contains(err.Error(), mock.ErrAccountNotFound.Error(), "error message should indicate account not found")
			t.Logf("Attempted sweep from non-existent account, got expected error: %v", err)
		})
	})
}
