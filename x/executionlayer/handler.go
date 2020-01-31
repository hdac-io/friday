package executionlayer

import (
	"fmt"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
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
//   2) Fixed transfer & payemtn WASMs are needed
func handlerMsgTransfer(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgTransfer) sdk.Result {
	k.SetAccountIfNotExists(ctx, msg.ToPubkey)
	result, log := execute(ctx, k, msg.MsgExecute)

	return getResult(result, log)
}

// Handle MsgExecute
func handlerMsgExecute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute) sdk.Result {
	result, log := execute(ctx, k, msg)
	return getResult(result, log)
}

func handlerMsgCreateValidator(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgCreateValidator) sdk.Result {
	eeAddress, err := sdk.GetEEAddressFromCryptoPubkey(msg.ValidatorPubKey)
	if err != nil {
		return getResult(false, err.Error())
	}

	validator, found := k.GetValidator(ctx, eeAddress)
	if !found {
		validator = types.Validator{}
	}

	validator.OperatorAddress = eeAddress
	validator.ConsPubKey = msg.ConsPubKey
	validator.Description = msg.Description
	validator.Stake = ""

	k.SetValidator(ctx, eeAddress, validator)

	return getResult(true, "")
}

func handlerMsgBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgBond) sdk.Result {
	result, log := execute(ctx, k, msg.MsgExecute)
	return getResult(result, log)
}

func handlerMsgUnBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgUnBond) sdk.Result {
	result, log := execute(ctx, k, msg.MsgExecute)

	return getResult(result, log)
}

func execute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute) (bool, string) {

	// Parameter preparation
	stateHash := ctx.CandidateBlock().State
	protocolVersion := k.MustGetProtocolVersion(ctx)
	execAddr := sdk.GetEEAddressFromSecp256k1PubKey(msg.ExecPubkey)
	log := ""

	// Execute
	deploys := []*ipc.DeployItem{
		util.MakeDeploy(execAddr.Bytes(), msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice, ctx.BlockTime().Unix(), ctx.ChainID())}
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

	candidateBlock := ctx.CandidateBlock()
	switch resExecute.GetResult().(type) {
	case *ipc.ExecuteResponse_Success:
		for _, res := range resExecute.GetSuccess().GetDeployResults() {
			switch res.GetExecutionResult().GetError().GetValue().(type) {
			case *ipc.DeployError_GasError:
				err = types.ErrGRpcExecuteDeployGasError(types.DefaultCodespace)
			case *ipc.DeployError_ExecError:
				err = types.ErrGRpcExecuteDeployExecError(types.DefaultCodespace, res.GetExecutionResult().GetError().GetExecError().GetMessage())
			}

			candidateBlock.Effects = append(candidateBlock.Effects, res.GetExecutionResult().GetEffects().GetTransformMap()...)
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
