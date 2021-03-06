package app

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	cmn "github.com/tepleton/tmlibs/common"
)

//nolint
const (
	ModuleNameBase = "base"
)

// Option just holds module/key/value triples from
// parsing the genesis file
type Option struct {
	Module string
	Key    string
	Value  string
}

// InitStater is anything that can handle app options
// from genesis file.
type InitStater interface {
	InitState(module, key, value string) (string, error)
}

// LoadGenesis parses the genesis file and sets the initial
// state based on that
func LoadGenesis(app InitStater, filePath string) error {
	opts, err := GetGenesisOptions(filePath)
	if err != nil {
		return err
	}

	// execute all the genesis init options
	// abort on any error
	for _, opt := range opts {
		_, err = app.InitState(opt.Module, opt.Key, opt.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetGenesisOptions parses the genesis file in a format
// that can easily be handed into InitStaters
func GetGenesisOptions(path string) ([]Option, error) {
	genDoc, err := loadGenesis(path)
	if err != nil {
		return nil, err
	}

	opts := genDoc.AppOptions
	cnt := 1 + len(opts.Accounts) + len(opts.pluginOptions)
	res := make([]Option, cnt)

	res[0] = Option{ModuleNameBase, ChainKey, genDoc.ChainID}
	i := 1

	// set accounts
	for _, acct := range opts.Accounts {
		res[i] = Option{"coin", "account", string(acct)}
		i++
	}

	// set plugin options
	for _, kv := range opts.pluginOptions {
		module, key := splitKey(kv.Key)
		res[i] = Option{module, key, kv.Value}
		i++
	}

	return res, nil
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FullGenesisDoc - includes tepleton (in the json, we ignore here)
type FullGenesisDoc struct {
	ChainID    string      `json:"chain_id"`
	AppOptions *GenesisDoc `json:"app_options"`
}

// GenesisDoc - All genesis values
type GenesisDoc struct {
	Accounts      []json.RawMessage `json:"accounts"`
	PluginOptions []json.RawMessage `json:"plugin_options"`

	pluginOptions []keyValue // unmarshaled rawmessages
}

func loadGenesis(filePath string) (*FullGenesisDoc, error) {
	bytes, err := cmn.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "loading genesis file")
	}

	// the basecoin genesis go-wire/data :)
	genDoc := new(FullGenesisDoc)
	err = json.Unmarshal(bytes, genDoc)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling genesis file")
	}

	if genDoc.AppOptions == nil {
		genDoc.AppOptions = new(GenesisDoc)
	}

	pluginOpts, err := parseGenesisList(genDoc.AppOptions.PluginOptions)
	if err != nil {
		return nil, err
	}
	genDoc.AppOptions.pluginOptions = pluginOpts
	return genDoc, nil
}

func parseGenesisList(kvzIn []json.RawMessage) (kvz []keyValue, err error) {
	if len(kvzIn)%2 != 0 {
		return nil, errors.New("genesis cannot have an odd number of items.  Format = [key1, value1, key2, value2, ...]")
	}

	for i := 0; i < len(kvzIn); i += 2 {
		kv := keyValue{}
		rawK := []byte(kvzIn[i])
		err := json.Unmarshal(rawK, &(kv.Key))
		if err != nil {
			return nil, errors.Errorf("Non-string key: %s", string(rawK))
		}
		// convert value to string if possible (otherwise raw json)
		rawV := kvzIn[i+1]
		err = json.Unmarshal(rawV, &(kv.Value))
		if err != nil {
			kv.Value = string(rawV)
		}
		kvz = append(kvz, kv)
	}
	return kvz, nil
}

// Splits the string at the first '/'.
// if there are none, assign default module ("base").
func splitKey(key string) (string, string) {
	if strings.Contains(key, "/") {
		keyParts := strings.SplitN(key, "/", 2)
		return keyParts[0], keyParts[1]
	}
	return ModuleNameBase, key
}
