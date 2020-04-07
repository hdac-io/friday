package types

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
)

const (
	ADDRESS = "address"
	UREF    = "uref"
	HASH    = "hash"
	LOCAL   = "local"

	SYSTEM = "system"
)

// QueryExecutionLayerDetail payload for a EE query
type QueryExecutionLayerDetail struct {
	BlockHash string `json:"state_hash"`
	KeyType   string `json:"key_type"`
	KeyData   string `json:"key_data"`
	Path      string `json:"path"`
}

// implement fmt.Stringer
func (q QueryExecutionLayerDetail) String() string {
	return fmt.Sprintf("Block Hash: %s\nKey type: %s\nKey data: %s\nPath: %s", q.BlockHash, q.KeyType, q.KeyData, q.Path)
}

// QueryGetBalanceDetail payload for balance query
type QueryGetBalanceDetail struct {
	BlockHash string         `json:"state_hash"`
	Address   sdk.AccAddress `json:"address"`
}

// implement fmt.Stringer
func (q QueryGetBalanceDetail) String() string {
	return fmt.Sprintf("State: %s\nQuery public key or readable name: %s", q.BlockHash, q.Address)
}

// QueryExecutionLayerResp is used for response of EE query
type QueryExecutionLayerResp struct {
	Value string `json:"value"`
}

// implement fmt.Stringer
func (q QueryExecutionLayerResp) String() string {
	return fmt.Sprintf("Value: %s", q.Value)
}

// defines the params for the following queries:
// - 'custom/%s/validator'
type QueryValidatorParams struct {
	ValidatorAddr sdk.AccAddress `json:"validator_address"`
}

func NewQueryValidatorParams(validatorAddr sdk.AccAddress) QueryValidatorParams {
	return QueryValidatorParams{
		ValidatorAddr: validatorAddr,
	}
}

// defines the params for the following queries:
// - 'custom/%s/delegator'
type QueryDelegatorParams struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_address"`
	ValidatorAddr sdk.AccAddress `json:"validator_address"`
}

func NewQueryDelegatorParams(delegaatorAddr sdk.AccAddress, validatorAddr sdk.AccAddress) QueryDelegatorParams {
	return QueryDelegatorParams{
		DelegatorAddr: delegaatorAddr,
		ValidatorAddr: validatorAddr,
	}
}
