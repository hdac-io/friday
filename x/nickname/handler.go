package nickname

import (
	"fmt"
	"reflect"

	sdk "github.com/hdac-io/friday/types"
)

// NewHandler returns a handler for "nameservice" type messages.
func NewHandler(k NicknameKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg, simulate bool, txIndex int, msgIndex int) sdk.Result {
		if simulate {
			return sdk.Result{}
		}

		switch msg := msg.(type) {
		case MsgSetAccount:
			return handleMsgSetAccount(ctx, k, msg, simulate)
		case MsgChangeKey:
			return handleMsgChangeKey(ctx, k, msg, simulate)
		default:
			errMsg := fmt.Sprintf("Unrecognized nameserver Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
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

// Handle a message to set name
func handleMsgSetAccount(ctx sdk.Context, k NicknameKeeper, msg MsgSetAccount, simulate bool) sdk.Result {
	res := k.SetNickname(ctx, msg.Nickname.MustToString(), msg.Address)
	processDone(ctx, simulate)

	return getResult(res, msg)
}

// Handle a message to change key
func handleMsgChangeKey(ctx sdk.Context, k NicknameKeeper, msg MsgChangeKey, simulate bool) sdk.Result {
	res := k.ChangeKey(ctx, msg.Nickname, msg.OldAddress, msg.NewAddress)
	processDone(ctx, simulate)

	return getResult(res, msg)
}

func processDone(ctx sdk.Context, simulate bool) {
	if !simulate {
		candidateBlock := ctx.CandidateBlock()
		candidateBlock.WaitGroup.Done()
		ctx = ctx.WithCandidateBlock(candidateBlock)
	}
}
