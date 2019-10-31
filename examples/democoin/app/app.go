package app

import (
	"encoding/json"

	wrsp "github.com/tepleton/wrsp/types"
	cmn "github.com/tepleton/tmlibs/common"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	bam "github.com/tepleton/tepleton-sdk/baseapp"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/bank"
	"github.com/tepleton/tepleton-sdk/x/ibc"
	"github.com/tepleton/tepleton-sdk/x/simplestake"

	"github.com/tepleton/tepleton-sdk/examples/democoin/types"
	"github.com/tepleton/tepleton-sdk/examples/democoin/x/cool"
	"github.com/tepleton/tepleton-sdk/examples/democoin/x/pow"
	"github.com/tepleton/tepleton-sdk/examples/democoin/x/sketchy"
)

const (
	appName = "DemocoinApp"
)

// Extended WRSP application
type DemocoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore    *sdk.KVStoreKey
	capKeyAccountStore *sdk.KVStoreKey
	capKeyPowStore     *sdk.KVStoreKey
	capKeyIBCStore     *sdk.KVStoreKey
	capKeyStakingStore *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper

	// Handle fees
	feeHandler sdk.FeeHandler
}

func NewDemocoinApp(logger log.Logger, db dbm.DB) *DemocoinApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &DemocoinApp{
		BaseApp:            bam.NewBaseApp(appName, logger, db),
		cdc:                cdc,
		capKeyMainStore:    sdk.NewKVStoreKey("main"),
		capKeyAccountStore: sdk.NewKVStoreKey("acc"),
		capKeyPowStore:     sdk.NewKVStoreKey("pow"),
		capKeyIBCStore:     sdk.NewKVStoreKey("ibc"),
		capKeyStakingStore: sdk.NewKVStoreKey("stake"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		cdc,
		app.capKeyMainStore, // target store
		&types.AppAccount{}, // prototype
	).Seal()

	// Add handlers.
	coinKeeper := bank.NewCoinKeeper(app.accountMapper)
	coolKeeper := cool.NewKeeper(app.capKeyMainStore, coinKeeper, app.RegisterCodespace(cool.DefaultCodespace))
	powKeeper := pow.NewKeeper(app.capKeyPowStore, pow.NewPowConfig("pow", int64(1)), coinKeeper, app.RegisterCodespace(pow.DefaultCodespace))
	ibcMapper := ibc.NewIBCMapper(app.cdc, app.capKeyIBCStore, app.RegisterCodespace(ibc.DefaultCodespace))
	stakeKeeper := simplestake.NewKeeper(app.capKeyStakingStore, coinKeeper, app.RegisterCodespace(simplestake.DefaultCodespace))
	app.Router().
		AddRoute("bank", bank.NewHandler(coinKeeper)).
		AddRoute("cool", cool.NewHandler(coolKeeper)).
		AddRoute("pow", powKeeper.Handler).
		AddRoute("sketchy", sketchy.NewHandler()).
		AddRoute("ibc", ibc.NewHandler(ibcMapper, coinKeeper)).
		AddRoute("simplestake", simplestake.NewHandler(stakeKeeper))

	// Define the feeHandler.
	app.feeHandler = auth.BurnFeeHandler

	// Initialize BaseApp.
	app.SetTxDecoder(app.txDecoder)
	app.SetInitChainer(app.initChainerFn(coolKeeper, powKeeper))
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore, app.capKeyPowStore, app.capKeyIBCStore, app.capKeyStakingStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeHandler))
	err := app.LoadLatestVersion(app.capKeyMainStore)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.SendMsg{}, "democoin/Send", nil)
	cdc.RegisterConcrete(bank.IssueMsg{}, "democoin/Issue", nil)
	cdc.RegisterConcrete(cool.QuizMsg{}, "democoin/Quiz", nil)
	cdc.RegisterConcrete(cool.SetTrendMsg{}, "democoin/SetTrend", nil)
	cdc.RegisterConcrete(pow.MineMsg{}, "democoin/Mine", nil)
	cdc.RegisterConcrete(ibc.IBCTransferMsg{}, "democoin/IBCTransferMsg", nil)
	cdc.RegisterConcrete(ibc.IBCReceiveMsg{}, "democoin/IBCReceiveMsg", nil)
	cdc.RegisterConcrete(simplestake.BondMsg{}, "democoin/BondMsg", nil)
	cdc.RegisterConcrete(simplestake.UnbondMsg{}, "democoin/UnbondMsg", nil)

	// Register AppAccount
	cdc.RegisterInterface((*sdk.Account)(nil), nil)
	cdc.RegisterConcrete(&types.AppAccount{}, "democoin/Account", nil)

	// Register crypto.
	wire.RegisterCrypto(cdc)

	return cdc
}

// custom logic for transaction decoding
func (app *DemocoinApp) txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx = sdk.StdTx{}

	if len(txBytes) == 0 {
		return nil, sdk.ErrTxDecode("txBytes are empty")
	}

	// StdTx.Msg is an interface. The concrete types
	// are registered by MakeTxCodec in bank.RegisterWire.
	err := app.cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxDecode("").Trace(err.Error())
	}
	return tx, nil
}

// custom logic for democoin initialization
func (app *DemocoinApp) initChainerFn(coolKeeper cool.Keeper, powKeeper pow.Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req wrsp.RequestInitChain) wrsp.ResponseInitChain {
		stateJSON := req.AppStateBytes

		genesisState := new(types.GenesisState)
		err := json.Unmarshal(stateJSON, genesisState)
		if err != nil {
			panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
			// return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, gacc := range genesisState.Accounts {
			acc, err := gacc.ToAppAccount()
			if err != nil {
				panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
				//	return sdk.ErrGenesisParse("").TraceCause(err, "")
			}
			app.accountMapper.SetAccount(ctx, acc)
		}

		// Application specific genesis handling
		err = coolKeeper.InitGenesis(ctx, genesisState.CoolGenesis)
		if err != nil {
			panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		err = powKeeper.InitGenesis(ctx, genesisState.PowGenesis)
		if err != nil {
			panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		return wrsp.ResponseInitChain{}
	}
}
