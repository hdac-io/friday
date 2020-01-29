package nickname

import (
	"github.com/hdac-io/friday/x/nickname/types"
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
)

var (
	NewMsgSetAccount = types.NewMsgSetNickname
	NewMsgChangeKey  = types.NewMsgChangeKey
	ModuleCdc        = types.ModuleCdc
	RegisterCodec    = types.RegisterCodec
	NewUnitAccount   = types.NewUnitAccount
	NewName          = types.NewName
)

type (
	MsgSetAccount       = types.MsgSetNickname
	MsgChangeKey        = types.MsgChangeKey
	QueryResUnitAccount = types.QueryResUnitAccount
	UnitAccount         = types.UnitAccount
	QueryReqUnitAccount = types.QueryReqUnitAccount
)
