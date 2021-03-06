package stack

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/state"
	"github.com/tepleton/basecoin/txs"
)

func TestChain(t *testing.T) {
	assert := assert.New(t)
	msg := "got it"
	chainID := "my-chain"

	raw := txs.NewRaw([]byte{1, 2, 3, 4})
	cases := []struct {
		tx       basecoin.Tx
		valid    bool
		errorMsg string
	}{
		{txs.NewChain(chainID, raw), true, ""},
		{txs.NewChain("someone-else", raw), false, "someone-else"},
		{raw, false, "No chain id provided"},
	}

	// generic args here...
	ctx := NewContext(chainID, log.NewNopLogger())
	store := state.NewMemKVStore()

	// build the stack
	ok := OKHandler{Log: msg}
	app := New(Chain{}).Use(ok)

	for idx, tc := range cases {
		i := strconv.Itoa(idx)

		// make sure check returns error, not a panic crash
		res, err := app.CheckTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", idx, err)
			assert.Equal(msg, res.Log, i)
		} else {
			if assert.NotNil(err, i) {
				assert.Contains(err.Error(), tc.errorMsg, i)
			}
		}

		// make sure deliver returns error, not a panic crash
		res, err = app.DeliverTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", idx, err)
			assert.Equal(msg, res.Log, i)
		} else {
			if assert.NotNil(err, i) {
				assert.Contains(err.Error(), tc.errorMsg, i)
			}
		}
	}
}
