package proofs

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	wire "github.com/tepleton/go-wire"
	"github.com/tepleton/go-wire/data"
	lc "github.com/tepleton/light-client"
	"github.com/tepleton/light-client/proofs"
	"github.com/tepleton/merkleeyes/iavl"
	"github.com/tepleton/tepleton/rpc/client"

	"github.com/tepleton/basecoin/client/commands"
)

// GetParsed does most of the work of the query commands, but is quite
// opinionated, so if you want more control about parsing, call Get
// directly.
//
// It will try to get the proof for the given key.  If it is successful,
// it will return the height and also unserialize proof.Data into the data
// argument (so pass in a pointer to the appropriate struct)
func GetParsed(key []byte, data interface{}, prove bool) (uint64, error) {
	bs, h, err := Get(key, prove)
	if err != nil {
		return 0, err
	}
	err = wire.ReadBinaryBytes(bs, data)
	if err != nil {
		return 0, err
	}
	return h, nil
}

// Get queries the given key and returns the value stored there and the
// height we checked at.
//
// If prove is true (and why shouldn't it be?),
// the data is fully verified before returning.  If prove is false,
// we just repeat whatever any (potentially malicious) node gives us.
// Only use that if you are running the full node yourself,
// and it is localhost or you have a secure connection (not HTTP)
func Get(key []byte, prove bool) (data.Bytes, uint64, error) {
	if !prove {
		node := commands.GetNode()
		resp, err := node.WRSPQuery("/key", key, false)
		return data.Bytes(resp.Value), resp.Height, err
	}
	val, h, _, err := GetWithProof(key)
	return val, h, err
}

// GetWithProof returns the values stored under a given key at the named
// height as in Get.  Additionally, it will return a validated merkle
// proof for the key-value pair if it exists, and all checks pass.
func GetWithProof(key []byte) (data.Bytes, uint64, *iavl.KeyExistsProof, error) {
	node := commands.GetNode()

	resp, err := node.WRSPQuery("/key", key, true)
	if err != nil {
		return nil, 0, nil, err
	}
	ph := int(resp.Height)

	// make sure the proof is the proper height
	if !resp.Code.IsOK() {
		return nil, 0, nil, errors.Errorf("Query error %d: %s", resp.Code, resp.Code.String())
	}
	// TODO: Handle null proofs
	if len(resp.Key) == 0 || len(resp.Value) == 0 || len(resp.Proof) == 0 {
		return nil, 0, nil, lc.ErrNoData()
	}
	if ph != 0 && ph != int(resp.Height) {
		return nil, 0, nil, lc.ErrHeightMismatch(ph, int(resp.Height))
	}

	check, err := GetCertifiedCheckpoint(ph)
	if err != nil {
		return nil, 0, nil, err
	}

	proof := new(iavl.KeyExistsProof)
	err = wire.ReadBinaryBytes(resp.Proof, &proof)
	if err != nil {
		return nil, 0, nil, err
	}

	// validate the proof against the certified header to ensure data integrity
	err = proof.Verify(resp.Key, resp.Value, check.Header.AppHash)
	if err != nil {
		return nil, 0, nil, err
	}

	return data.Bytes(resp.Value), resp.Height, proof, nil
}

// GetCertifiedCheckpoint gets the signed header for a given height
// and certifies it.  Returns error if unable to get a proven header.
func GetCertifiedCheckpoint(h int) (empty lc.Checkpoint, err error) {
	// here is the certifier, root of all trust
	node := commands.GetNode()
	cert, err := commands.GetCertifier()
	if err != nil {
		return
	}

	// get the checkpoint for this height

	// FIXME: cannot use cert.GetByHeight for now, as it also requires
	// Validators and will fail on querying tepleton for non-current height.
	// When this is supported, we should use it instead...
	client.WaitForHeight(node, h, nil)
	commit, err := node.Commit(h)
	if err != nil {
		return
	}
	check := lc.Checkpoint{
		Header: commit.Header,
		Commit: commit.Commit,
	}

	// validate downloaded checkpoint with our request and trust store.
	if check.Height() != h {
		return empty, lc.ErrHeightMismatch(h, check.Height())
	}
	err = cert.Certify(check)
	return check, nil
}

// ParseHexKey parses the key flag as hex and converts to bytes or returns error
// argname is used to customize the error message
func ParseHexKey(args []string, argname string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.Errorf("Missing required argument [%s]", argname)
	}
	if len(args) > 1 {
		return nil, errors.Errorf("Only accepts one argument [%s]", argname)
	}
	rawkey := args[0]
	if rawkey == "" {
		return nil, errors.Errorf("[%s] argument must be non-empty ", argname)
	}
	// with tx, we always just parse key as hex and use to lookup
	return proofs.ParseHexKey(rawkey)
}

// GetHeight reads the viper config for the query height
func GetHeight() int {
	return viper.GetInt(FlagHeight)
}

type proof struct {
	Height uint64      `json:"height"`
	Data   interface{} `json:"data"`
}

// FoutputProof writes the output of wrapping height and info
// in the form {"data": <the_data>, "height": <the_height>}
// to the provider io.Writer
func FoutputProof(w io.Writer, v interface{}, height uint64) error {
	wrap := &proof{height, v}
	blob, err := data.ToJSON(wrap)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", blob)
	return err
}

// OutputProof prints the proof to stdout
// reuse this for printing proofs and we should enhance this for text/json,
// better presentation of height
func OutputProof(data interface{}, height uint64) error {
	return FoutputProof(os.Stdout, data, height)
}
