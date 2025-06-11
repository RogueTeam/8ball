package monero

import (
	"context"
	"errors"
	"fmt"

	"anarchy.ttfm.onion/gateway/blockchains"
	"anarchy.ttfm.onion/gateway/internal/walletrpc/rpc"
	"anarchy.ttfm.onion/gateway/utils"
)

type Config struct {
	Client *rpc.Client
}

type Wallet struct {
	client *rpc.Client
}

var (
	ErrNoAccountFound   = errors.New("no account found with that name")
	ErrInvalidAddrIndex = errors.New("invalid address index")
	ErrInvalidAddress   = errors.New("invalid address")
)

var _ blockchains.Wallet = (*Wallet)(nil)

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

func (w *Wallet) NewAccount(req blockchains.NewAccountRequest) (account blockchains.Account, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	var createAccount = rpc.CreateAccountRequest{
		Label: req.Label,
	}
	a, err := w.client.CreateAccount(ctx, &createAccount)
	if err != nil {
		return account, fmt.Errorf("failed to create account: %w", err)
	}

	err = w.client.Store(ctx)
	if err != nil {
		return account, fmt.Errorf("failed to save changes: %w", err)
	}

	account = blockchains.Account{
		Address:         a.Address,
		Index:           a.AccountIndex,
		Balance:         0,
		UnlockedBalance: 0,
	}
	return
}

func convertPriority(p blockchains.Priority) (priority rpc.Priority, err error) {
	switch p {
	case "":
		return rpc.PriorityDefault, nil
	case blockchains.PriorityLow:
		return rpc.PriorityUnimportant, nil
	case blockchains.PriorityMedium:
		return rpc.PriorityNormal, nil
	case blockchains.PriorityHigh:
		return rpc.PriorityElevated, nil
	default:
		return priority, blockchains.ErrInvalidPriority
	}
}

func (w *Wallet) SweepAll(req blockchains.SweepRequest) (sweep blockchains.Sweep, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	err = w.validateAddress(ctx, req.Destination)
	if err != nil {
		return sweep, fmt.Errorf("failed to validate destination address: %w", err)
	}

	priority, err := convertPriority(req.Priority)
	if err != nil {
		return sweep, fmt.Errorf("failed to convert priority: %w", err)
	}

	var trans = rpc.SweepAllRequest{
		Address:      req.Destination,
		AccountIndex: req.SourceIndex,
		// SubaddrIndices: []uint64{},
		Priority:      priority,
		RingSize:      16, // Fixed by the network. May require update in the future
		UnlockTime:    req.UnlockTime,
		GetTxKeys:     true,
		GetTxHex:      true,
		GetTxMetadata: true,
	}

	res, err := w.client.SweepAll(ctx, &trans)
	if err != nil {
		return sweep, fmt.Errorf("failed to transfer monero: %w", err)
	}

	err = w.client.Store(ctx)
	if err != nil {
		return sweep, fmt.Errorf("failed to save changes: %w", err)
	}

	sweep = blockchains.Sweep{
		Address:     res.TxHashList,
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      utils.MapInt[int, uint64](res.AmountList),
		Fee:         utils.MapInt[int, uint64](res.FeeList),
	}

	return
}

func (w *Wallet) Transfer(req blockchains.TransferRequest) (transfer blockchains.Transfer, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	err = w.validateAddress(ctx, req.Destination)
	if err != nil {
		return transfer, fmt.Errorf("failed to validate destination address: %w: %s", err, req.Destination)
	}

	priority, err := convertPriority(req.Priority)
	if err != nil {
		return transfer, fmt.Errorf("failed to convert priority: %w", err)
	}

	var trans = rpc.TransferRequest{
		Destinations: []rpc.Destination{
			{Amount: req.Amount, Address: req.Destination},
		},
		AccountIndex: req.SourceIndex,
		// SubaddrIndices: []uint64{},
		Priority:      priority,
		RingSize:      16, // Fixed by the network. May require update in the future
		UnlockTime:    req.UnlockTime,
		GetTxKey:      true,
		GetTxHex:      true,
		GetTxMetadata: true,
	}

	res, err := w.client.Transfer(ctx, &trans)
	if err != nil {
		return transfer, fmt.Errorf("failed to transfer monero: %w", err)
	}

	err = w.client.Store(ctx)
	if err != nil {
		return transfer, fmt.Errorf("failed to save changes: %w", err)
	}

	transfer = blockchains.Transfer{
		Address:     res.TxHash,
		SourceIndex: req.SourceIndex,
		Destination: req.Destination,
		Amount:      res.Amount,
		Fee:         res.Fee,
	}

	return
}

func (w *Wallet) Account(req blockchains.AccountRequest) (account blockchains.Account, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	addr, err := w.client.GetAddress(ctx, &rpc.GetAddressRequest{AccountIndex: req.Index})
	if err != nil {
		return account, fmt.Errorf("failed to get account address: %w", err)
	}

	accountBalance, err := w.client.GetBalance(ctx, &rpc.GetBalanceRequest{
		AccountIndex: req.Index,
	})
	if err != nil {
		return account, fmt.Errorf("failed to get account balance: %w", err)
	}

	account = blockchains.Account{
		Address:         addr.Address,
		Index:           req.Index,
		Balance:         accountBalance.Balance,
		UnlockedBalance: accountBalance.UnlockedBalance,
	}
	return
}

func (w *Wallet) ValidateAddress(req blockchains.ValidateAddressRequest) (valid blockchains.ValidateAddress, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	err = w.validateAddress(ctx, req.Address)
	if err != nil {
		return valid, fmt.Errorf("failed to validate address: %w", err)
	}

	valid = blockchains.ValidateAddress{
		Valid: true,
	}
	return
}

func New(config Config) (w Wallet) {
	w.client = config.Client
	return w
}
