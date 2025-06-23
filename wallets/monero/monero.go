package monero

import (
	"context"
	"errors"
	"fmt"

	"anarchy.ttfm/8ball/internal/walletrpc/rpc"
	"anarchy.ttfm/8ball/utils"
	wallets "anarchy.ttfm/8ball/wallets"
)

const MoneroUnit = 1_000_000_000_000

type Config struct {
	Accounts bool
	Client   *rpc.Client
}

type Wallet struct {
	accounts bool
	client   *rpc.Client
}

var (
	ErrNoAccountFound   = errors.New("no account found with that name")
	ErrInvalidAddrIndex = errors.New("invalid address index")
	ErrInvalidAddress   = errors.New("invalid address")
)

var _ wallets.Wallet = (*Wallet)(nil)

func (w *Wallet) Sync(ctx context.Context, full bool) (err error) {
	var height uint64
	if full {
		height = 1
	}

	_, err = w.client.Refresh(ctx, &rpc.RefreshRequest{StartHeight: height})
	if err != nil {
		return fmt.Errorf("failed to refresh wallet at height 0: %w", err)
	}

	err = w.client.RescanSpent(ctx)
	if err != nil {
		return fmt.Errorf("failed to rescan for spent outputs: %w", err)
	}

	return
}

func (w *Wallet) validateAddress(ctx context.Context, address string) (err error) {
	var validate = rpc.ValidateAddressRequest{
		Address: address,
		//AllowOpenalias: true,
	}
	valid, err := w.client.ValidateAddress(ctx, &validate)
	if err != nil {
		return fmt.Errorf("failed to validate address: %s: %w", address, err)
	}

	if !valid.Valid {
		return ErrInvalidAddress
	}
	return nil
}

func (w *Wallet) NewAddress(ctx context.Context, req wallets.NewAddressRequest) (address wallets.Address, err error) {
	if w.accounts {
		var createAccount = rpc.CreateAccountRequest{
			Label: req.Label,
		}
		a, err := w.client.CreateAccount(ctx, &createAccount)
		if err != nil {
			return address, fmt.Errorf("failed to create account: %w", err)
		}

		err = w.client.Store(ctx)
		if err != nil {
			return address, fmt.Errorf("failed to save changes: %w", err)
		}

		address = wallets.Address{
			Address:         a.Address,
			Index:           a.AccountIndex,
			Balance:         0,
			UnlockedBalance: 0,
		}
	} else {
		var createAddress = rpc.CreateAddressRequest{
			AccountIndex: 0,
			Label:        req.Label,
		}
		a, err := w.client.CreateAddress(ctx, &createAddress)
		if err != nil {
			return address, fmt.Errorf("failed to create address: %w", err)
		}

		err = w.client.Store(ctx)
		if err != nil {
			return address, fmt.Errorf("failed to save changes: %w", err)
		}

		address = wallets.Address{
			Address:         a.Address,
			Index:           a.AddressIndex,
			Balance:         0,
			UnlockedBalance: 0,
		}
	}
	return
}

func convertPriority(p wallets.Priority) (priority rpc.Priority, err error) {
	switch p {
	case "":
		return rpc.PriorityDefault, nil
	case wallets.PriorityLow:
		return rpc.PriorityUnimportant, nil
	case wallets.PriorityMedium:
		return rpc.PriorityNormal, nil
	case wallets.PriorityHigh:
		return rpc.PriorityElevated, nil
	default:
		return priority, wallets.ErrInvalidPriority
	}
}

func (w *Wallet) SweepAll(ctx context.Context, req wallets.SweepRequest) (sweep wallets.Sweep, err error) {
	err = w.validateAddress(ctx, req.Destination)
	if err != nil {
		return sweep, fmt.Errorf("failed to validate destination address: %w", err)
	}

	priority, err := convertPriority(req.Priority)
	if err != nil {
		return sweep, fmt.Errorf("failed to convert priority: %w", err)
	}

	var trans rpc.SweepAllRequest
	if w.accounts {
		trans = rpc.SweepAllRequest{
			Address:       req.Destination,
			AccountIndex:  req.SourceIndex,
			Priority:      priority,
			Outputs:       1,
			BelowAmount:   0xFFFFFFFFFFFFFFFF,
			RingSize:      16, // Fixed by the network. May require update in the future
			UnlockTime:    req.UnlockTime,
			GetTxKeys:     true,
			GetTxHex:      true,
			GetTxMetadata: true,
		}
	} else {
		trans = rpc.SweepAllRequest{
			Address:        req.Destination,
			AccountIndex:   0,
			SubaddrIndices: []uint64{req.SourceIndex},
			Priority:       priority,
			Outputs:        1,
			BelowAmount:    0xFFFFFFFFFFFFFFFF,
			RingSize:       16, // Fixed by the network. May require update in the future
			UnlockTime:     req.UnlockTime,
			GetTxKeys:      true,
			GetTxHex:       true,
			GetTxMetadata:  true,
		}
	}

	res, err := w.client.SweepAll(ctx, &trans)
	if err != nil {
		return sweep, fmt.Errorf("failed to transfer monero: %w", err)
	}

	err = w.client.Store(ctx)
	if err != nil {
		return sweep, fmt.Errorf("failed to save changes: %w", err)
	}

	sweep = wallets.Sweep{
		Address:     res.TxHashList[0],
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      utils.MapInt[int, uint64](res.AmountList)[0],
		Fee:         utils.MapInt[int, uint64](res.FeeList)[0],
	}

	return
}

