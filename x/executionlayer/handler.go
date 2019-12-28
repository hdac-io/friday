package executionlayer

import (
	"fmt"
	"reflect"

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
func handlerMsgTransfer(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgTransfer) sdk.Result {
	err := k.Transfer(ctx, msg.TokenOwnerAccount, msg.FromAccount, msg.ToAccount, msg.TransferCode, msg.TransferArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}
	return getResult(true, msg)
}

// Handle MsgExecute
func handlerMsgExecute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute) sdk.Result {
	err := k.Execute(ctx, msg.BlockHash, msg.ExecAccount, msg.ContractOwnerAccount,
		msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}
	return getResult(true, msg)
}

func handlerMsgCreateValidator(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgCreateValidator) sdk.Result {
	validator, found := k.GetValidator(ctx, msg.DelegatorAddress)
	if !found {
		validator = types.Validator{}
	}

	validator.OperatorAddress = msg.ValidatorAddress
	validator.ConsPubKey = msg.PubKey
	validator.Description = msg.Description

	k.SetValidator(ctx, msg.DelegatorAddress, validator)

	return getResult(true, msg)
}

func handlerMsgBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgBond) sdk.Result {
	_, found := k.GetValidator(ctx, msg.ExecAccount)
	if !found {
		return getResult(false, msg)
	}

	blockHash := k.GetCandidateBlockHash(ctx)
	unitHash := k.GetUnitHashMap(ctx, blockHash)

	// Execute
	deploys := util.MakeInitDeploys()
	deploy := util.MakeDeploy(types.ToPublicKey(msg.ContractOwnerAccount), msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice, ctx.BlockTime().Unix(), ctx.ChainID())
	deploys = util.AddDeploy(deploys, deploy)

	protocolVersion := k.MustGetProtocolVersion(ctx)
	effects, errGrpc := grpc.Execute(k.client, unitHash.EEState, ctx.BlockTime().Unix(), deploys, &protocolVersion)
	if errGrpc != "" {
		return getResult(false, msg)
	}

	// Commit
	postStateHash, bonds, errGrpc := grpc.Commit(k.client, unitHash.EEState, effects, &protocolVersion)
	if errGrpc != "" {
		return getResult(false, msg)
	}

	k.SetEEState(ctx, blockHash, postStateHash)
	k.SetCandidateBlockBond(ctx, bonds)

	return getResult(true, msg)
}

func handlerMsgUnBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgUnBond) sdk.Result {
	_, found := k.GetValidator(ctx, msg.ExecAccount)
	if !found {
		return getResult(false, msg)
	}

	blockHash := k.GetCandidateBlockHash(ctx)
	unitHash := k.GetUnitHashMap(ctx, blockHash)

	// Execute
	deploys := util.MakeInitDeploys()
	deploy := util.MakeDeploy(types.ToPublicKey(msg.ContractOwnerAccount), msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice, ctx.BlockTime().Unix(), ctx.ChainID())
	deploys = util.AddDeploy(deploys, deploy)

	protocolVersion := k.MustGetProtocolVersion(ctx)
	effects, errGrpc := grpc.Execute(k.client, unitHash.EEState, ctx.BlockTime().Unix(), deploys, &protocolVersion)
	if errGrpc != "" {
		return getResult(false, msg)
	}

	// Commit
	postStateHash, bonds, errGrpc := grpc.Commit(k.client, unitHash.EEState, effects, &protocolVersion)
	if errGrpc != "" {
		return getResult(false, msg)
	}

	k.SetEEState(ctx, blockHash, postStateHash)
	k.SetCandidateBlockBond(ctx, bonds)

	return getResult(true, msg)
}

func getResult(ok bool, msg sdk.Msg) sdk.Result {
	res := sdk.Result{}
	if ok {
		res.Code = sdk.CodeOK
	} else {
		res.Code = sdk.CodeUnknownRequest
	}

	events := sdk.EmptyEvents()
	event := sdk.Event{}
	v := reflect.ValueOf(msg)
	typeOfV := v.Type()
	for i := 0; i < v.NumField(); i++ {
		event.AppendAttributes(
			sdk.NewAttribute(typeOfV.Field(i).Name, fmt.Sprintf("%v", v.Field(i).Interface())),
		)
	}
	events.AppendEvent(event)
	res.Events = events

	return res
}
