package executionlayer

import (
	"fmt"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc/transforms"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// NewHandler returns a handler for "executionlayer" type messages.
func NewHandler(k ExecutionLayerKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg, simulate bool) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgExecute:
			return handlerMsgExecute(ctx, k, msg, simulate)
		case types.MsgTransfer:
			return handlerMsgTransfer(ctx, k, msg, simulate)
		case types.MsgCreateValidator:
			return handlerMsgCreateValidator(ctx, k, msg, simulate)
		case types.MsgEditValidator:
			return handlerMsgEditValidator(ctx, k, msg, simulate)
		case types.MsgBond:
			return handlerMsgBond(ctx, k, msg, simulate)
		case types.MsgUnBond:
			return handlerMsgUnBond(ctx, k, msg, simulate)
		case types.MsgDelegate:
			return handlerMsgDelegate(ctx, k, msg, simulate)
		case types.MsgUndelegate:
			return handlerMsgUndelgate(ctx, k, msg, simulate)
		case types.MsgRedelegate:
			return handlerMsgRedelegate(ctx, k, msg, simulate)
		case types.MsgVote:
			return handlerMsgVote(ctx, k, msg, simulate)
		case types.MsgUnvote:
			return handlerMsgUnvote(ctx, k, msg, simulate)
		case types.MsgClaim:
			return handlerMsgClaim(ctx, k, msg, simulate)
		default:
			errMsg := fmt.Sprintf("unrecognized execution layer messgae type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgExecute
// Transfer function executes "Execute" of Execution layer, that is specialized for transfer
// Difference of general execution
//   1) Raw account is needed for checking address existence
//   2) Fixed transfer & payment WASMs are needed
func handlerMsgTransfer(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgTransfer, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.TransferMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_ListType{ListType: &state.CLType_List{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U8}}}}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_BytesValue{
						BytesValue: msg.ToAddress}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: msg.Amount}}}}}}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)
	if result == true {
		k.SetAccountIfNotExists(ctx, msg.ToAddress)
	}
	return getResult(result, log)
}

// Handle MsgExecute
func handlerMsgExecute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute, simulate bool) sdk.Result {
	result, log := execute(ctx, k, msg, simulate)
	return getResult(result, log)
}

func handlerMsgCreateValidator(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgCreateValidator, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		validator = types.Validator{}
	}

	validator.OperatorAddress = msg.ValidatorAddress
	validator.ConsPubKey = msg.ConsPubKey
	validator.Stake = ""
	description, err := validator.Description.UpdateDescription(msg.Description)
	validator.Description = description

	if proxyContractHash != nil {
		paymentAmount := "0"
		if err == nil {
			paymentAmount = types.BASIC_PAY_AMOUNT
		}

		sessionArgsStr, parseError := getPayAmountSessionArgsStr(paymentAmount)

		msgExecute := NewMsgExecute(
			msg.ContractAddress,
			msg.ValidatorAddress,
			util.HASH,
			proxyContractHash,
			sessionArgsStr,
			msg.Fee,
		)

		result, log := execute(ctx, k, msgExecute, simulate)

		if err != nil {
			return getResult(false, err.Error())
		} else if parseError != nil {
			return getResult(false, parseError.Error())
		} else if log != "" || !result {
			return getResult(false, log)
		}
	}

	k.SetValidator(ctx, msg.ValidatorAddress, validator)
	return getResult(true, "")
}

func handlerMsgEditValidator(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgEditValidator, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)

	paymentAmount := "0"
	if found && err == nil {
		paymentAmount = types.BASIC_PAY_AMOUNT
	}

	sessionArgsStr, parseError := getPayAmountSessionArgsStr(paymentAmount)

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.ValidatorAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)

	result, log := execute(ctx, k, msgExecute, simulate)

	if !found {
		return getResult(false, "validator does not exist for that address")
	} else if err != nil {
		return getResult(false, err.Error())
	} else if parseError != nil {
		return getResult(false, parseError.Error())
	} else if log != "" || !result {
		return getResult(false, log)
	}

	validator.Description = description
	k.SetValidator(ctx, msg.ValidatorAddress, validator)
	return getResult(true, "")
}

func handlerMsgBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgBond, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.BondMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: msg.Amount}}}}}}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)
	return getResult(result, log)
}

func handlerMsgUnBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgUnBond, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Name: "method",
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.UnbondMethodName}}}},
		&consensus.Deploy_Arg{
			Name: "amount",
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_OptionType{OptionType: &state.CLType_Option{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}}}}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_OptionValue{
						OptionValue: &state.CLValueInstance_Option{
							Value: &state.CLValueInstance_Value{
								Value: &state.CLValueInstance_Value_U512{
									U512: &state.CLValueInstance_U512{
										Value: string(msg.Amount)}}}}}}}}}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)

	return getResult(result, log)
}

