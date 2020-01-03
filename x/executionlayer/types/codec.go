package types

import (
	"github.com/hdac-io/friday/codec"
)

// ModuleCdc is used as a codec in types package
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgExecute{}, "executionengine/Execute", nil)
	cdc.RegisterConcrete(MsgTransfer{}, "executionengine/Transfer", nil)
}
