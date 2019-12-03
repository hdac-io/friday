package executionlayer

import (
	"bytes"
	"fmt"
	"reflect"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
)

// NewHandler returns a handler for "executionlayer" type messages.
func NewHandler(k ExecutionLayerKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgExecute:
			return handlerMsgExecute(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized bank message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgExecute
func handlerMsgExecute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute) sdk.Result {
	if bytes.Equal(msg.BlockState, []byte{0}) {
		msg.BlockState = ctx.BlockHeader().LastBlockId.Hash
	}
	unitHash := k.GetUnitHashMap(ctx, msg.BlockState)

	// Execute
	deploys := util.MakeInitDeploys()
	deploy := util.MakeDeploy(util.EncodeToHexString(types.ToPublicKey(msg.ContractOwnerAccount)), msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice, ctx.BlockTime().Unix(), ctx.ChainID())
	deploys = util.AddDeploy(deploys, deploy)
	effects, errGrpc := grpc.Execute(k.client, unitHash.EEState, ctx.BlockTime().Unix(), deploys, k.protocolVersion)
	if errGrpc != "" {
		return getResult(false, msg)
	}

	// Commit
	postStateHash, _, errGrpc := grpc.Commit(k.client, unitHash.EEState, effects, k.protocolVersion)
	if errGrpc != "" {
		return getResult(false, msg)
	}

	k.SetEEState(ctx, msg.BlockState, postStateHash)

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
