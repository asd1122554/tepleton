package app

import (
	"testing"

	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/stretchr/testify/assert"
	crypto "github.com/tepleton/go-crypto"
)

func TestToAccount(t *testing.T) {
	priv = crypto.GenPrivKeyEd25519()
	addr = priv.PubKey().Address()
	authAcc := auth.NewBaseAccountWithAddress(addr)
	genAcc := NewGenesisAccount(authAcc)
	assert.Equal(t, authAcc, genAcc.ToAccount())
}

func TestGaiaAppGenTx(t *testing.T) {
	cdc := MakeCodec()

	//TODO test that key overwrite flags work / no overwrites if set off
	//TODO test validator created has provided pubkey
	//TODO test the account created has the correct pubkey
}

func TestGaiaAppGenState(t *testing.T) {
	cdc := MakeCodec()

	// TODO test must provide at least genesis transaction
	// TODO test with both one and two genesis transactions:
	// TODO        correct: genesis account created, canididates created, pool token variance
}
