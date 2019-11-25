package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tepleton/tepleton-sdk/tests/mock"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/bank"

	wrsp "github.com/tepleton/wrsp/types"
	crypto "github.com/tepleton/go-crypto"
)

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	keyIBC := sdk.NewKVStoreKey("ibc")
	ibcMapper := NewMapper(mapp.Cdc, keyIBC, mapp.RegisterCodespace(DefaultCodespace))
	coinKeeper := bank.NewKeeper(mapp.AccountMapper)
	mapp.Router().AddRoute("ibc", NewHandler(ibcMapper, coinKeeper))

	mapp.CompleteSetup(t, []*sdk.KVStoreKey{keyIBC})
	return mapp
}

func TestIBCMsgs(t *testing.T) {
	gapp := getMockApp(t)

	sourceChain := "source-chain"
	destChain := "dest-chain"

	priv1 := crypto.GenPrivKeyEd25519()
	addr1 := priv1.PubKey().Address()
	coins := sdk.Coins{{"foocoin", 10}}
	var emptyCoins sdk.Coins

	acc := &auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	accs := []auth.Account{acc}

	mock.SetGenesis(gapp, accs)

	// A checkTx context (true)
	ctxCheck := gapp.BaseApp.NewContext(true, wrsp.Header{})
	res1 := gapp.AccountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc, res1)

	packet := IBCPacket{
		SrcAddr:   addr1,
		DestAddr:  addr1,
		Coins:     coins,
		SrcChain:  sourceChain,
		DestChain: destChain,
	}

	transferMsg := IBCTransferMsg{
		IBCPacket: packet,
	}

	receiveMsg := IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   addr1,
		Sequence:  0,
	}

	mock.SignCheckDeliver(t, gapp, transferMsg, []int64{0}, true, priv1)
	mock.CheckBalance(t, gapp, addr1, emptyCoins)
	mock.SignCheckDeliver(t, gapp, transferMsg, []int64{1}, false, priv1)
	mock.SignCheckDeliver(t, gapp, receiveMsg, []int64{2}, true, priv1)
	mock.CheckBalance(t, gapp, addr1, coins)
	mock.SignCheckDeliver(t, gapp, receiveMsg, []int64{3}, false, priv1)
}