func (w *Wallet) Transfer(ctx context.Context, req wallets.TransferRequest) (transfer wallets.Transfer, err error) {
	err = w.validateAddress(ctx, req.Destination)
	if err != nil {
		return transfer, fmt.Errorf("failed to validate destination address: %w: %s", err, req.Destination)
	}

	priority, err := convertPriority(req.Priority)
	if err != nil {
		return transfer, fmt.Errorf("failed to convert priority: %w", err)
	}
	var trans rpc.TransferRequest
	if w.accounts {
		trans = rpc.TransferRequest{
			Destinations: []rpc.Destination{
				{Amount: req.Amount, Address: req.Destination},
			},
			AccountIndex:  req.SourceIndex,
			Priority:      priority,
			RingSize:      16, // Fixed by the network. May require update in the future
			UnlockTime:    req.UnlockTime,
			GetTxKey:      true,
			GetTxHex:      true,
			GetTxMetadata: true,
		}
	} else {
		trans = rpc.TransferRequest{
			Destinations: []rpc.Destination{
				{Amount: req.Amount, Address: req.Destination},
			},
			AccountIndex:   0,
			SubaddrIndices: []uint64{req.SourceIndex},
			Priority:       priority,
			RingSize:       16, // Fixed by the network. May require update in the future
			UnlockTime:     req.UnlockTime,
			GetTxKey:       true,
			GetTxHex:       true,
			GetTxMetadata:  true,
		}
	}

	res, err := w.client.Transfer(ctx, &trans)
	if err != nil {
		return transfer, fmt.Errorf("failed to transfer monero: %w", err)
	}

	err = w.client.Store(ctx)
	if err != nil {
		return transfer, fmt.Errorf("failed to save changes: %w", err)
	}

	transfer = wallets.Transfer{
		Address:     res.TxHash,
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      res.Amount,
		Fee:         res.Fee,
	}

	err = w.client.Store(ctx)
	if err != nil {
		return transfer, fmt.Errorf("failed to save changes: %w", err)
	}

	transfer = wallets.Transfer{
		Address:     res.TxHash,
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      res.Amount,
		Fee:         res.Fee,
	}

	return
}

func (w *Wallet) Address(ctx context.Context, req wallets.AddressRequest) (address wallets.Address, err error) {
	if w.accounts {
		addr, err := w.client.GetAddress(ctx, &rpc.GetAddressRequest{
			AccountIndex: req.Index,
		})
		if err != nil {
			return address, fmt.Errorf("failed to get account balance: %w", err)
		}

		var balance uint64
		var unlocked uint64
		for _, entry := range addr.Addresses {
			balance += entry.Balance
			unlocked += entry.UnlockedBalance
		}
		address = wallets.Address{
			Address:         addr.Address,
			Index:           req.Index,
			Balance:         balance,
			UnlockedBalance: unlocked,
		}
	} else {
		addressBalance, err := w.client.GetBalance(ctx, &rpc.GetBalanceRequest{
			AccountIndex:   0,
			AddressIndices: []uint64{req.Index},
		})
		if err != nil {
			return address, fmt.Errorf("failed to get address balance: %w", err)
		}

		address = wallets.Address{
			Address:         addressBalance.PerSubaddress[0].Address,
			Index:           req.Index,
			Balance:         addressBalance.PerSubaddress[0].Balance,
			UnlockedBalance: addressBalance.PerSubaddress[0].UnlockedBalance,
		}
	}
	return
}

func (w *Wallet) ValidateAddress(ctx context.Context, req wallets.ValidateAddressRequest) (valid wallets.ValidateAddress, err error) {
	err = w.validateAddress(ctx, req.Address)
	if err != nil {
		return valid, fmt.Errorf("failed to validate address: %w", err)
	}

	valid = wallets.ValidateAddress{
		Valid: true,
	}
	return
}

func (w *Wallet) Transaction(ctx context.Context, req wallets.TransactionRequest) (tx wallets.Transaction, err error) {
	var getTransfer rpc.GetTransferByTxidRequest
	if w.accounts {
		getTransfer = rpc.GetTransferByTxidRequest{
			AccountIndex: req.SourceIndex,
			Txid:         req.TransactionId,
		}
	} else {
		getTransfer = rpc.GetTransferByTxidRequest{
			Txid: req.TransactionId,
		}
	}

	transaction, err := w.client.GetTransferByTxid(ctx, &getTransfer)
	if err != nil {
		return tx, fmt.Errorf("failed to retrieve transfer by id: %w", err)
	}

	transfer := transaction.Transfer
	tx = wallets.Transaction{
		Address: transfer.Address,
		Amount:  transfer.Amount,
	}

	switch transfer.Type {
	case "pending", "pool":
		tx.Status = wallets.TransactionStatusPending
	case "out":
		tx.Status = wallets.TransactionStatusCompleted
	case "failed":
		tx.Status = wallets.TransactionStatusFailed
	default:
		return tx, errors.New("unsupported tx type")
	}

	if len(transfer.Destinations) > 0 {
		tx.Destination = transfer.Destinations[0].Address
	}

	return tx, nil
}

func New(config Config) (w *Wallet) {
	w = &Wallet{
		accounts: config.Accounts,
		client:   config.Client,
	}
	return w
}
