/*
Package commands contains any general setup/helpers valid for all subcommands
*/
package commands

import (
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tepleton/light-client/certifiers"
	"github.com/tepleton/light-client/certifiers/client"
	"github.com/tepleton/light-client/certifiers/files"
	"github.com/tepleton/tmlibs/cli"
	cmn "github.com/tepleton/tmlibs/common"

	rpcclient "github.com/tepleton/tepleton/rpc/client"

	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/modules/auth"
)

var (
	trustedProv certifiers.Provider
	sourceProv  certifiers.Provider
)

const (
	ChainFlag = "chain-id"
	NodeFlag  = "node"
)

// AddBasicFlags adds --node and --chain-id, which we need for everything
func AddBasicFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(ChainFlag, "", "Chain ID of tepleton node")
	cmd.PersistentFlags().String(NodeFlag, "", "<host>:<port> to tepleton rpc interface for this chain")
}

// GetChainID reads ChainID from the flags
func GetChainID() string {
	return viper.GetString(ChainFlag)
}

// GetNode prepares a simple rpc.Client from the flags
func GetNode() rpcclient.Client {
	return rpcclient.NewHTTP(viper.GetString(NodeFlag), "/websocket")
}

// GetProviders creates a trusted (local) seed provider and a remote
// provider based on configuration.
func GetProviders() (trusted certifiers.Provider, source certifiers.Provider) {
	if trustedProv == nil || sourceProv == nil {
		// initialize provider with files stored in homedir
		rootDir := viper.GetString(cli.HomeFlag)
		trustedProv = certifiers.NewCacheProvider(
			certifiers.NewMemStoreProvider(),
			files.NewProvider(rootDir),
		)
		node := viper.GetString(NodeFlag)
		sourceProv = client.NewHTTP(node)
	}
	return trustedProv, sourceProv
}

// GetCertifier constructs a dynamic certifier from the config info
func GetCertifier() (*certifiers.InquiringCertifier, error) {
	// load up the latest store....
	trust, source := GetProviders()

	// this gets the most recent verified seed
	seed, err := certifiers.LatestSeed(trust)
	if certifiers.IsSeedNotFoundErr(err) {
		return nil, errors.New("Please run init first to establish a root of trust")
	}
	if err != nil {
		return nil, err
	}
	cert := certifiers.NewInquiring(
		viper.GetString(ChainFlag), seed.Validators, trust, source)
	return cert, nil
}

// ParseActor parses an address of form:
// [<chain>:][<app>:]<hex address>
// into a basecoin.Actor.
// If app is not specified or "", then assume auth.NameSigs
func ParseActor(input string) (res basecoin.Actor, err error) {
	chain, app := "", auth.NameSigs
	input = strings.TrimSpace(input)
	spl := strings.SplitN(input, ":", 3)

	if len(spl) == 3 {
		chain = spl[0]
		spl = spl[1:]
	}
	if len(spl) == 2 {
		if spl[0] != "" {
			app = spl[0]
		}
		spl = spl[1:]
	}

	addr, err := hex.DecodeString(cmn.StripHex(spl[0]))
	if err != nil {
		return res, errors.Errorf("Address is invalid hex: %v\n", err)
	}
	res = basecoin.Actor{
		ChainID: chain,
		App:     app,
		Address: addr,
	}
	return
}

// ParseActors takes a comma-separated list of actors and parses them into
// a slice
func ParseActors(key string) (signers []basecoin.Actor, err error) {
	var act basecoin.Actor
	for _, k := range strings.Split(key, ",") {
		act, err = ParseActor(k)
		if err != nil {
			return
		}
		signers = append(signers, act)
	}
	return
}