package core

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/tepleton/tepleton-sdk/wire"
	rpcclient "github.com/tepleton/tepleton/rpc/client"
	ctypes "github.com/tepleton/tepleton/rpc/core/types"
	cmn "github.com/tepleton/tmlibs/common"

	"github.com/tepleton/tepleton-sdk/client"
	"github.com/tepleton/tepleton-sdk/client/keys"
	sdk "github.com/tepleton/tepleton-sdk/types"
)

// Broadcast the transaction bytes to Tendermint
func (ctx CoreContext) BroadcastTx(tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {

	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return res, err
	}

	if res.CheckTx.Code != uint32(0) {
		return res, errors.Errorf("CheckTx failed: (%d) %s",
			res.CheckTx.Code,
			res.CheckTx.Log)
	}
	if res.DeliverTx.Code != uint32(0) {
		return res, errors.Errorf("DeliverTx failed: (%d) %s",
			res.DeliverTx.Code,
			res.DeliverTx.Log)
	}
	return res, err
}

// Query from Tendermint with the provided key and storename
func (ctx CoreContext) Query(key cmn.HexBytes, storeName string) (res []byte, err error) {

	path := fmt.Sprintf("/%s/key", storeName)
	node, err := ctx.GetNode()
	if err != nil {
		return res, err
	}

	opts := rpcclient.WRSPQueryOptions{
		Height:  ctx.Height,
		Trusted: ctx.TrustNode,
	}
	result, err := node.WRSPQueryWithOptions(path, key, opts)
	if err != nil {
		return res, err
	}
	resp := result.Response
	if resp.Code != uint32(0) {
		return res, errors.Errorf("Query failed: (%d) %s", resp.Code, resp.Log)
	}
	return resp.Value, nil
}

// Get the from address from the name flag
func (ctx CoreContext) GetFromAddress() (from sdk.Address, err error) {

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	name := ctx.FromAddressName
	if name == "" {
		return nil, errors.Errorf("must provide a from address name")
	}

	info, err := keybase.Get(name)
	if err != nil {
		return nil, errors.Errorf("No key for: %s", name)
	}

	return info.PubKey.Address(), nil
}

// sign and build the transaction from the msg
func (ctx CoreContext) SignAndBuild(name, passphrase string, msg sdk.Msg, cdc *wire.Codec) ([]byte, error) {

	// build the Sign Messsage from the Standard Message
	chainID := ctx.ChainID
	if chainID == "" {
		return nil, errors.Errorf("Chain ID required but not specified")
	}
	sequence := ctx.Sequence
	signMsg := sdk.StdSignMsg{
		ChainID:   chainID,
		Sequences: []int64{sequence},
		Msg:       msg,
	}

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	// sign and build
	bz := signMsg.Bytes()

	sig, pubkey, err := keybase.Sign(name, passphrase, bz)
	if err != nil {
		return nil, err
	}
	sigs := []sdk.StdSignature{{
		PubKey:    pubkey,
		Signature: sig,
		Sequence:  sequence,
	}}

	// marshal bytes
	tx := sdk.NewStdTx(signMsg.Msg, signMsg.Fee, sigs)

	return cdc.MarshalBinary(tx)
}

// sign and build the transaction from the msg
func (ctx CoreContext) SignBuildBroadcast(name string, msg sdk.Msg, cdc *wire.Codec) (*ctypes.ResultBroadcastTxCommit, error) {
	passphrase, err := ctx.GetPassphraseFromStdin(name)
	if err != nil {
		return nil, err
	}

	txBytes, err := ctx.SignAndBuild(name, passphrase, msg, cdc)
	if err != nil {
		return nil, err
	}

	return ctx.BroadcastTx(txBytes)
}

// get the next sequence for the account address
func (c CoreContext) NextSequence(address []byte) (int64, error) {
	if c.Decoder == nil {
		return 0, errors.New("AccountDecoder required but not provided")
	}

	res, err := c.Query(address, c.AccountStore)
	if err != nil {
		return 0, err
	}

	account, err := c.Decoder(res)
	if err != nil {
		panic(err)
	}

	return account.GetSequence(), nil
}

// get passphrase from std input
func (ctx CoreContext) GetPassphraseFromStdin(name string) (pass string, err error) {
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)
	return client.GetPassword(prompt, buf)
}

// GetNode prepares a simple rpc.Client
func (ctx CoreContext) GetNode() (rpcclient.Client, error) {
	if ctx.Client == nil {
		return nil, errors.New("Must define node URI")
	}
	return ctx.Client, nil
}
