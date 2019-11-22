package nameservice

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
)

// NewHandler returns a handler for "nameservice" type messages.
func NewHandler(k AccountKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetAccount:
			return handleMsgSetAccount(ctx, k, msg)
		case MsgChangeKey:
			return handleMsgChangeKey(ctx, k, msg)
		case MsgAddrCheck:
			return handleMsgAddrCheck(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized nameserver Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle a message to set name
func handleMsgSetAccount(ctx sdk.Context, k AccountKeeper, msg MsgSetAccount) sdk.Result {
	k.SetUnitAccount(ctx, msg.ID, msg.Address)
	return sdk.Result{}
}

func handleMsgAddrCheck(ctx sdk.Context, k AccountKeeper, msg MsgAddrCheck) sdk.Result {
	result := k.AddrCheck(ctx, msg.ID, msg.Address)
	res := ""
	if result {
		res = "true"
	} else {
		res = "false"
	}
	return sdk.Result{
		Log: res,
	}
}

// Handle a message to change key
func handleMsgChangeKey(ctx sdk.Context, k AccountKeeper, msg MsgChangeKey) sdk.Result {
	k.ChangeKey(ctx, msg.ID, msg.OldAddress, msg.NewAddress)
	return sdk.Result{}
}
