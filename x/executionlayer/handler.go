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
	err := k.Transfer(ctx, msg.TokenContractAddress, msg.FromAddress, msg.ToAddress, msg.TransferCode, msg.TransferArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}
	return getResult(true, msg)
}

// Handle MsgExecute
func handlerMsgExecute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute) sdk.Result {
	err := k.Execute(ctx, msg.ExecAddress, msg.ContractAddress,
		msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}
	return getResult(true, msg)
}

func handlerMsgCreateValidator(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgCreateValidator) sdk.Result {
	eeAddress := sdk.AccAddress(msg.ValidatorPubKey.Address()).ToEEAddress()
	validator, found := k.GetValidator(ctx, eeAddress)
	if !found {
		validator = types.Validator{}
	}

	validator.OperatorAddress = eeAddress
	validator.ConsPubKey = msg.ConsPubKey
	validator.Description = msg.Description
	validator.Stake = ""

	k.SetValidator(ctx, eeAddress, validator)

	return getResult(true, msg)
}

func handlerMsgBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgBond) sdk.Result {
	err := k.Execute(ctx, msg.BonderAddress, msg.TokenContractAddress, msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}
	return getResult(true, msg)
}

func handlerMsgUnBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgUnBond) sdk.Result {
	err := k.Execute(ctx, msg.UnbonderAddress, msg.TokenContractAddress, msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}

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
