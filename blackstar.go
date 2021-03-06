package blackstar

import (
	"github.com/tepleton/blackstar/types"
	"github.com/tepleton/go-crypto"
	"github.com/tepleton/go-wire"
	eyes "github.com/tepleton/merkleeyes/client"
	wrsp "github.com/tepleton/wrsp/types"
)

const version = "0.1"
const maxTxSize = 10240

type Blackstar struct {
	eyesCli *eyes.MerkleEyesClient
}

func NewBlackstar(eyesCli *eyes.MerkleEyesClient) *Blackstar {
	return &Blackstar{
		eyesCli: eyesCli,
	}
}

func (app *Blackstar) Info() string {
	return "Blackstar v" + version
}

type SetAccount struct {
	PubKey  crypto.PubKey
	Account types.Account
}

func (app *Blackstar) SetOption(key string, value string) (log string) {
	if key == "setAccount" {
		var err error
		var setAccount SetAccount
		wire.ReadJSONPtr(&setAccount, []byte(value), &err)
		if err != nil {
			return "Error decoding SetAccount message: " + err.Error()
		}
		pubKeyBytes := wire.BinaryBytes(setAccount.PubKey)
		accBytes := wire.BinaryBytes(setAccount.Account)
		err = app.eyesCli.SetSync(pubKeyBytes, accBytes)
		if err != nil {
			return "Error saving account: " + err.Error()
		}
		return "Success"
	}
	return "Unrecognized option key " + key
}

func (app *Blackstar) AppendTx(txBytes []byte) (code wrsp.CodeType, result []byte, log string) {
	if len(txBytes) > maxTxSize {
		return wrsp.CodeType_EncodingError, nil, "Tx size exceeds maximum"
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return wrsp.CodeType_EncodingError, nil, "Error decoding tx: " + err.Error()
	}
	// Validate tx
	code, errStr := validateTx(tx)
	if errStr != "" {
		return code, nil, "Error validating tx: " + errStr
	}
	// Load accounts
	accMap := loadAccounts(app.eyesCli, allPubKeys(tx))
	// Execute tx
	accs, code, errStr := execTx(tx, accMap)
	if errStr != "" {
		return code, nil, "Error executing tx: " + errStr
	}
	// Store accounts
	storeAccounts(app.eyesCli, accs)
	return wrsp.CodeType_OK, nil, "Success"
}

func (app *Blackstar) CheckTx(txBytes []byte) (code wrsp.CodeType, result []byte, log string) {
	if len(txBytes) > maxTxSize {
		return wrsp.CodeType_EncodingError, nil, "Tx size exceeds maximum"
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return wrsp.CodeType_EncodingError, nil, "Error decoding tx: " + err.Error()
	}
	// Validate tx
	code, errStr := validateTx(tx)
	if errStr != "" {
		return code, nil, "Error validating tx: " + errStr
	}
	// Load accounts
	accMap := loadAccounts(app.eyesCli, allPubKeys(tx))
	// Execute tx
	_, code, errStr = execTx(tx, accMap)
	if errStr != "" {
		return code, nil, "Error (mock) executing tx: " + errStr
	}
	return wrsp.CodeType_OK, nil, "Success"
}

func (app *Blackstar) Query(query []byte) (code wrsp.CodeType, result []byte, log string) {
	return wrsp.CodeType_OK, nil, ""
	value, err := app.eyesCli.GetSync(query)
	if err != nil {
		panic("Error making query: " + err.Error())
	}
	return wrsp.CodeType_OK, value, "Success"
}

func (app *Blackstar) GetHash() (hash []byte, log string) {
	hash, log, err := app.eyesCli.GetHashSync()
	if err != nil {
		panic("Error getting hash: " + err.Error())
	}
	return hash, "Success"
}

//----------------------------------------

func validateTx(tx types.Tx) (code wrsp.CodeType, errStr string) {
	if len(tx.Inputs) == 0 {
		return wrsp.CodeType_EncodingError, "Tx.Inputs length cannot be 0"
	}
	seenPubKeys := map[string]bool{}
	signBytes := txSignBytes(tx)
	for _, input := range tx.Inputs {
		code, errStr = validateInput(input, signBytes)
		if errStr != "" {
			return
		}
		keyString := input.PubKey.KeyString()
		if seenPubKeys[keyString] {
			return wrsp.CodeType_EncodingError, "Duplicate input pubKey"
		}
		seenPubKeys[keyString] = true
	}
	for _, output := range tx.Outputs {
		code, errStr = validateOutput(output)
		if errStr != "" {
			return
		}
		keyString := output.PubKey.KeyString()
		if seenPubKeys[keyString] {
			return wrsp.CodeType_EncodingError, "Duplicate output pubKey"
		}
		seenPubKeys[keyString] = true
	}
	sumInputs, overflow := sumAmounts(tx.Inputs, nil, 0)
	if overflow {
		return wrsp.CodeType_EncodingError, "Input amount overflow"
	}
	sumOutputsPlus, overflow := sumAmounts(nil, tx.Outputs, len(tx.Inputs)+len(tx.Outputs))
	if overflow {
		return wrsp.CodeType_EncodingError, "Output amount overflow"
	}
	if sumInputs < sumOutputsPlus {
		return wrsp.CodeType_InsufficientFees, "Insufficient fees"
	}
	return wrsp.CodeType_OK, ""
}

