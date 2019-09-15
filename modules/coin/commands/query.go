package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	lc "github.com/tepleton/light-client"

	lcmd "github.com/tepleton/basecoin/client/commands"
	proofcmd "github.com/tepleton/basecoin/client/commands/proofs"
	"github.com/tepleton/basecoin/modules/auth"
	"github.com/tepleton/basecoin/modules/coin"
	"github.com/tepleton/basecoin/stack"
)

// AccountQueryCmd - command to query an account
var AccountQueryCmd = &cobra.Command{
	Use:   "account [address]",
	Short: "Get details of an account, with proof",
	RunE:  lcmd.RequireInit(accountQueryCmd),
}

func accountQueryCmd(cmd *cobra.Command, args []string) error {
	addr, err := proofcmd.ParseHexKey(args, "address")
	if err != nil {
		return err
	}
	key := stack.PrefixedKey(coin.NameCoin, auth.SigPerm(addr).Bytes())

	acc := coin.Account{}
	proof, err := proofcmd.GetAndParseAppProof(key, &acc)
	if lc.IsNoDataErr(err) {
		return errors.Errorf("Account bytes are empty for address %X ", addr)
	} else if err != nil {
		return err
	}

	return proofcmd.OutputProof(acc, proof.BlockHeight())
}