package rest

import (
	"encoding/hex"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tepleton/go-crypto/keys"

	"github.com/tepleton/tepleton-sdk/client/context"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
	"github.com/tepleton/tepleton-sdk/x/bank/commands"
	"github.com/tepleton/tepleton-sdk/x/ibc"
)

type transferBody struct {
	// Fees             sdk.Coin  `json="fees"`
	Amount           sdk.Coins `json:"amount"`
	LocalAccountName string    `json:"name"`
	Password         string    `json:"password"`
	SrcChainID       string    `json:"src_chain_id"`
	Sequence         int64     `json:"sequence"`
}

// TransferRequestHandler - http request handler to transfer coins to a address
// on a different chain via IBC
func TransferRequestHandler(cdc *wire.Codec, kb keys.Keybase) func(http.ResponseWriter, *http.Request) {
	c := commands.Commander{cdc}
	ctx := context.NewCoreContextFromViper()
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		destChainID := vars["destchain"]
		address := vars["address"]

		var m transferBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = cdc.UnmarshalJSON(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		bz, err := hex.DecodeString(address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		to := sdk.Address(bz)

		// build message
		packet := ibc.NewIBCPacket(info.PubKey.Address(), to, m.Amount, m.SrcChainID, destChainID)
		msg := ibc.IBCTransferMsg{packet}

		// sign
		ctx = ctx.WithSequence(m.Sequence)
		txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, msg, c.Cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// send
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := cdc.MarshalJSON(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
