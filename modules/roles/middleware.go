package roles

import (
	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/stack"
	"github.com/tepleton/basecoin/state"
)

// Middleware allows us to add a requested role as a permission
// if the tx requests it and has sufficient authority
type Middleware struct {
	stack.PassOption
}

var _ stack.Middleware = Middleware{}

// NewMiddleware creates a role-checking middleware
func NewMiddleware() Middleware {
	return Middleware{}
}

// Name - return name space
func (Middleware) Name() string {
	return NameRole
}

// CheckTx tries to assume the named role if requested.
// If no role is requested, do nothing.
// If insufficient authority to assume the role, return error.
func (m Middleware) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	// if this is not an AssumeRoleTx, then continue
	assume, ok := tx.Unwrap().(AssumeRoleTx)
	if !ok { // this also breaks the recursion below
		return next.CheckTx(ctx, store, tx)
	}

	ctx, err = assumeRole(ctx, store, assume)
	if err != nil {
		return res, err
	}

	// one could add multiple role statements, repeat as needed
	return m.CheckTx(ctx, store, assume.Tx, next)
}

// DeliverTx tries to assume the named role if requested.
// If no role is requested, do nothing.
// If insufficient authority to assume the role, return error.
func (m Middleware) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	// if this is not an AssumeRoleTx, then continue
	assume, ok := tx.Unwrap().(AssumeRoleTx)
	if !ok { // this also breaks the recursion below
		return next.DeliverTx(ctx, store, tx)
	}

	ctx, err = assumeRole(ctx, store, assume)
	if err != nil {
		return res, err
	}

	// one could add multiple role statements, repeat as needed
	return m.DeliverTx(ctx, store, assume.Tx, next)
}

func assumeRole(ctx basecoin.Context, store state.KVStore, assume AssumeRoleTx) (basecoin.Context, error) {
	err := assume.ValidateBasic()
	if err != nil {
		return nil, err
	}

	role, err := loadRole(store, assume.Role)
	if err != nil {
		return nil, err
	}

	if !role.IsAuthorized(ctx) {
		return nil, ErrInsufficientSigs()
	}
	ctx = ctx.WithPermissions(NewPerm(assume.Role))
	return ctx, nil
}
