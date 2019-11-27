package executionlayer

import (
	"github.com/hdac-io/friday/x/executionlayer/types"
)

const (
	ModuleName      = types.ModuleName
	RouterKey       = types.RouterKey
	HashMapStoreKey = types.HashMapStoreKey
	DeployStoreKey  = types.DeployStoreKey
)

var (
	// function aliases
	NewMsgExecute  = types.NewMsgExecute
	RegisterCodec  = types.RegisterCodec
	NewUnitHashMap = types.NewUnitHashMap

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	MsgExecute                = types.MsgExecute
	QueryExecutionLayer       = types.QueryExecutionLayer
	UnitHashMap               = types.UnitHashMap
	QueryExecutionLayerResp   = types.QueryExecutionLayerResp
	QueryExecutionLayerDetail = types.QueryExecutionLayerDetail
	QueryGetBalance           = types.QueryGetBalance
	QueryGetBalanceDetail     = types.QueryGetBalanceDetail
)
