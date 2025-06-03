package monero

import (
	"context"
	"errors"
	"fmt"
	"log"

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

	var createAddress = rpc.CreateAddressRequest{
		AccountIndex: w.accountAddressIndex,
		Label:        req.Label,
	}
	addr, err := w.client.CreateAddress(ctx, &createAddress)
	if err != nil {
		return account, fmt.Errorf("failed to create address: %w", err)
	}

	err = w.client.Store(ctx)
	if err != nil {
		return account, fmt.Errorf("failed to save changes: %w", err)
	}

	account = blockchains.Account{
		Address:         addr.Address,
		Index:           addr.AddressIndex,
		Balance:         0,
		UnlockedBalance: 0,
	}
	return
}

func convertPriority(p blockchains.Priority) (priority rpc.Priority, err error) {
	switch p {
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

	priority, err := convertPriority(req.Priority)
	if err != nil {
		return sweep, fmt.Errorf("failed to convert priority: %w", err)
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

	err = w.client.Store(ctx)
	if err != nil {
		return sweep, fmt.Errorf("failed to save changes: %w", err)
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
		return transfer, fmt.Errorf("failed to validate source address: %w: %s", err, req.Source)
	}

	err = w.validateAddress(ctx, req.Destination)
	if err != nil {
		return transfer, fmt.Errorf("failed to validate destination address: %w: %s", err, req.Destination)
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

	priority, err := convertPriority(req.Priority)
	if err != nil {
		return transfer, fmt.Errorf("failed to convert priority: %w", err)
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

	err = w.client.Store(ctx)
	if err != nil {
		return transfer, fmt.Errorf("failed to save changes: %w", err)
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

	err = w.validateAddress(ctx, req.Address)
	if err != nil {
		return balance, fmt.Errorf("failed to validate address: %w", err)
	}

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

	for index, account := range accounts.SubaddressAccounts {
		log.Println("-", account.Label, "==", config.Account, "=", config.Account == account.Label)
		if account.Label != config.Account {
			continue
		}

		var getAddress = rpc.GetAddressRequest{AccountIndex: uint64(index)}
		addr, err := w.client.GetAddress(ctx, &getAddress)
		if err != nil {
			return w, fmt.Errorf("failed to get address: %w", err)
		}

		log.Println(account)
		w.accountAddress = addr.Address
		w.accountAddressIndex = uint64(index)
		break
	}

	if w.accountAddress == "" {
		return w, fmt.Errorf("%w: %s", ErrNoAccountFound, config.Account)
	}

	return
}
