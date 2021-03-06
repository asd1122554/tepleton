package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tepleton/tepleton-sdk/client/builder"
	"github.com/tepleton/tepleton-sdk/wire"

	"github.com/tepleton/tepleton-sdk/examples/basecoin/x/cool"
)

// what cool transaction
func CoolTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cool [answer]",
		Short: "What's cooler than being cool?",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || len(args[0]) == 0 {
				return errors.New("You must provide an answer")
			}

			// get the from address from the name flag
			from, err := builder.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			msg := cool.NewQuizMsg(from, args[0])

			// build and sign the transaction, then broadcast to Tendermint
			res, err := builder.SignBuildBroadcast(msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
}

// set what cool transaction
func SetTrendTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "setcool [answer]",
		Short: "You're so cool, tell us what is cool!",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || len(args[0]) == 0 {
				return errors.New("You must provide an answer")
			}

			// get the from address from the name flag
			from, err := builder.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			msg := cool.NewSetTrendMsg(from, args[0])

			// build and sign the transaction, then broadcast to Tendermint
			res, err := builder.SignBuildBroadcast(msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
}
