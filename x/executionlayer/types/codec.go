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
	cdc.RegisterInterface((*ContractAddress)(nil), nil)

	cdc.RegisterConcrete(MsgCreateValidator{}, "executionengine/CreateValidator", nil)
	cdc.RegisterConcrete(MsgEditValidator{}, "executionengine/EditValidator", nil)
	cdc.RegisterConcrete(MsgExecute{}, "executionengine/Execute", nil)
	cdc.RegisterConcrete(MsgTransfer{}, "executionengine/Transfer", nil)
	cdc.RegisterConcrete(MsgBond{}, "executionengine/Bond", nil)
	cdc.RegisterConcrete(MsgUnBond{}, "executionengine/UnBond", nil)
	cdc.RegisterConcrete(MsgDelegate{}, "executionengine/Delegate", nil)
	cdc.RegisterConcrete(MsgUndelegate{}, "executionengine/Undelegate", nil)
	cdc.RegisterConcrete(MsgRedelegate{}, "executionengine/Redelegate", nil)
	cdc.RegisterConcrete(MsgVote{}, "executionengine/Vote", nil)
	cdc.RegisterConcrete(MsgUnvote{}, "executionengine/Unvote", nil)
	cdc.RegisterConcrete(MsgClaim{}, "executionengine/Claim", nil)
	cdc.RegisterConcrete(ContractHashAddress{}, "types/ContractHashAddress", nil)
	cdc.RegisterConcrete(ContractUrefAddress{}, "types/ContractUrefAddress", nil)
}
