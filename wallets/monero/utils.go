package monero

import (
	"context"
	"fmt"

	"github.com/RogueTeam/8ball/internal/walletrpc/rpc"
	wallets "github.com/RogueTeam/8ball/wallets"
)

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
