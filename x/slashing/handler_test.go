package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/x/stake"
)

func TestCannotUnrevokeUnlessRevoked(t *testing.T) {
	// initial setup
	ctx, ck, sk, keeper := createTestInput(t)
	slh := NewHandler(keeper)
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	got := stake.NewHandler(sk)(ctx, newTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
	require.True(t, sdk.NewRatFromInt(amt).Equal(sk.Validator(ctx, addr).GetPower()))

	// assert non-revoked validator can't be unrevoked
	got = slh(ctx, NewMsgUnrevoke(addr))
	require.False(t, got.IsOK(), "allowed unrevoke of non-revoked validator")
	require.Equal(t, sdk.ToWRSPCode(DefaultCodespace, CodeValidatorNotRevoked), got.Code)
}
