package types

import (
	"github.com/hdac-io/friday/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateValidator{}, "friday/MsgCreateValidator", nil)
	cdc.RegisterConcrete(MsgEditValidator{}, "friday/MsgEditValidator", nil)
	cdc.RegisterConcrete(MsgDelegate{}, "friday/MsgDelegate", nil)
	cdc.RegisterConcrete(MsgUndelegate{}, "friday/MsgUndelegate", nil)
	cdc.RegisterConcrete(MsgBeginRedelegate{}, "friday/MsgBeginRedelegate", nil)
}

// generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
