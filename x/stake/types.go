package stake

import (
	"bytes"

	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
	wrsp "github.com/tepleton/wrsp/types"
	crypto "github.com/tepleton/go-crypto"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Pool       Pool         `json:"pool"`
	Params     Params       `json:"params"`
	Validators []Validator  `json:"validators"`
	Bonds      []Delegation `json:"bonds"`
}

func NewGenesisState(pool Pool, params Params, validators []Validator, bonds []Delegation) GenesisState {
	return GenesisState{
		Pool:       pool,
		Params:     params,
		Validators: validators,
		Bonds:      bonds,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Pool:   initialPool(),
		Params: defaultParams(),
	}
}

//_________________________________________________________________________

// Params defines the high level settings for staking
type Params struct {
	InflationRateChange sdk.Rat `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Rat `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Rat `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Rat `json:"goal_bonded"`           // Goal of percent bonded atoms

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination
}

func (p Params) equal(p2 Params) bool {
	return p.InflationRateChange.Equal(p2.InflationRateChange) &&
		p.InflationMax.Equal(p2.InflationMax) &&
		p.InflationMin.Equal(p2.InflationMin) &&
		p.GoalBonded.Equal(p2.GoalBonded) &&
		p.MaxValidators == p2.MaxValidators &&
		p.BondDenom == p2.BondDenom
}

func defaultParams() Params {
	return Params{
		InflationRateChange: sdk.NewRat(13, 100),
		InflationMax:        sdk.NewRat(20, 100),
		InflationMin:        sdk.NewRat(7, 100),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}

//_________________________________________________________________________

// Pool - dynamic parameters of the current state
type Pool struct {
	TotalSupply       int64   `json:"total_supply"`        // total supply of all tokens
	BondedShares      sdk.Rat `json:"bonded_shares"`       // sum of all shares distributed for the Bonded Pool
	UnbondingShares   sdk.Rat `json:"unbonding_shares"`    // shares moving from Bonded to Unbonded Pool
	UnbondedShares    sdk.Rat `json:"unbonded_shares"`     // sum of all shares distributed for the Unbonded Pool
	BondedPool        int64   `json:"bonded_pool"`         // reserve of bonded tokens
	UnbondingPool     int64   `json:"unbonding_pool"`      // tokens moving from bonded to unbonded pool
	UnbondedPool      int64   `json:"unbonded_pool"`       // reserve of unbonded tokens held with validators
	InflationLastTime int64   `json:"inflation_last_time"` // block which the last inflation was processed // TODO make time
	Inflation         sdk.Rat `json:"inflation"`           // current annual inflation rate

	DateLastCommissionReset int64 `json:"date_last_commission_reset"` // unix timestamp for last commission accounting reset (daily)

	// Fee Related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // last recorded bonded shares - for fee calcualtions
}

func (p Pool) equal(p2 Pool) bool {
	return p.TotalSupply == p2.TotalSupply &&
		p.BondedShares.Equal(p2.BondedShares) &&
		p.UnbondedShares.Equal(p2.UnbondedShares) &&
		p.BondedPool == p2.BondedPool &&
		p.UnbondedPool == p2.UnbondedPool &&
		p.InflationLastTime == p2.InflationLastTime &&
		p.Inflation.Equal(p2.Inflation) &&
		p.DateLastCommissionReset == p2.DateLastCommissionReset &&
		p.PrevBondedShares.Equal(p2.PrevBondedShares)
}

// initial pool for testing
func initialPool() Pool {
	return Pool{
		TotalSupply:             0,
		BondedShares:            sdk.ZeroRat(),
		UnbondingShares:         sdk.ZeroRat(),
		UnbondedShares:          sdk.ZeroRat(),
		BondedPool:              0,
		UnbondingPool:           0,
		UnbondedPool:            0,
		InflationLastTime:       0,
		Inflation:               sdk.NewRat(7, 100),
		DateLastCommissionReset: 0,
		PrevBondedShares:        sdk.ZeroRat(),
	}
}

//_________________________________________________________________________

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// validator, the validator is credited with a Delegation whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
type Validator struct {
	Status          sdk.ValidatorStatus `json:"status"`           // Bonded status
	Address         sdk.Address         `json:"address"`          // Sender of BondTx - UnbondTx returns here
	PubKey          crypto.PubKey       `json:"pub_key"`          // Pubkey of validator
	BondedShares    sdk.Rat             `json:"bonded_shares"`    // total shares of bonded global hold pool
	UnbondingShares sdk.Rat             `json:"unbonding_shares"` // total shares of unbonding global hold pool
	UnbondedShares  sdk.Rat             `json:"unbonded_shares"`  // total shares of unbonded global hold pool
	DelegatorShares sdk.Rat             `json:"liabilities"`      // total shares issued to a validator's delegators

	Description        Description `json:"description"`            // Description terms for the validator
	BondHeight         int64       `json:"validator_bond_height"`  // Earliest height as a bonded validator
	BondIntraTxCounter int16       `json:"validator_bond_counter"` // Block-local tx index of validator change
	ProposerRewardPool sdk.Coins   `json:"proposer_reward_pool"`   // XXX reward pool collected from being the proposer

	Commission            sdk.Rat `json:"commission"`              // XXX the commission rate of fees charged to any delegators
	CommissionMax         sdk.Rat `json:"commission_max"`          // XXX maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Rat `json:"commission_change_rate"`  // XXX maximum daily increase of the validator commission
	CommissionChangeToday sdk.Rat `json:"commission_change_today"` // XXX commission rate change today, reset each day (UTC time)

	// fee related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // total shares of a global hold pools
}

// Validators - list of Validators
type Validators []Validator

// NewValidator - initialize a new validator
func NewValidator(address sdk.Address, pubKey crypto.PubKey, description Description) Validator {
	return Validator{
		Status:                sdk.Unbonded,
		Address:               address,
		PubKey:                pubKey,
		BondedShares:          sdk.ZeroRat(),
		DelegatorShares:       sdk.ZeroRat(),
		Description:           description,
		BondHeight:            int64(0),
		BondIntraTxCounter:    int16(0),
		ProposerRewardPool:    sdk.Coins{},
		Commission:            sdk.ZeroRat(),
		CommissionMax:         sdk.ZeroRat(),
		CommissionChangeRate:  sdk.ZeroRat(),
		CommissionChangeToday: sdk.ZeroRat(),
		PrevBondedShares:      sdk.ZeroRat(),
	}
}

func (v Validator) equal(c2 Validator) bool {
	return v.Status == c2.Status &&
		v.PubKey.Equals(c2.PubKey) &&
		bytes.Equal(v.Address, c2.Address) &&
		v.BondedShares.Equal(c2.BondedShares) &&
		v.DelegatorShares.Equal(c2.DelegatorShares) &&
		v.Description == c2.Description &&
		v.BondHeight == c2.BondHeight &&
		//v.BondIntraTxCounter == c2.BondIntraTxCounter && // counter is always changing
		v.ProposerRewardPool.IsEqual(c2.ProposerRewardPool) &&
		v.Commission.Equal(c2.Commission) &&
		v.CommissionMax.Equal(c2.CommissionMax) &&
		v.CommissionChangeRate.Equal(c2.CommissionChangeRate) &&
		v.CommissionChangeToday.Equal(c2.CommissionChangeToday) &&
		v.PrevBondedShares.Equal(c2.PrevBondedShares)
}

// Description - description fields for a validator
type Description struct {
	Moniker  string `json:"moniker"`
	Identity string `json:"identity"`
	Website  string `json:"website"`
	Details  string `json:"details"`
}

func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

// get the exchange rate of global pool shares over delegator shares
func (v Validator) delegatorShareExRate() sdk.Rat {
	if v.DelegatorShares.IsZero() {
		return sdk.OneRat()
	}
	switch v.Status {
	case sdk.Bonded:
		return v.BondedShares.Quo(v.DelegatorShares)
	case sdk.Unbonding:
		return v.UnbondingShares.Quo(v.DelegatorShares)
	default: //sdk.Unbonded, sdk.Revoked:
		return v.UnbondedShares.Quo(v.DelegatorShares)
	}
}

// wrsp validator from stake validator type
func (v Validator) wrspValidator(cdc *wire.Codec) wrsp.Validator {
	return wrsp.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  v.BondedShares.Evaluate(),
	}
}

// wrsp validator from stake validator type
// with zero power used for validator updates
func (v Validator) wrspValidatorZero(cdc *wire.Codec) wrsp.Validator {
	return wrsp.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  0,
	}
}

//XXX updateDescription function
//XXX enforce limit to number of description characters

//______________________________________________________________________

// ensure fulfills the sdk validator types
var _ sdk.Validator = Validator{}

// nolint - for sdk.Validator
func (v Validator) GetStatus() sdk.ValidatorStatus { return v.Status }
func (v Validator) GetAddress() sdk.Address        { return v.Address }
func (v Validator) GetPubKey() crypto.PubKey       { return v.PubKey }
func (v Validator) GetPower() sdk.Rat              { return v.BondedShares }
func (v Validator) GetBondHeight() int64           { return v.BondHeight }

//_________________________________________________________________________

// Delegation represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
// TODO better way of managing space
type Delegation struct {
	DelegatorAddr sdk.Address `json:"delegator_addr"`
	ValidatorAddr sdk.Address `json:"validator_addr"`
	Shares        sdk.Rat     `json:"shares"`
	Height        int64       `json:"height"` // Last height bond updated
}

func (b Delegation) equal(b2 Delegation) bool {
	return bytes.Equal(b.DelegatorAddr, b2.DelegatorAddr) &&
		bytes.Equal(b.ValidatorAddr, b2.ValidatorAddr) &&
		b.Height == b2.Height &&
		b.Shares.Equal(b2.Shares)
}

// ensure fulfills the sdk validator types
var _ sdk.Delegation = Delegation{}

// nolint - for sdk.Delegation
func (b Delegation) GetDelegator() sdk.Address { return b.DelegatorAddr }
func (b Delegation) GetValidator() sdk.Address { return b.ValidatorAddr }
func (b Delegation) GetBondShares() sdk.Rat    { return b.Shares }
