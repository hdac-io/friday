package executionlayer

import (
	"github.com/hdac-io/friday/x/executionlayer/types"
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
)

var (
	NewMsgExecute  = types.NewMsgExecute
	ModuleCdc      = types.ModuleCdc
	RegisterCodec  = types.RegisterCodec
	NewUnitHashMap = types.NewUnitHashMap
)

type (
	MsgExecute          = types.MsgExecute
	QueryExecutionLayer = types.QueryExecutionLayer
	UnitHashMap         = types.UnitHashMap
)
