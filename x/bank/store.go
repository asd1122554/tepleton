package bank

import (
	"fmt"

	sdk "github.com/tepleton/tepleton-sdk/types"
	crypto "github.com/tepleton/go-crypto"
)

// CoinStore manages transfers between accounts
type CoinStore struct {
	store sdk.AccountStore
}

// SubtractCoins subtracts amt from the coins at the addr.
func (cs CoinStore) SubtractCoins(ctx sdk.Context, addr crypto.Address, amt sdk.Coins) (sdk.Coins, error) {
	acc := cs.store.GetAccount(ctx, addr)
	if acc == nil {
		return amt, fmt.Errorf("Sending account (%s) does not exist", addr)
	}

	coins := acc.GetCoins()
	newCoins := coins.Minus(amt)
	if !newCoins.IsNotNegative() {
		return amt, ErrInsufficientCoins(fmt.Sprintf("%s < %s", coins, amt))
	}

	acc.SetCoins(newCoins)
	cs.store.SetAccount(ctx, acc)
	return newCoins, nil
}

// AddCoins adds amt to the coins at the addr.
func (cs CoinStore) AddCoins(ctx sdk.Context, addr crypto.Address, amt sdk.Coins) (sdk.Coins, error) {
	acc := cs.store.GetAccount(ctx, addr)
	if acc == nil {
		acc = cs.store.NewAccountWithAddress(ctx, addr)
	}

	coins := acc.GetCoins()
	newCoins := coins.Plus(amt)

	acc.SetCoins(newCoins)
	cs.store.SetAccount(ctx, acc)
	return newCoins, nil
}
