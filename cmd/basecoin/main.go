package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tepleton/basecoin/cmd/commands"
	"github.com/tepleton/tmlibs/cli"
)

func main() {
	var RootCmd = &cobra.Command{
		Use:   "basecoin",
		Short: "A cryptocurrency framework in Golang based on Tendermint-Core",
	}

	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.RelayCmd,
		commands.TxCmd,
		commands.UnsafeResetAllCmd,
		commands.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(RootCmd, "BC", os.ExpandEnv("$HOME/.basecoin"))
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
