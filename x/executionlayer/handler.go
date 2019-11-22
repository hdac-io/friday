package executionlayer

import (
	"fmt"

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
	stateHash := k.GetNextState(ctx, msg.BlockState)

	// Execute
	deploys := util.MakeInitDeploys()
	deploy := util.MakeDeploy(util.EncodeToHexString(msg.ContractOwnerAccount), msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice, ctx.BlockTime().Unix(), ctx.ChainID())
	deploys = util.AddDeploy(deploys, deploy)
	effects, errGrpc := grpc.Execute(k.client, stateHash, ctx.BlockTime().Unix(), deploys, k.protocolVersion)
	if errGrpc != "" {
		return sdk.Result{}
	}

	// Commit
	postStateHash, validators, errGrpc := grpc.Commit(k.client, stateHash, effects, k.protocolVersion)
	if errGrpc != "" {
		return sdk.Result{}
	}

	k.SetNextState(ctx, msg.BlockState, postStateHash)

	_ = validators

	return sdk.Result{}
}
