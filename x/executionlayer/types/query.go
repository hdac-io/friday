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
	KeyType string `json:"key_type"`
	KeyData string `json:"key_data"`
	Path    string `json:"path"`
}

// implement fmt.Stringer
func (q QueryExecutionLayerDetail) String() string {
	return fmt.Sprintf("Key type: %s\nKey data: %s\nPath: %s", q.KeyType, q.KeyData, q.Path)
}

// QueryGetBalanceDetail payload for balance query
type QueryGetBalanceDetail struct {
	Address sdk.AccAddress `json:"address"`
}

// implement fmt.Stringer
func (q QueryGetBalanceDetail) String() string {
	return fmt.Sprintf("Query public key or readable name: %s", q.Address)
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

// defines the params for the following queries:
// - 'custom/%s/voter'
type QueryVoterParams interface {
	GetAddress() sdk.AccAddress
	GetContract() ContractAddress
}

var _ QueryVoterParams = QueryVoterParamsUref{}
var _ QueryVoterParams = QueryVoterParamsHash{}

type QueryVoterParamsUref struct {
	Address  sdk.AccAddress      `json:"address"`
	Contract ContractUrefAddress `json:"contract_address"`
}

func NewQueryVoterUrefParams(address sdk.AccAddress, contractAddress ContractUrefAddress) QueryVoterParamsUref {
	return QueryVoterParamsUref{
		Address:  address,
		Contract: contractAddress,
	}
}

func (q QueryVoterParamsUref) GetAddress() sdk.AccAddress {
	return q.Address
}

func (q QueryVoterParamsUref) GetContract() ContractAddress {
	return q.Contract
}

type QueryVoterParamsHash struct {
	Address  sdk.AccAddress      `json:"address"`
	Contract ContractHashAddress `json:"contract_address"`
}

func NewQueryVoterHashParams(address sdk.AccAddress, contractAddress ContractHashAddress) QueryVoterParamsHash {
	return QueryVoterParamsHash{
		Address:  address,
		Contract: contractAddress,
	}
}

func (q QueryVoterParamsHash) GetAddress() sdk.AccAddress {
	return q.Address
}

func (q QueryVoterParamsHash) GetContract() ContractAddress {
	return q.Contract
}

// QueryGetReward payload for reward query
type QueryGetReward struct {
	Address sdk.AccAddress `json:"address"`
}

func NewQueryGetReward(address sdk.AccAddress) QueryGetReward {
	return QueryGetReward{
		Address: address,
	}
}

// implement fmt.Stringer
func (q QueryGetReward) String() string {
	return fmt.Sprintf("Query public key or readable name: %s", q.Address)
}

// QueryGetCommission payload for commission query
type QueryGetCommission struct {
	Address sdk.AccAddress `json:"address"`
}

func NewQueryGetCommission(address sdk.AccAddress) QueryGetCommission {
	return QueryGetCommission{
		Address: address,
	}
}

// implement fmt.Stringer
func (q QueryGetCommission) String() string {
	return fmt.Sprintf("Query public key or readable name: %s", q.Address)
}
