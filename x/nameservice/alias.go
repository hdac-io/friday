package nameservice

import (
	"github.com/hdac-io/friday/x/nameservice/types"
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
)

var (
	NewMsgSetAccount = types.NewMsgSetAccount
	NewMsgAddrCheck  = types.NewMsgAddrCheck
	NewMsgChangeKey  = types.NewMsgChangeKey
	ModuleCdc        = types.ModuleCdc
	RegisterCodec    = types.RegisterCodec
	NewUnitAccount   = types.NewUnitAccount
	NewName          = types.NewName
)

type (
	MsgSetAccount       = types.MsgSetAccount
	MsgAddrCheck        = types.MsgAddrCheck
	MsgChangeKey        = types.MsgChangeKey
	QueryResUnitAccount = types.QueryResUnitAccount
	UnitAccount         = types.UnitAccount
)
