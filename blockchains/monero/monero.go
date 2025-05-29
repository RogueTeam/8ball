package monero

import (
	"context"
	"errors"
	"fmt"

	"anarchy.ttfm.onion/gateway/blockchains"
	"anarchy.ttfm.onion/gateway/utils"
	"github.com/dev-warrior777/go-monero/rpc"
)

type Config struct {
	Client *rpc.Client

	Account  string
	Filename string
	Password string
}

type Wallet struct {
	client *rpc.Client

	accountAddress      string
	accountAddressIndex uint64
}

var (
	ErrNoAccountFound   = errors.New("no account found with that name")
	ErrInvalidAddrIndex = errors.New("invalid address index")
	ErrInvalidAddress   = errors.New("invalid address")
)

var _ blockchains.Wallet = (*Wallet)(nil)

func (w *Wallet) validateAddress(ctx context.Context, address string) (err error) {
	var validate = rpc.ValidateAddressRequest{
		Address:        address,
		AllowOpenalias: true,
	}
	valid, err := w.client.ValidateAddress(ctx, &validate)
	if err != nil {
		return fmt.Errorf("failed to validate address: %w", err)
	}

	if !valid.Valid {
		return ErrInvalidAddress
	}
	return nil
}

func (w *Wallet) NewAddress(req blockchains.NewAddressRequest) (address blockchains.Address, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	var createAddress = rpc.CreateAddressRequest{
		AccountIndex: w.accountAddressIndex,
		Label:        req.Label,
	}
	addr, err := w.client.CreateAddress(ctx, &createAddress)
	if err != nil {
		return address, fmt.Errorf("failed to create address: %w", err)
	}

	address = blockchains.Address{
		AccountAddress: w.accountAddress,
		AccountIndex:   w.accountAddressIndex,
		Address:        addr.Address,
		Index:          addr.AddressIndex,
	}
	return
}

