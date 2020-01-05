package executionlayer

import (
	"fmt"
	"reflect"
	"strconv"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
	abci "github.com/hdac-io/tendermint/abci/types"
	tmtypes "github.com/hdac-io/tendermint/types"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
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
	err := k.Transfer(ctx, msg.TokenContractAddress, msg.FromPubkey, msg.ToPubkey, msg.TransferCode, msg.TransferArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}
	return getResult(true, msg)
}

// Handle MsgExecute
func handlerMsgExecute(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgExecute) sdk.Result {
	err := k.Execute(ctx, msg.BlockHash, msg.ExecPubkey, msg.ContractAddress,
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
	validator.Stake = ""

	k.SetValidator(ctx, msg.DelegatorAddress, validator)

	return getResult(true, msg)
}

func handlerMsgBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgBond) sdk.Result {
	err := k.Execute(ctx, []byte{0}, msg.FromPubkey, msg.TokenContractAddress, msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}
	return getResult(true, msg)
}

func handlerMsgUnBond(ctx sdk.Context, k ExecutionLayerKeeper, msg types.MsgUnBond) sdk.Result {
	err := k.Execute(ctx, []byte{0}, msg.FromPubkey, msg.TokenContractAddress, msg.SessionCode, msg.SessionArgs, msg.PaymentCode, msg.PaymentArgs, msg.GasPrice)
	if err != nil {
		return getResult(false, msg)
	}

	return getResult(true, msg)
}

func EndBloker(ctx sdk.Context, k ExecutionLayerKeeper) []abci.ValidatorUpdate {
	var validatorUpdates []abci.ValidatorUpdate

	validators := k.GetAllValidators(ctx)

	resultbonds := k.GetCandidateBlockBond(ctx)
	resultBondsMap := make(map[string]*ipc.Bond)
	for _, bond := range resultbonds {
		resultBondsMap[string(bond.GetValidatorPublicKey())] = bond
	}

	var power string
	for _, validator := range validators {
		resultBond, found := resultBondsMap[string(types.ToPublicKey(validator.OperatorAddress))]
		if found {
			if validator.Stake == resultBond.GetStake().GetValue() {
				continue
			}
			power = resultBond.GetStake().GetValue()
			validator.Stake = resultBond.GetStake().GetValue()
		} else {
			if validator.Stake != "" {
				power = "0"
				validator.Stake = ""
			}
		}
		// TODO : There is a GasLimit error when the bonding value is greater than 7_000_000.
		coin, err := strconv.ParseInt(power, 10, 64)
		if err != nil {
			continue
		}
		validatorUpdate := abci.ValidatorUpdate{
			PubKey: tmtypes.TM2PB.PubKey(validator.ConsPubKey),
			Power:  coin,
		}
		validatorUpdates = append(validatorUpdates, validatorUpdate)
		k.SetValidator(ctx, validator.OperatorAddress, validator)
	}

	return validatorUpdates
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
