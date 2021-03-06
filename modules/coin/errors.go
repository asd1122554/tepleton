//nolint
package coin

import (
	"fmt"

	wrsp "github.com/tepleton/wrsp/types"

	"github.com/tepleton/tepleton-sdk/errors"
)

var (
	errNoAccount          = fmt.Errorf("No such account")
	errInsufficientFunds  = fmt.Errorf("Insufficient funds")
	errInsufficientCredit = fmt.Errorf("Insufficient credit")
	errNoInputs           = fmt.Errorf("No input coins")
	errNoOutputs          = fmt.Errorf("No output coins")
	errInvalidAddress     = fmt.Errorf("Invalid address")
	errInvalidCoins       = fmt.Errorf("Invalid coins")

	invalidInput   = wrsp.CodeType_BaseInvalidInput
	invalidOutput  = wrsp.CodeType_BaseInvalidOutput
	unknownAddress = wrsp.CodeType_BaseUnknownAddress
	unknownRequest = wrsp.CodeType_UnknownRequest
)

// here are some generic handlers to grab classes of errors based on code
func IsInputErr(err error) bool {
	return errors.HasErrorCode(err, invalidInput)
}
func IsOutputErr(err error) bool {
	return errors.HasErrorCode(err, invalidOutput)
}
func IsAddressErr(err error) bool {
	return errors.HasErrorCode(err, unknownAddress)
}
func IsCoinErr(err error) bool {
	return err != nil && (IsInputErr(err) || IsOutputErr(err) || IsAddressErr(err))
}

func ErrNoAccount() errors.TMError {
	return errors.WithCode(errNoAccount, unknownAddress)
}

func IsNoAccountErr(err error) bool {
	return errors.IsSameError(errNoAccount, err)
}

func ErrInvalidAddress() errors.TMError {
	return errors.WithCode(errInvalidAddress, invalidInput)
}
func IsInvalidAddressErr(err error) bool {
	return errors.IsSameError(errInvalidAddress, err)
}

func ErrInvalidCoins() errors.TMError {
	return errors.WithCode(errInvalidCoins, invalidInput)
}
func IsInvalidCoinsErr(err error) bool {
	return errors.IsSameError(errInvalidCoins, err)
}

func ErrInsufficientFunds() errors.TMError {
	return errors.WithCode(errInsufficientFunds, invalidInput)
}
func IsInsufficientFundsErr(err error) bool {
	return errors.IsSameError(errInsufficientFunds, err)
}

func ErrInsufficientCredit() errors.TMError {
	return errors.WithCode(errInsufficientCredit, invalidInput)
}
func IsInsufficientCreditErr(err error) bool {
	return errors.IsSameError(errInsufficientCredit, err)
}

func ErrNoInputs() errors.TMError {
	return errors.WithCode(errNoInputs, invalidInput)
}
func IsNoInputsErr(err error) bool {
	return errors.IsSameError(errNoInputs, err)
}

func ErrNoOutputs() errors.TMError {
	return errors.WithCode(errNoOutputs, invalidOutput)
}
func IsNoOutputsErr(err error) bool {
	return errors.IsSameError(errNoOutputs, err)
}
