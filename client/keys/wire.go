package keys

import (
	"github.com/tepleton/tepleton-sdk/wire"
)

var cdc *wire.Codec

func init() {
	cdc = wire.NewCodec()
	wire.RegisterCrypto(cdc)
}

func MarshalJSON(o interface{}) ([]byte, error) {
	return cdc.MarshalJSON(o)
}
