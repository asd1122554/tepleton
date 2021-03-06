package main

import (
	"fmt"
	"os"

	"github.com/tepleton/tepleton-sdk/app"
	"github.com/tepleton/tepleton-sdk/store"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/bank"
	"github.com/tepleton/wrsp/server"
	"github.com/tepleton/go-wire"
	cmn "github.com/tepleton/tmlibs/common"
	dbm "github.com/tepleton/tmlibs/db"

	bcm "github.com/tepleton/tepleton-sdk/examples/basecoin/types"
)

func main() {

	// Create the underlying leveldb datastore which will
	// persist the Merkle tree inner & leaf nodes.
	db, err := dbm.NewGoLevelDB("basecoin", "basecoin-data")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create CommitStoreLoader.
	cacheSize := 10000
	numHistory := int64(100)
	mainLoader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)
	ibcLoader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)

	// The key to access the main KVStore.
	var mainStoreKey = sdk.NewKVStoreKey("main")
	var ibcStoreKey = sdk.NewKVStoreKey("ibc")

	// Create MultiStore
	multiStore := store.NewCommitMultiStore(db)
	multiStore.SetSubstoreLoader(mainStoreKey, mainLoader)
	multiStore.SetSubstoreLoader(ibcStoreKey, ibcLoader)

	// Create the Application.
	app := app.NewApp("basecoin", multiStore)

	// Set Tx decoder
	app.SetTxDecoder(decodeTx)

	var accStore = auth.NewAccountStore(mainStoreKey, bcm.AppAccountCodec{})
	var authAnteHandler = auth.NewAnteHandler(accStore)

	// Handle charging fees and checking signatures.
	app.SetDefaultAnteHandler(authAnteHandler)

	// Add routes to App.
	app.Router().AddRoute("bank", bank.NewHandler(accStore))

	// TODO: load genesis
	// TODO: InitChain with validators
	// accounts := auth.NewAccountStore(multiStore.GetKVStore("main"))
	// TODO: set the genesis accounts

	// Load the stores.
	if err := app.LoadLatestVersion(mainStoreKey); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start the WRSP server
	srv, err := server.NewServer("0.0.0.0:46658", "socket", app)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	srv.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
	return
}

//----------------------------------------
// Misc.

func registerMsgs() {
	wire.RegisterInterface((*sdk.Msg)(nil), nil)
	wire.RegisterConcrete(&bank.SendMsg{}, "com.tepleton.basecoin.send_msg", nil)
}

func decodeTx(txBytes []byte) (sdk.Tx, error) {
	var tx = sdk.StdTx{}
	err := wire.UnmarshalBinary(txBytes, &tx)
	return tx, err
}