func handlerMsgDelegate(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgDelegate, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.DelegateMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_ListType{ListType: &state.CLType_List{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U8}}}}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_BytesValue{
						BytesValue: msg.ValAddress}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: msg.Amount}}}}}}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)

	return getResult(result, log)
}

func handlerMsgUndelgate(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgUndelegate, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.UndelegateMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_ListType{ListType: &state.CLType_List{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U8}}}}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_BytesValue{
						BytesValue: msg.ValAddress}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_OptionType{OptionType: &state.CLType_Option{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}}}}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_OptionValue{
						OptionValue: &state.CLValueInstance_Option{
							Value: &state.CLValueInstance_Value{
								Value: &state.CLValueInstance_Value_U512{
									U512: &state.CLValueInstance_U512{
										Value: msg.Amount}}}}}}}}}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)

	return getResult(result, log)
}

func handlerMsgRedelegate(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgRedelegate, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.RedelegateMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_ListType{ListType: &state.CLType_List{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U8}}}}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_BytesValue{
						BytesValue: msg.SrcValAddress}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_ListType{ListType: &state.CLType_List{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U8}}}}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_BytesValue{
						BytesValue: msg.DestValAddress}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: msg.Amount}}}}}}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)

	return getResult(result, log)
}

func handlerMsgVote(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgVote, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	var sessionArgs []*consensus.Deploy_Arg

	if strings.HasPrefix(msg.TargetContractAddress, sdk.Bech32PrefixContractURef) {
		contractAddr, err := sdk.ContractUrefAddressFromBech32(msg.TargetContractAddress)
		if err != nil {
			getResult(false, err.Error())
		}

		sessionArgs = []*consensus.Deploy_Arg{
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_StrValue{
							StrValue: types.VoteMethodName}}}},
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_KEY}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_Key{
							Key: &state.Key{Value: &state.Key_Hash_{Hash: &state.Key_Hash{Hash: contractAddr.Bytes()}}}}}}},
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_U512{
							U512: &state.CLValueInstance_U512{
								Value: msg.Amount}}}}}}

	} else if strings.HasPrefix(msg.TargetContractAddress, sdk.Bech32PrefixContractHash) {
		contractAddr, err := sdk.ContractHashAddressFromBech32(msg.TargetContractAddress)
		if err != nil {
			getResult(false, err.Error())
		}

		sessionArgs = []*consensus.Deploy_Arg{
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_StrValue{
							StrValue: types.VoteMethodName}}}},
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_KEY}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_Key{
							Key: &state.Key{Value: &state.Key_Hash_{Hash: &state.Key_Hash{Hash: contractAddr.Bytes()}}}}}}},
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_U512{
							U512: &state.CLValueInstance_U512{
								Value: msg.Amount}}}}}}

	}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)

	return getResult(result, log)
}

func handlerMsgUnvote(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgUnvote, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	var sessionArgs []*consensus.Deploy_Arg

	if strings.HasPrefix(msg.TargetContractAddress, sdk.Bech32PrefixContractURef) {
		contractAddr, err := sdk.ContractUrefAddressFromBech32(msg.TargetContractAddress)
		if err != nil {
			getResult(false, err.Error())
		}

		sessionArgs = []*consensus.Deploy_Arg{
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_StrValue{
							StrValue: types.UnvoteMethodName}}}},
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_KEY}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_Key{
							Key: &state.Key{Value: &state.Key_Hash_{Hash: &state.Key_Hash{Hash: contractAddr.Bytes()}}}}}}},
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_OptionType{OptionType: &state.CLType_Option{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}}}}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_OptionValue{
							OptionValue: &state.CLValueInstance_Option{
								Value: &state.CLValueInstance_Value{
									Value: &state.CLValueInstance_Value_U512{
										U512: &state.CLValueInstance_U512{
											Value: msg.Amount}}}}}}}}}

	} else if strings.HasPrefix(msg.TargetContractAddress, sdk.Bech32PrefixContractHash) {
		contractAddr, err := sdk.ContractHashAddressFromBech32(msg.TargetContractAddress)
		if err != nil {
			getResult(false, err.Error())
		}

		sessionArgs = []*consensus.Deploy_Arg{
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_StrValue{
							StrValue: types.UnvoteMethodName}}}},
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_KEY}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_Key{
							Key: &state.Key{Value: &state.Key_Hash_{Hash: &state.Key_Hash{Hash: contractAddr.Bytes()}}}}}}},
			&consensus.Deploy_Arg{
				Value: &state.CLValueInstance{
					ClType: &state.CLType{Variants: &state.CLType_OptionType{OptionType: &state.CLType_Option{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}}}}},
					Value: &state.CLValueInstance_Value{
						Value: &state.CLValueInstance_Value_OptionValue{
							OptionValue: &state.CLValueInstance_Option{
								Value: &state.CLValueInstance_Value{
									Value: &state.CLValueInstance_Value_U512{
										U512: &state.CLValueInstance_U512{
											Value: msg.Amount}}}}}}}}}

	}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)

	return getResult(result, log)
}

