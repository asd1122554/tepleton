package assoc

import (
	"bytes"

	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
)

// ValidatorSet defines
type ValidatorSet struct {
	sdk.ValidatorSet

	key sdk.KVStoreGetter
	cdc *wire.Codec

	maxAssoc int
	addrLen  int
}

var _ sdk.ValidatorSet = ValidatorSet{}

// NewValidatorSet returns new ValidatorSet with underlying ValidatorSet
func NewValidatorSet(cdc *wire.Codec, key sdk.KVStoreGetter, valset sdk.ValidatorSet, maxAssoc int, addrLen int) ValidatorSet {
	if maxAssoc < 0 || addrLen < 0 {
		panic("Cannot use negative integer for NewValidatorSet")
	}
	return ValidatorSet{
		ValidatorSet: valset,

		key: key,
		cdc: cdc,

		maxAssoc: maxAssoc,
		addrLen:  addrLen,
	}
}

// Implements sdk.ValidatorSet
func (valset ValidatorSet) Validator(ctx sdk.Context, addr sdk.Address) (res sdk.Validator) {
	store := valset.key.KVStore(ctx)
	base := store.Get(GetBaseKey(addr))
	res = valset.ValidatorSet.Validator(ctx, base)
	if res == nil {
		res = valset.ValidatorSet.Validator(ctx, addr)
	}
	return
}

// GetBaseKey :: sdk.Address -> sdk.Address
func GetBaseKey(addr sdk.Address) []byte {
	return append([]byte{0x00}, addr...)
}

// GetAssocPrefix :: sdk.Address -> (sdk.Address -> byte)
func GetAssocPrefix(base sdk.Address) []byte {
	return append([]byte{0x01}, base...)
}

// GetAssocKey :: (sdk.Address, sdk.Address) -> byte
func GetAssocKey(base sdk.Address, assoc sdk.Address) []byte {
	return append(append([]byte{0x01}, base...), assoc...)
}

// Associate associates new address with validator address
func (valset ValidatorSet) Associate(ctx sdk.Context, base sdk.Address, assoc sdk.Address) bool {
	if len(base) != valset.addrLen || len(assoc) != valset.addrLen {
		return false
	}
	store := valset.key.KVStore(ctx)
	// If someone already owns the associated address
	if store.Get(GetBaseKey(assoc)) != nil {
		return false
	}
	store.Set(GetBaseKey(assoc), base)
	store.Set(GetAssocKey(base, assoc), []byte{0x00})
	return true
}

// Dissociate removes association between addresses
func (valset ValidatorSet) Dissociate(ctx sdk.Context, base sdk.Address, assoc sdk.Address) bool {
	if len(base) != valset.addrLen || len(assoc) != valset.addrLen {
		return false
	}
	store := valset.key.KVStore(ctx)
	// No associated address found for given validator
	if !bytes.Equal(store.Get(GetBaseKey(assoc)), base) {
		return false
	}
	store.Delete(GetBaseKey(assoc))
	store.Delete(GetAssocKey(base, assoc))
	return true
}

// Associations returns all associated addresses with a validator
func (valset ValidatorSet) Associations(ctx sdk.Context, base sdk.Address) (res []sdk.Address) {
	store := valset.key.KVStore(ctx)
	res = make([]sdk.Address, valset.maxAssoc)
	iter := sdk.KVStorePrefixIterator(store, GetAssocPrefix(base))
	i := 0
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		res[i] = key[len(key)-valset.addrLen:]
		i++
	}
	return res[:i]
}
