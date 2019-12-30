package readablename

import (
	"github.com/hdac-io/friday/x/readablename/types"
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
)

var (
	NewMsgSetAccount = types.NewMsgSetAccount
	NewMsgChangeKey  = types.NewMsgChangeKey
	ModuleCdc        = types.ModuleCdc
	RegisterCodec    = types.RegisterCodec
	NewUnitAccount   = types.NewUnitAccount
	NewName          = types.NewName
)

type (
	MsgSetAccount       = types.MsgSetAccount
	MsgChangeKey        = types.MsgChangeKey
	QueryResUnitAccount = types.QueryResUnitAccount
	UnitAccount         = types.UnitAccount
	QueryReqUnitAccount = types.QueryReqUnitAccount
)