func handlerMsgClaim(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgClaim, simulate bool) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	var methodName string
	switch msg.RewardOrCommission {
	case types.CommissionValue:
		methodName = types.ClaimCommissionMethodName
	case types.RewardValue:
		methodName = types.ClaimRewardMethodName
	default:
		return getResult(false, "Must be reward or commission")
	}

	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: methodName}}}}}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		return getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		msg.ContractAddress,
		msg.FromAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		msg.Fee,
	)
	result, log := execute(ctx, k, msgExecute, simulate)

	return getResult(result, log)
}

func execute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute, simulate bool) (bool, string) {
	proxyContractHash := k.GetProxyContractHash(ctx)
	// Parameter preparation
	var stateHash []byte
	if simulate {
		stateHash = k.GetUnitHashMap(ctx, ctx.BlockHeight()).EEState
	} else {
		stateHash = ctx.CandidateBlock().State
	}
	protocolVersion := k.MustGetProtocolVersion(ctx)
	log := ""

	paymentArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.PaymentMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: string(msg.Fee)}}}}}}

	paymentArgsJson, err := DeployArgsToJsonString(paymentArgs)
	if err != nil {
		return false, err.Error()
	}

	executeAddress := []byte{}
	if len(msg.ExecAddress) == sdk.AddrLen {
		executeAddress = msg.ExecAddress
	} else {
		executeAddress = msg.ExecAddress
	}

	// Execute
	deploys := []*ipc.DeployItem{}
	deploy, err := util.MakeDeploy(
		executeAddress,
		msg.SessionType, msg.SessionCode, msg.SessionArgs,
		util.HASH, proxyContractHash, paymentArgsJson,
		types.BASIC_GAS, ctx.BlockTime().Unix(), ctx.ChainID())
	if err != nil {
		return false, err.Error()
	}
	deploys = append(deploys, deploy)
	reqExecute := &ipc.ExecuteRequest{
		ParentStateHash: stateHash,
		BlockTime:       uint64(ctx.BlockTime().Unix()),
		Deploys:         deploys,
		ProtocolVersion: &protocolVersion,
	}
	resExecute, err := k.client.Execute(ctx.Context(), reqExecute)
	if err != nil {
		return false, err.Error()
	}

	effects := []*transforms.TransformEntry{}
	switch resExecute.GetResult().(type) {
	case *ipc.ExecuteResponse_Success:
		for _, res := range resExecute.GetSuccess().GetDeployResults() {
			switch res.GetExecutionResult().GetError().GetValue().(type) {
			case *ipc.DeployError_GasError:
				err = types.ErrGRpcExecuteDeployGasError(types.DefaultCodespace)
			case *ipc.DeployError_ExecError:
				err = types.ErrGRpcExecuteDeployExecError(types.DefaultCodespace, res.GetExecutionResult().GetError().GetExecError().GetMessage())
			}

			effects = append(effects, res.GetExecutionResult().GetEffects().GetTransformMap()...)
			if err != nil {
				log += fmt.Sprintf(log, err.Error())
			}
		}
	case *ipc.ExecuteResponse_MissingParent:
		err = types.ErrGRpcExecuteMissingParent(types.DefaultCodespace, util.EncodeToHexString(resExecute.GetMissingParent().GetHash()))
		log += err.Error()
	default:
		err = fmt.Errorf("Unknown result : %s", resExecute.String())
		log += err.Error()
	}

	if simulate {
		return log == "", log
	}

	// Commit
	postStateHash, bonds, errGrpc := grpc.Commit(k.client, stateHash, effects, &protocolVersion)
	log += errGrpc

	candidateBlock := ctx.CandidateBlock()
	candidateBlock.State = postStateHash
	candidateBlock.Bonds = bonds

	result := false
	if log == "" {
		result = true
	}

	return result, log
}

func getResult(ok bool, log string) sdk.Result {
	res := sdk.Result{}
	if ok {
		res.Code = sdk.CodeOK
	} else {
		res.Code = sdk.CodeUnknownRequest
	}
	res.Log = log

	return res
}

func getPayAmountSessionArgsStr(amount string) (string, error) {
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.PaymentMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: amount}}}}}}
	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)

	return sessionArgsStr, err
}
