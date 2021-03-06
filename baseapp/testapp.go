package baseapp

import (
	wrsp "github.com/tepleton/wrsp/types"

	"github.com/tepleton/tepleton-sdk/baseapp/testtx"
	sdk "github.com/tepleton/tepleton-sdk/types"
)

// TestApp wraps BaseApp with helper methods,
// and exposes more functionality than otherwise needed.
type TestApp struct {
	*BaseApp

	// These get set as we receive them.
	*wrsp.ResponseBeginBlock
	*wrsp.ResponseEndBlock
}

func NewTestApp(bapp *BaseApp) *TestApp {
	app := &TestApp{
		BaseApp: bapp,
	}
	return app
}

// execute BaseApp BeginBlock
func (tapp *TestApp) RunBeginBlock() {
	if tapp.header != nil {
		panic("TestApp.header not nil, BeginBlock already run, or EndBlock not yet run.")
	}
	cms := tapp.CommitMultiStore()
	lastCommit := cms.LastCommitID()
	header := wrsp.Header{
		ChainID:        "chain_" + tapp.BaseApp.name,
		Height:         lastCommit.Version + 1,
		Time:           -1, // TODO
		NumTxs:         -1, // TODO
		LastCommitHash: lastCommit.Hash,
		DataHash:       nil, // TODO
		ValidatorsHash: nil, // TODO
		AppHash:        nil, // TODO
	}
	res := tapp.BeginBlock(wrsp.RequestBeginBlock{
		Hash:                nil, // TODO
		Header:              header,
		AbsentValidators:    nil, // TODO
		ByzantineValidators: nil, // TODO
	})
	tapp.ResponseBeginBlock = &res
	return
}

// kill resources used by basecapp
func (tapp *TestApp) Close() {
	tapp.db.Close()
}

func (tapp *TestApp) ensureBeginBlock() {
	if tapp.header == nil {
		panic("TestApp.header was nil, call TestApp.RunBeginBlock()")
	}
}

// run tx through CheckTx of TestApp
func (tapp *TestApp) RunCheckTx(tx sdk.Tx) sdk.Result {
	tapp.ensureBeginBlock()
	return tapp.BaseApp.runTx(true, nil, tx)
}

// run tx through DeliverTx of TestApp
func (tapp *TestApp) RunDeliverTx(tx sdk.Tx) sdk.Result {
	tapp.ensureBeginBlock()
	return tapp.BaseApp.runTx(false, nil, tx)
}

// run tx through CheckTx of TestApp
// NOTE: Skips authentication by wrapping msg in TestTx{}.
func (tapp *TestApp) RunCheckMsg(msg sdk.Msg) sdk.Result {
	var tx = testtx.TestTx{msg}
	return tapp.RunCheckTx(tx)
}

// run tx through DeliverTx of TestApp
// NOTE: Skips authentication by wrapping msg in TestTx{}.
func (tapp *TestApp) RunDeliverMsg(msg sdk.Msg) sdk.Result {
	var tx = testtx.TestTx{msg}
	return tapp.RunDeliverTx(tx)
}

// return the commited multistore
func (tapp *TestApp) CommitMultiStore() sdk.CommitMultiStore {
	return tapp.BaseApp.cms
}

// return a cache-wrap CheckTx state of multistore
func (tapp *TestApp) MultiStoreCheck() sdk.MultiStore {
	return tapp.BaseApp.msCheck
}

// return a cache-wrap DeliverTx state of multistore
func (tapp *TestApp) MultiStoreDeliver() sdk.MultiStore {
	return tapp.BaseApp.msDeliver
}
