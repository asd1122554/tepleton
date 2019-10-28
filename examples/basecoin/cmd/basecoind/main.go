package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/tmlibs/cli"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/tepleton-sdk/examples/basecoin/app"
	"github.com/tepleton/tepleton-sdk/server"
)

// basecoindCmd is the entry point for this binary
var (
	context = server.NewContext(nil, nil)
	rootCmd = &cobra.Command{
		Use:               "basecoind",
		Short:             "Basecoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(context),
	}
)

func generateApp(rootDir string, logger log.Logger) (wrsp.Application, error) {
	dataDir := filepath.Join(rootDir, "data")
	dbMain, err := dbm.NewGoLevelDB("basecoin", dataDir)
	if err != nil {
		return nil, err
	}
	dbAcc, err := dbm.NewGoLevelDB("basecoin-acc", dataDir)
	if err != nil {
		return nil, err
	}
	dbIBC, err := dbm.NewGoLevelDB("basecoin-ibc", dataDir)
	if err != nil {
		return nil, err
	}
	dbStaking, err := dbm.NewGoLevelDB("basecoin-staking", dataDir)
	if err != nil {
		return nil, err
	}
	dbs := map[string]dbm.DB{
		"main":    dbMain,
		"acc":     dbAcc,
		"ibc":     dbIBC,
		"staking": dbStaking,
	}
	bapp := app.NewBasecoinApp(logger, dbs)
	return bapp, nil
}

func main() {
	server.AddCommands(rootCmd, server.DefaultGenAppState, generateApp, context)

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.basecoind")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	executor.Execute()
}
