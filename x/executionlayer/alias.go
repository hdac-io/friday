package executionlayer

import (
	"github.com/hdac-io/friday/x/executionlayer/types"
)

const (
	ModuleName      = types.ModuleName
	RouterKey       = types.RouterKey
	HashMapStoreKey = types.HashMapStoreKey
)

var (
	// function aliases
	NewMsgExecute  = types.NewMsgExecute
	NewMsgTransfer = types.NewMsgTransfer
	NewMsgBond     = types.NewMsgBond
	NewMsgUnBond   = types.NewMsgUnBond
	RegisterCodec  = types.RegisterCodec
	NewUnitHashMap = types.NewUnitHashMap

	// variable aliases
	ModuleCdc               = types.ModuleCdc
	ValidatorKey            = types.ValidatorKey
	ValidatorsByConsAddrKey = types.ValidatorsByConsAddrKey

	// error
	ErrValidatorOwnerExists            = types.ErrValidatorOwnerExists
	ErrValidatorPubKeyExists           = types.ErrValidatorPubKeyExists
	ErrValidatorPubKeyTypeNotSupported = types.ErrValidatorPubKeyTypeNotSupported
)

type (
	MsgExecute                = types.MsgExecute
	MsgBond                   = types.MsgBond
	MsgUnBond                 = types.MsgUnBond
	MsgCreateValidator        = types.MsgCreateValidator
	MsgEditValidator          = types.MsgEditValidator
	UnitHashMap               = types.UnitHashMap
	QueryExecutionLayerDetail = types.QueryExecutionLayerDetail
	QueryGetBalanceDetail     = types.QueryGetBalanceDetail
	QueryGetStakeDetail       = types.QueryGetStakeDetail
	QueryGetVoteDetail        = types.QueryGetVoteDetail
	QueryValidatorParams      = types.QueryValidatorParams
	QueryDelegatorParams      = types.QueryDelegatorParams
	QueryVoterParams          = types.QueryVoterParams
	QueryVoterParamsUref      = types.QueryVoterParamsUref
	QueryVoterParamsHash      = types.QueryVoterParamsHash
)
