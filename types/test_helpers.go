package types

// Helper functions for testing

import (
	"github.com/tepleton/go-crypto"
	cmn "github.com/tepleton/tmlibs/common"
)

// Creates a PrivAccount from secret.
// The amount is not set.
func PrivAccountFromSecret(secret string) PrivAccount {
	privKey :=
		crypto.GenPrivKeyEd25519FromSecret([]byte(secret)).Wrap()
	privAccount := PrivAccount{
		PrivKey: privKey,
		Account: Account{
			PubKey: privKey.PubKey(),
		},
	}
	return privAccount
}

// Make `num` random accounts
func RandAccounts(num int, minAmount int64, maxAmount int64) []PrivAccount {
	privAccs := make([]PrivAccount, num)
	for i := 0; i < num; i++ {

		balance := minAmount
		if maxAmount > minAmount {
			balance += cmn.RandInt64() % (maxAmount - minAmount)
		}

		privKey := crypto.GenPrivKeyEd25519().Wrap()
		pubKey := privKey.PubKey()
		privAccs[i] = PrivAccount{
			PrivKey: privKey,
			Account: Account{
				PubKey:  pubKey,
				Balance: Coins{Coin{"", balance}},
			},
		}
	}

	return privAccs
}

/////////////////////////////////////////////////////////////////

//func MakeAccs(secrets ...string) (accs []PrivAccount) {
//	for _, secret := range secrets {
//		privAcc := PrivAccountFromSecret(secret)
//		privAcc.Account.Balance = Coins{{"mycoin", 7}}
//		accs = append(accs, privAcc)
//	}
//	return
//}

func MakeAcc(secret string) PrivAccount {
	privAcc := PrivAccountFromSecret(secret)
	privAcc.Account.Balance = Coins{{"mycoin", 7}}
	return privAcc
}

func Accs2TxInputs(seq int, accs ...PrivAccount) []TxInput {
	var txs []TxInput
	for _, acc := range accs {
		tx := NewTxInput(
			acc.Account.PubKey,
			Coins{{"mycoin", 5}},
			seq)
		txs = append(txs, tx)
	}
	return txs
}

//turn a list of accounts into basic list of transaction outputs
func Accs2TxOutputs(accs ...PrivAccount) []TxOutput {
	var txs []TxOutput
	for _, acc := range accs {
		tx := TxOutput{
			acc.Account.PubKey.Address(),
			Coins{{"mycoin", 4}}}
		txs = append(txs, tx)
	}
	return txs
}

func MakeSendTx(seq int, accOut PrivAccount, accsIn ...PrivAccount) *SendTx {
	tx := &SendTx{
		Gas:     0,
		Fee:     Coin{"mycoin", 1},
		Inputs:  Accs2TxInputs(seq, accsIn...),
		Outputs: Accs2TxOutputs(accOut),
	}

	return tx
}

func SignTx(chainID string, tx *SendTx, accs ...PrivAccount) {
	signBytes := tx.SignBytes(chainID)
	for i, _ := range tx.Inputs {
		tx.Inputs[i].Signature = accs[i].Sign(signBytes)
	}
}
