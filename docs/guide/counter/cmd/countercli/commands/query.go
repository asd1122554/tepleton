package commands

import (
	"github.com/spf13/cobra"

	proofcmd "github.com/tepleton/basecoin/client/commands/proofs"

	"github.com/tepleton/basecoin/docs/guide/counter/plugins/counter"
	"github.com/tepleton/basecoin/stack"
)

//CounterQueryCmd - CLI command to query the counter state
var CounterQueryCmd = &cobra.Command{
	Use:   "counter",
	Short: "Query counter state, with proof",
	RunE:  counterQueryCmd,
}

func counterQueryCmd(cmd *cobra.Command, args []string) error {
	key := stack.PrefixedKey(counter.NameCounter, counter.StateKey())

	var cp counter.State
	proof, err := proofcmd.GetAndParseAppProof(key, &cp)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(cp, proof.BlockHeight())
}
