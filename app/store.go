package app

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/iavl"
	cmn "github.com/tepleton/tmlibs/common"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	sdk "github.com/tepleton/tepleton-sdk"
	"github.com/tepleton/tepleton-sdk/errors"
	sm "github.com/tepleton/tepleton-sdk/state"
)

// DefaultHistorySize is how many blocks of history to store for WRSP queries
const DefaultHistorySize = 10

// StoreApp contains a data store and all info needed
// to perform queries and handshakes.
//
// It should be embeded in another struct for CheckTx,
// DeliverTx and initializing state from the genesis.
type StoreApp struct {
	// Name is what is returned from info
	Name string

	// this is the database state
	info  *sm.ChainState
	state *sm.State

	// cached validator changes from DeliverTx
	pending []*wrsp.Validator

	// height is last committed block, DeliverTx is the next one
	height uint64

	logger log.Logger
}

// NewStoreApp creates a data store to handle queries
func NewStoreApp(appName, dbName string, cacheSize int, logger log.Logger) (*StoreApp, error) {
	state, err := loadState(dbName, cacheSize, DefaultHistorySize)
	if err != nil {
		return nil, err
	}
	app := &StoreApp{
		Name:   appName,
		state:  state,
		height: state.LatestHeight(),
		info:   sm.NewChainState(),
		logger: logger.With("module", "app"),
	}
	return app, nil
}

// MockStoreApp returns a Store app with no persistence
func MockStoreApp(appName string, logger log.Logger) (*StoreApp, error) {
	return NewStoreApp(appName, "", 0, logger)
}

// GetChainID returns the currently stored chain
func (app *StoreApp) GetChainID() string {
	return app.info.GetChainID(app.state.Committed())
}

// Logger returns the application base logger
func (app *StoreApp) Logger() log.Logger {
	return app.logger
}

// Hash gets the last hash stored in the database
func (app *StoreApp) Hash() []byte {
	return app.state.LatestHash()
}

// Append returns the working state for DeliverTx
func (app *StoreApp) Append() sdk.SimpleDB {
	return app.state.Append()
}

// Check returns the working state for CheckTx
func (app *StoreApp) Check() sdk.SimpleDB {
	return app.state.Check()
}

// CommittedHeight gets the last block height committed
// to the db
func (app *StoreApp) CommittedHeight() uint64 {
	return app.height
}

// WorkingHeight gets the current block we are writing
func (app *StoreApp) WorkingHeight() uint64 {
	return app.height + 1
}

// Info implements wrsp.Application. It returns the height and hash,
// as well as the wrsp name and version.
//
// The height is the block that holds the transactions, not the apphash itself.
func (app *StoreApp) Info(req wrsp.RequestInfo) wrsp.ResponseInfo {
	hash := app.Hash()

	app.logger.Info("Info synced",
		"height", app.CommittedHeight(),
		"hash", fmt.Sprintf("%X", hash))

	return wrsp.ResponseInfo{
		Data:             app.Name,
		LastBlockHeight:  app.CommittedHeight(),
		LastBlockAppHash: hash,
	}
}

// SetOption - WRSP
func (app *StoreApp) SetOption(key string, value string) string {
	return "Not Implemented"
}

// Query - WRSP
func (app *StoreApp) Query(reqQuery wrsp.RequestQuery) (resQuery wrsp.ResponseQuery) {
	if len(reqQuery.Data) == 0 {
		resQuery.Log = "Query cannot be zero length"
		resQuery.Code = wrsp.CodeType_EncodingError
		return
	}

	// set the query response height to current
	tree := app.state.Committed()

	height := reqQuery.Height
	if height == 0 {
		// TODO: once the rpc actually passes in non-zero
		// heights we can use to query right after a tx
		// we must retrun most recent, even if apphash
		// is not yet in the blockchain

		withProof := app.CommittedHeight() - 1
		if tree.Tree.VersionExists(withProof) {
			height = withProof
		} else {
			height = app.CommittedHeight()
		}
	}
	resQuery.Height = height

	switch reqQuery.Path {
	case "/store", "/key": // Get by key
		key := reqQuery.Data // Data holds the key bytes
		resQuery.Key = key
		if reqQuery.Prove {
			value, proof, err := tree.GetVersionedWithProof(key, height)
			if err != nil {
				resQuery.Log = err.Error()
				break
			}
			resQuery.Value = value
			resQuery.Proof = proof.Bytes()
		} else {
			value := tree.Get(key)
			resQuery.Value = value
		}

	default:
		resQuery.Code = wrsp.CodeType_UnknownRequest
		resQuery.Log = cmn.Fmt("Unexpected Query path: %v", reqQuery.Path)
	}
	return
}

// Commit implements wrsp.Application
func (app *StoreApp) Commit() (res wrsp.Result) {
	app.height++

	hash, err := app.state.Commit(app.height)
	if err != nil {
		// die if we can't commit, not to recover
		panic(err)
	}
	app.logger.Debug("Commit synced",
		"height", app.height,
		"hash", fmt.Sprintf("%X", hash),
	)

	if app.state.Size() == 0 {
		return wrsp.NewResultOK(nil, "Empty hash for empty tree")
	}
	return wrsp.NewResultOK(hash, "")
}

// InitChain - WRSP
func (app *StoreApp) InitChain(req wrsp.RequestInitChain) {}

// BeginBlock - WRSP
func (app *StoreApp) BeginBlock(req wrsp.RequestBeginBlock) {}

// EndBlock - WRSP
// Returns a list of all validator changes made in this block
func (app *StoreApp) EndBlock(height uint64) (res wrsp.ResponseEndBlock) {
	// TODO: cleanup in case a validator exists multiple times in the list
	res.Diffs = app.pending
	app.pending = nil
	return
}

// AddValChange is meant to be called by apps on DeliverTx
// results, this is added to the cache for the endblock
// changeset
func (app *StoreApp) AddValChange(diffs []*wrsp.Validator) {
	for _, d := range diffs {
		idx := pubKeyIndex(d, app.pending)
		if idx >= 0 {
			app.pending[idx] = d
		} else {
			app.pending = append(app.pending, d)
		}
	}
}

// return index of list with validator of same PubKey, or -1 if no match
func pubKeyIndex(val *wrsp.Validator, list []*wrsp.Validator) int {
	for i, v := range list {
		if bytes.Equal(val.PubKey, v.PubKey) {
			return i
		}
	}
	return -1
}

func loadState(dbName string, cacheSize int, historySize uint64) (*sm.State, error) {
	// memory backed case, just for testing
	if dbName == "" {
		tree := iavl.NewVersionedTree(0, dbm.NewMemDB())
		return sm.NewState(tree, historySize), nil
	}

	// Expand the path fully
	dbPath, err := filepath.Abs(dbName)
	if err != nil {
		return nil, errors.ErrInternal("Invalid Database Name")
	}

	// Some external calls accidently add a ".db", which is now removed
	dbPath = strings.TrimSuffix(dbPath, path.Ext(dbPath))

	// Split the database name into it's components (dir, name)
	dir := path.Dir(dbPath)
	name := path.Base(dbPath)

	// Open database called "dir/name.db", if it doesn't exist it will be created
	db := dbm.NewDB(name, dbm.LevelDBBackendStr, dir)
	tree := iavl.NewVersionedTree(cacheSize, db)
	if err = tree.Load(); err != nil {
		return nil, errors.ErrInternal("Loading tree: " + err.Error())
	}

	return sm.NewState(tree, historySize), nil
}
