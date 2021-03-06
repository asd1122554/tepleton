package simplestake

import (
	"fmt"

	"testing"

	"github.com/stretchr/testify/assert"

	wrsp "github.com/tepleton/wrsp/types"
	crypto "github.com/tepleton/go-crypto"
	dbm "github.com/tepleton/tmlibs/db"

	"github.com/tepleton/tepleton-sdk/store"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/bank"
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	authKey := sdk.NewKVStoreKey("authkey")
	capKey := sdk.NewKVStoreKey("capkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, authKey, capKey
}

func TestKeeperGetSet(t *testing.T) {
	ms, _, capKey := setupMultiStore()

	ctx := sdk.NewContext(ms, wrsp.Header{}, false, nil)
	stakeKeeper := NewKeeper(capKey, bank.NewCoinKeeper(nil))
	addr := sdk.Address([]byte("some-address"))

	bi := stakeKeeper.getBondInfo(ctx, addr)
	assert.Equal(t, bi, bondInfo{})

	privKey := crypto.GenPrivKeyEd25519()

	bi = bondInfo{
		PubKey: privKey.PubKey(),
		Power:  int64(10),
	}
	fmt.Printf("Pubkey: %v\n", privKey.PubKey())
	stakeKeeper.setBondInfo(ctx, addr, bi)

	savedBi := stakeKeeper.getBondInfo(ctx, addr)
	assert.NotNil(t, savedBi)
	fmt.Printf("Bond Info: %v\n", savedBi)
	assert.Equal(t, int64(10), savedBi.Power)
}

func TestBonding(t *testing.T) {
	ms, authKey, capKey := setupMultiStore()
	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)

	ctx := sdk.NewContext(ms, wrsp.Header{}, false, nil)

	accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})
	coinKeeper := bank.NewCoinKeeper(accountMapper)
	stakeKeeper := NewKeeper(capKey, coinKeeper)
	addr := sdk.Address([]byte("some-address"))
	privKey := crypto.GenPrivKeyEd25519()
	pubKey := privKey.PubKey()

	_, _, err := stakeKeeper.unbondWithoutCoins(ctx, addr)
	assert.Equal(t, err, ErrInvalidUnbond())

	_, err = stakeKeeper.bondWithoutCoins(ctx, addr, pubKey, sdk.Coin{"steak", 10})
	assert.Nil(t, err)

	power, err := stakeKeeper.bondWithoutCoins(ctx, addr, pubKey, sdk.Coin{"steak", 10})
	assert.Equal(t, int64(20), power)

	pk, _, err := stakeKeeper.unbondWithoutCoins(ctx, addr)
	assert.Nil(t, err)
	assert.Equal(t, pubKey, pk)

	_, _, err = stakeKeeper.unbondWithoutCoins(ctx, addr)
	assert.Equal(t, err, ErrInvalidUnbond())
}