func (w *Wallet) SweepAll(req blockchains.SweepRequest) (sweep blockchains.Sweep, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	err = w.validateAddress(ctx, req.Source)
	if err != nil {
		return sweep, fmt.Errorf("failed to validate source address: %w", err)
	}

	err = w.validateAddress(ctx, req.Destination)
	if err != nil {
		return sweep, fmt.Errorf("failed to validate destination address: %w", err)
	}

	var getSrcIndex = rpc.GetAddressIndexRequest{
		Address: req.Source,
	}
	addrIndex, err := w.client.GetAddressIndex(ctx, &getSrcIndex)
	if err != nil {
		return sweep, fmt.Errorf("failed to get source address index: %w", err)
	}

	if addrIndex.Index.Major != w.accountAddressIndex {
		return sweep, fmt.Errorf("source address index doesn't match account: %w", ErrInvalidAddrIndex)
	}

	var priority rpc.Priority
	switch req.Priority {
	case blockchains.PriorityLow:
		priority = rpc.PriorityUnimportant
	case blockchains.PriorityMedium:
		priority = rpc.PriorityNormal
	case blockchains.PriorityHigh:
		priority = rpc.PriorityElevated
	default:
		return sweep, blockchains.ErrInvalidPriority
	}

	var trans = rpc.SweepAllRequest{
		Address:      req.Destination,
		AccountIndex: w.accountAddressIndex,
		SubaddrIndices: []uint64{
			addrIndex.Index.Minor,
		},
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

	sweep = blockchains.Sweep{
		Address:     res.TxHashList,
		Source:      req.Source,
		Destination: req.Destination,
		Amount:      utils.MapInt[int, uint64](res.AmountList),
		Fee:         utils.MapInt[int, uint64](res.FeeList),
	}

	return
}

func (w *Wallet) Transfer(req blockchains.TransferRequest) (transfer blockchains.Transfer, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	err = w.validateAddress(ctx, req.Source)
	if err != nil {
		return transfer, fmt.Errorf("failed to validate source address: %w", err)
	}

	err = w.validateAddress(ctx, req.Destination)
	if err != nil {
		return transfer, fmt.Errorf("failed to validate destination address: %w", err)
	}

	var getSrcIndex = rpc.GetAddressIndexRequest{
		Address: req.Source,
	}
	addrIndex, err := w.client.GetAddressIndex(ctx, &getSrcIndex)
	if err != nil {
		return transfer, fmt.Errorf("failed to get source address index: %w", err)
	}

	if addrIndex.Index.Major != w.accountAddressIndex {
		return transfer, fmt.Errorf("source address index doesn't match account: %w", ErrInvalidAddrIndex)
	}

	var priority rpc.Priority
	switch req.Priority {
	case blockchains.PriorityLow:
		priority = rpc.PriorityUnimportant
	case blockchains.PriorityMedium:
		priority = rpc.PriorityNormal
	case blockchains.PriorityHigh:
		priority = rpc.PriorityElevated
	default:
		return transfer, blockchains.ErrInvalidPriority
	}

	var trans = rpc.TransferRequest{
		Destinations: []rpc.Destination{
			{Amount: req.Amount, Address: req.Destination},
		},
		AccountIndex: w.accountAddressIndex,
		SubaddrIndices: []uint64{
			addrIndex.Index.Minor,
		},
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

	transfer = blockchains.Transfer{
		Address:     res.TxHash,
		Source:      req.Source,
		Destination: req.Destination,
		Amount:      res.Amount,
		Fee:         res.Fee,
	}

	return
}

func (w *Wallet) Balance() (balance blockchains.Balance, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	accountBalance, err := w.client.GetBalance(ctx, &rpc.GetBalanceRequest{
		AccountIndex: w.accountAddressIndex,
	})
	if err != nil {
		return balance, fmt.Errorf("failed to get account balance: %w", err)
	}

	balance = blockchains.Balance{
		Address:  w.accountAddress,
		Amount:   accountBalance.Balance,
		Unlocked: accountBalance.UnlockedBalance,
	}
	return
}

func (w *Wallet) AddressBalance(req blockchains.AddressBalanceRequest) (balance blockchains.Balance, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	var getSrcIndex = rpc.GetAddressIndexRequest{
		Address: req.Address,
	}
	addrIndex, err := w.client.GetAddressIndex(ctx, &getSrcIndex)
	if err != nil {
		return balance, fmt.Errorf("failed to get source address index: %w", err)
	}

	if addrIndex.Index.Major != w.accountAddressIndex {
		return balance, fmt.Errorf("source address index doesn't match account: %w", ErrInvalidAddrIndex)
	}

	accountBalance, err := w.client.GetBalance(ctx, &rpc.GetBalanceRequest{
		AccountIndex:   w.accountAddressIndex,
		AddressIndices: []uint64{addrIndex.Index.Minor},
	})
	if err != nil {
		return balance, fmt.Errorf("failed to get account balance: %w", err)
	}

	balance = blockchains.Balance{
		Address:  w.accountAddress,
		Amount:   accountBalance.Balance,
		Unlocked: accountBalance.UnlockedBalance,
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

func New(config Config) (w Wallet, err error) {
	ctx, cancel := utils.NewContext()
	defer cancel()

	w.client = config.Client

	// Open Monero wallet
	var openWallet = rpc.OpenWalletRequest{
		Filename: config.Filename,
		Password: config.Password,
	}
	err = w.client.OpenWallet(ctx, &openWallet)
	if err != nil {
		return w, fmt.Errorf("failed to open wallet: %w", err)
	}

	// Get account index
	var getAccounts rpc.GetAccountsRequest
	accounts, err := w.client.GetAccounts(ctx, &getAccounts)
	if err != nil {
		return w, fmt.Errorf("failed to list wallet accounts: %w", err)
	}

	for _, account := range accounts.SubaddressAccounts {
		if account.Label != config.Account {
			continue
		}

		w.accountAddress = account.Address
		w.accountAddressIndex = account.AddressIndex
		break
	}

	if w.accountAddress == "" {
		return w, ErrNoAccountFound
	}

	return
}
