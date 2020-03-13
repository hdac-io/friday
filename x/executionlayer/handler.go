package executionlayer

import (
	"fmt"

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
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgExecute:
			return handlerMsgExecute(ctx, k, msg)
		case types.MsgTransfer:
			return handlerMsgTransfer(ctx, k, msg)
		case types.MsgCreateValidator:
			return handlerMsgCreateValidator(ctx, k, msg)
		case types.MsgEditValidator:
			return handlerMsgEditValidator(ctx, k, msg)
		case types.MsgBond:
			return handlerMsgBond(ctx, k, msg)
		case types.MsgUnBond:
			return handlerMsgUnBond(ctx, k, msg)
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
func handlerMsgTransfer(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgTransfer) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_StringValue{
					StringValue: types.TransferMethodName}}},
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_BytesValue{
					BytesValue: msg.ToAddress.ToEEAddress()}}},
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_BigInt{
					BigInt: &state.BigInt{Value: msg.Amount, BitWidth: 512}}}},
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
		msg.GasPrice,
	)
	result, log := execute(ctx, k, msgExecute)
	if result == true {
		k.SetAccountIfNotExists(ctx, msg.ToAddress)
	}
	return getResult(result, log)
}

// Handle MsgExecute
func handlerMsgExecute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute) sdk.Result {
	result, log := execute(ctx, k, msg)
	return getResult(result, log)
}

func handlerMsgCreateValidator(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgCreateValidator) sdk.Result {
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		validator = types.Validator{}
	}

	validator.OperatorAddress = msg.ValidatorAddress
	validator.ConsPubKey = msg.ConsPubKey
	validator.Description = msg.Description
	validator.Stake = ""

	k.SetValidator(ctx, msg.ValidatorAddress, validator)

	return getResult(true, "")
}

func handlerMsgEditValidator(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgEditValidator) sdk.Result {
	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return getResult(false, "validator does not exist for that address")
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return getResult(false, err.Error())
	}

	validator.Description = description

	k.SetValidator(ctx, msg.ValidatorAddress, validator)
	return getResult(true, "")
}

func handlerMsgBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgBond) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_StringValue{
					StringValue: types.BondMethodName}}},
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_BigInt{
					BigInt: &state.BigInt{Value: msg.Amount, BitWidth: 512}}}}}

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
		msg.GasPrice,
	)
	result, log := execute(ctx, k, msgExecute)
	return getResult(result, log)
}

func handlerMsgUnBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgUnBond) sdk.Result {
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_StringValue{
					StringValue: types.UnbondMethodName}}},
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_OptionalValue{
					OptionalValue: &consensus.Deploy_Arg_Value{
						Value: &consensus.Deploy_Arg_Value_BigInt{
							BigInt: &state.BigInt{Value: string(msg.Amount), BitWidth: 512}}}}}}}

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
		msg.GasPrice,
	)
	result, log := execute(ctx, k, msgExecute)

	return getResult(result, log)
}

func execute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute) (bool, string) {
	proxyContractHash := k.GetProxyContractHash(ctx)
	// Parameter preparation
	stateHash := ctx.CandidateBlock().State
	protocolVersion := k.MustGetProtocolVersion(ctx)
	log := ""

	paymentArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_StringValue{
					StringValue: types.PaymentMethodName}}},
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_BigInt{
					BigInt: &state.BigInt{Value: string(msg.Fee), BitWidth: 512}}}}}

	paymentArgsJson, err := DeployArgsToJsonString(paymentArgs)
	if err != nil {
		return false, err.Error()
	}

	// Execute
	deploys := []*ipc.DeployItem{}
	deploy, err := util.MakeDeploy(
		msg.ExecAddress.ToEEAddress(),
		msg.SessionType, msg.SessionCode, msg.SessionArgs,
		util.HASH, proxyContractHash, paymentArgsJson,
		msg.GasPrice, ctx.BlockTime().Unix(), ctx.ChainID())
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
				log = fmt.Sprintf(log, err.Error())
			}
		}
	case *ipc.ExecuteResponse_MissingParent:
		err = types.ErrGRpcExecuteMissingParent(types.DefaultCodespace, util.EncodeToHexString(resExecute.GetMissingParent().GetHash()))
		return false, err.Error()
	default:
		err = fmt.Errorf("Unknown result : %s", resExecute.String())
		return false, err.Error()
	}

	// Commit
	postStateHash, bonds, errGrpc := grpc.Commit(k.client, stateHash, effects, &protocolVersion)
	if errGrpc != "" {
		return false, errGrpc
	}

	candidateBlock := ctx.CandidateBlock()
	candidateBlock.State = postStateHash
	candidateBlock.Bonds = bonds

	return true, log
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
