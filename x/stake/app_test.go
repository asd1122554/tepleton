package stake

import (
	"testing"

	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/auth/mock"
	"github.com/tepleton/tepleton-sdk/x/bank"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wrsp "github.com/tepleton/wrsp/types"
	crypto "github.com/tepleton/go-crypto"
)

var (
	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = priv1.PubKey().Address()
	priv2 = crypto.GenPrivKeyEd25519()
	addr2 = priv2.PubKey().Address()
	addr3 = crypto.GenPrivKeyEd25519().PubKey().Address()
	priv4 = crypto.GenPrivKeyEd25519()
	addr4 = priv4.PubKey().Address()
	coins = sdk.Coins{{"foocoin", 10}}
	fee   = auth.StdFee{
		sdk.Coins{{"foocoin", 0}},
		100000,
	}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) (*mock.App, Keeper) {
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	keyStake := sdk.NewKVStoreKey("stake")
	coinKeeper := bank.NewKeeper(mapp.AccountMapper)
	keeper := NewKeeper(mapp.Cdc, keyStake, coinKeeper, mapp.RegisterCodespace(DefaultCodespace))
	mapp.Router().AddRoute("stake", NewHandler(keeper))

	mapp.SetEndBlocker(getEndBlocker(keeper))
	mapp.SetInitChainer(getInitChainer(mapp, keeper))

	mapp.CompleteSetup(t, []*sdk.KVStoreKey{keyStake})
	return mapp, keeper
}

// stake endblocker
func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req wrsp.RequestEndBlock) wrsp.ResponseEndBlock {
		validatorUpdates := EndBlocker(ctx, keeper)
		return wrsp.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
		}
	}
}

// overwrite the mock init chainer
func getInitChainer(mapp *mock.App, keeper Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req wrsp.RequestInitChain) wrsp.ResponseInitChain {
		mapp.InitChainer(ctx, req)
		InitGenesis(ctx, keeper, DefaultGenesisState())

		return wrsp.ResponseInitChain{}
	}
}

//__________________________________________________________________________________________

func checkValidator(t *testing.T, mapp *mock.App, keeper Keeper,
	addr sdk.Address, expFound bool) Validator {

	ctxCheck := mapp.BaseApp.NewContext(true, wrsp.Header{})
	validator, found := keeper.GetValidator(ctxCheck, addr1)
	assert.Equal(t, expFound, found)
	return validator
}

func checkDelegation(t *testing.T, mapp *mock.App, keeper Keeper, delegatorAddr,
	validatorAddr sdk.Address, expFound bool, expShares sdk.Rat) {

	ctxCheck := mapp.BaseApp.NewContext(true, wrsp.Header{})
	delegation, found := keeper.GetDelegation(ctxCheck, delegatorAddr, validatorAddr)
	if expFound {
		assert.True(t, found)
		assert.True(sdk.RatEq(t, expShares, delegation.Shares))
		return
	}
	assert.False(t, found)
}

func TestStakeMsgs(t *testing.T) {
	mapp, keeper := getMockApp(t)

	genCoin := sdk.Coin{"steak", 42}
	bondCoin := sdk.Coin{"steak", 10}

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{genCoin},
	}
	accs := []auth.Account{acc1, acc2}

	mock.SetGenesis(mapp, accs)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{genCoin})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{genCoin})

	////////////////////
	// Create Validator

	description := NewDescription("foo_moniker", "", "", "")
	createValidatorMsg := NewMsgCreateValidator(
		addr1, priv1.PubKey(), bondCoin, description,
	)
	mock.SignCheckDeliver(t, mapp.BaseApp, createValidatorMsg, []int64{0}, []int64{0}, true, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{genCoin.Minus(bondCoin)})
	mapp.BeginBlock(wrsp.RequestBeginBlock{})

	validator := checkValidator(t, mapp, keeper, addr1, true)
	require.Equal(t, addr1, validator.Owner)
	require.Equal(t, sdk.Bonded, validator.Status())
	require.True(sdk.RatEq(t, sdk.NewRat(10), validator.PoolShares.Bonded()))

	// check the bond that should have been created as well
	checkDelegation(t, mapp, keeper, addr1, addr1, true, sdk.NewRat(10))

	////////////////////
	// Edit Validator

	description = NewDescription("bar_moniker", "", "", "")
	editValidatorMsg := NewMsgEditValidator(addr1, description)
	mock.SignCheckDeliver(t, mapp.BaseApp, editValidatorMsg, []int64{0}, []int64{1}, true, priv1)
	validator = checkValidator(t, mapp, keeper, addr1, true)
	require.Equal(t, description, validator.Description)

	////////////////////
	// Delegate

	mock.CheckBalance(t, mapp, addr2, sdk.Coins{genCoin})
	delegateMsg := NewMsgDelegate(addr2, addr1, bondCoin)
	mock.SignCheckDeliver(t, mapp.BaseApp, delegateMsg, []int64{1}, []int64{0}, true, priv2)
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{genCoin.Minus(bondCoin)})
	checkDelegation(t, mapp, keeper, addr2, addr1, true, sdk.NewRat(10))

	////////////////////
	// Unbond

	unbondMsg := NewMsgUnbond(addr2, addr1, "MAX")
	mock.SignCheckDeliver(t, mapp.BaseApp, unbondMsg, []int64{1}, []int64{1}, true, priv2)
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{genCoin})
	checkDelegation(t, mapp, keeper, addr2, addr1, false, sdk.Rat{})
}