// NOTE: Tx is a struct, so it's copying the value
func txSignBytes(tx types.Tx) []byte {
	for i, input := range tx.Inputs {
		input.Signature = nil
		tx.Inputs[i] = input
	}
	return wire.BinaryBytes(tx)
}

func validateInput(input types.Input, signBytes []byte) (code wrsp.CodeType, errStr string) {
	if input.Amount == 0 {
		return wrsp.CodeType_EncodingError, "Input amount cannot be zero"
	}
	if input.PubKey == nil {
		return wrsp.CodeType_EncodingError, "Input pubKey cannot be nil"
	}
	if !input.PubKey.VerifyBytes(signBytes, input.Signature) {
		return wrsp.CodeType_Unauthorized, "Invalid ignature"
	}
	return wrsp.CodeType_OK, ""
}

func validateOutput(output types.Output) (code wrsp.CodeType, errStr string) {
	if output.Amount == 0 {
		return wrsp.CodeType_EncodingError, "Output amount cannot be zero"
	}
	if output.PubKey == nil {
		return wrsp.CodeType_EncodingError, "Output pubKey cannot be nil"
	}
	return wrsp.CodeType_OK, ""
}

func sumAmounts(inputs []types.Input, outputs []types.Output, more int) (total uint64, overflow bool) {
	total = uint64(more)
	for _, input := range inputs {
		total2 := total + input.Amount
		if total2 < total {
			return 0, true
		}
		total = total2
	}
	for _, output := range outputs {
		total2 := total + output.Amount
		if total2 < total {
			return 0, true
		}
		total = total2
	}
	return total, false
}

func allPubKeys(tx types.Tx) (pubKeys []crypto.PubKey) {
	pubKeys = make([]crypto.PubKey, len(tx.Inputs)+len(tx.Outputs))
	for _, input := range tx.Inputs {
		pubKeys = append(pubKeys, input.PubKey)
	}
	for _, output := range tx.Outputs {
		pubKeys = append(pubKeys, output.PubKey)
	}
	return pubKeys
}

// Returns accounts in order of types.Tx inputs and outputs
func execTx(tx types.Tx, accMap map[string]types.Account) (accs []types.Account, code wrsp.CodeType, errStr string) {
	accs = make([]types.Account, 0, len(tx.Inputs)+len(tx.Outputs))
	// Deduct from inputs
	for _, input := range tx.Inputs {
		var acc, ok = accMap[input.PubKey.KeyString()]
		if !ok {
			return nil, wrsp.CodeType_UnknownAccount, "Input account does not exist"
		}
		if acc.Sequence != input.Sequence {
			return nil, wrsp.CodeType_BadNonce, "Invalid sequence"
		}
		if acc.Balance < input.Amount {
			return nil, wrsp.CodeType_InsufficientFunds, "Insufficient funds"
		}
		// Good!
		acc.Sequence++
		acc.Balance -= input.Amount
		accs = append(accs, acc)
	}
	// Add to outputs
	for _, output := range tx.Outputs {
		var acc, ok = accMap[output.PubKey.KeyString()]
		if !ok {
			// Create new account if it doesn't already exist.
			acc = types.Account{
				PubKey:  output.PubKey,
				Balance: output.Amount,
			}
			accMap[output.PubKey.KeyString()] = acc
			continue
		}
		// Good!
		if (acc.Balance + output.Amount) < acc.Balance {
			return nil, wrsp.CodeType_InternalError, "Output balance overflow in execTx"
		}
		acc.Balance += output.Amount
		accs = append(accs, acc)
	}
	return accs, wrsp.CodeType_OK, ""
}

//----------------------------------------

func loadAccounts(eyesCli *eyes.MerkleEyesClient, pubKeys []crypto.PubKey) map[string]types.Account {
	return nil
}

func storeAccounts(eyesCli *eyes.MerkleEyesClient, accs []types.Account) {
}
