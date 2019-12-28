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
	validator := types.Validator{
		OperatorAddress: msg.ValidatorAddress,
		ConsPubKey:      msg.PubKey,
		Description:     msg.Description,
	}

	k.SetValidator(ctx, msg.DelegatorAddress, validator)

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
