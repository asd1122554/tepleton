//nolint
package roles

import (
	"fmt"

	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/basecoin/errors"
)

var (
	errNoRole           = fmt.Errorf("No such role")
	errRoleExists       = fmt.Errorf("Role already exists")
	errNotMember        = fmt.Errorf("Not a member")
	errInsufficientSigs = fmt.Errorf("Not enough signatures")
	errNoMembers        = fmt.Errorf("No members specified")
	errTooManyMembers   = fmt.Errorf("Too many members specified")
)

func ErrNoRole() errors.TMError {
	return errors.WithCode(errNoRole, wrsp.CodeType_Unauthorized)
}
func IsNoRoleErr(err error) bool {
	return errors.IsSameError(errNoRole, err)
}

func ErrRoleExists() errors.TMError {
	return errors.WithCode(errRoleExists, wrsp.CodeType_Unauthorized)
}
func IsRoleExistsErr(err error) bool {
	return errors.IsSameError(errRoleExists, err)
}

func ErrNotMember() errors.TMError {
	return errors.WithCode(errNotMember, wrsp.CodeType_Unauthorized)
}
func IsNotMemberErr(err error) bool {
	return errors.IsSameError(errNotMember, err)
}

func ErrInsufficientSigs() errors.TMError {
	return errors.WithCode(errInsufficientSigs, wrsp.CodeType_Unauthorized)
}
func IsInsufficientSigsErr(err error) bool {
	return errors.IsSameError(errInsufficientSigs, err)
}

func ErrNoMembers() errors.TMError {
	return errors.WithCode(errNoMembers, wrsp.CodeType_Unauthorized)
}
func IsNoMembersErr(err error) bool {
	return errors.IsSameError(errNoMembers, err)
}

func ErrTooManyMembers() errors.TMError {
	return errors.WithCode(errTooManyMembers, wrsp.CodeType_Unauthorized)
}
func IsTooManyMembersErr(err error) bool {
	return errors.IsSameError(errTooManyMembers, err)
}