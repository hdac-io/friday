package types

import (
	"github.com/hdac-io/friday/codec"
)

// ModuleCdc is used as a codec in types package
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgExecute{}, "executionengine/Execute", nil)
}
