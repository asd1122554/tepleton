package tests

import (
	"github.com/tepleton/basecoin/types"
	. "github.com/tepleton/go-common"
	"github.com/tepleton/go-crypto"
)

// Creates a PrivAccount from secret.
// The amount is not set.
func PrivAccountFromSecret(secret string) types.PrivAccount {
	privKey := crypto.GenPrivKeyEd25519FromSecret([]byte(secret))
	privAccount := types.PrivAccount{
		PrivKey: privKey,
		PubKey:  privKey.PubKey(),
		Account: types.Account{
			Sequence: 0,
			Balance:  0,
		},
	}
	return privAccount
}

// Make `num` random accounts
func RandAccounts(num int, minAmount uint64, maxAmount uint64) []types.PrivAccount {
	privAccs := make([]types.PrivAccount, num)
	for i := 0; i < num; i++ {

		balance := minAmount
		if maxAmount > minAmount {
			balance += RandUint64() % (maxAmount - minAmount)
		}

		privKey := crypto.GenPrivKeyEd25519()
		pubKey := privKey.PubKey()
		privAccs[i] = types.PrivAccount{
			PrivKey: privKey,
			PubKey:  pubKey,
			Account: types.Account{
				Sequence: 0,
				Balance:  balance,
			},
		}
	}

	return privAccs
}