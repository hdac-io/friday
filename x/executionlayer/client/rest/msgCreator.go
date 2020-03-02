package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/rest"
	cliutil "github.com/hdac-io/friday/x/executionlayer/client/util"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

type transferReq struct {
	BaseReq                    rest.BaseReq `json:"base_req"`
	TokenContractAddress       string       `json:"token_contract_address"`
	RecipientAddressOrNickname string       `json:"recipient_address_or_nickname"`
	Amount                     string       `json:"amount"`
	Fee                        string       `json:"fee"`
}

func transferMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req transferReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	var senderAddr sdk.AccAddress
	senderAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
	if err != nil {
		senderAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.BaseReq.From)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse sender address or name: %s", req.BaseReq.From)
		}
	}

	req.BaseReq.From = senderAddr.String()
	if !req.BaseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// Parameter touching
	var recipientAddr sdk.AccAddress
	recipientAddr, err = sdk.AccAddressFromBech32(req.RecipientAddressOrNickname)
	if err != nil {
		recipientAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.RecipientAddressOrNickname)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse recipient address or name: %s", req.RecipientAddressOrNickname)
		}
	}

	amount, err := cliutil.UnitConverterRemovePoint(req.Amount)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	fee, err := cliutil.UnitConverterRemovePoint(req.Fee)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	gas, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	// create the message
	eeMsg := types.NewMsgTransfer(req.TokenContractAddress, senderAddr, recipientAddr, amount, fee, gas)
	err = eeMsg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{eeMsg}, nil
}

type bondReq struct {
	BaseReq              rest.BaseReq `json:"base_req"`
	TokenContractAddress string       `json:"token_contract_address"`
	Amount               string       `json:"amount"`
	Fee                  string       `json:"fee"`
}

func bondUnbondMsgCreator(bondIsTrue bool, w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req bondReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	// Parameter touching
	var addr sdk.AccAddress
	addr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
	if err != nil {
		addr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.BaseReq.From)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse address or name: %s", req.BaseReq.From)
		}
	}

	req.BaseReq.From = addr.String()
	if !req.BaseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	amount, err := cliutil.UnitConverterRemovePoint(req.Amount)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	fee, err := cliutil.UnitConverterRemovePoint(req.Fee)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	var msg sdk.Msg
	gas, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	if bondIsTrue == true {
		msg = types.NewMsgBond(req.TokenContractAddress, addr, amount, fee, gas)
	} else {
		msg = types.NewMsgUnBond(req.TokenContractAddress, addr, amount, fee, gas)
	}

	// create the message
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

func getBalanceQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request, storeName string) ([]byte, error) {
	vars := r.URL.Query()
	straddr := vars.Get("address")
	blockHashStr := vars.Get("block")

	addr, err := cliutil.GetAddress(cliCtx.Codec, cliCtx, straddr)
	if err != nil {
		return nil, err
	}

	var bz []byte

	if blockHashStr == "" {
		queryData := types.QueryGetBalance{
			Address: addr,
		}
		bz = cliCtx.Codec.MustMarshalJSON(queryData)
		//return bz, nil
	} else {
		blockHash, err := hex.DecodeString(blockHashStr)
		if err != nil {
			return nil, err
		}
		queryData := types.QueryGetBalanceDetail{
			Address:   addr,
			StateHash: blockHash,
		}
		bz = cliCtx.Codec.MustMarshalJSON(queryData)
	}

	return bz, nil
}

type createValidatorReq struct {
	BaseReq     rest.BaseReq      `json:"base_req"`
	ConsPubKey  string            `json:"cons_pub_key"`
	Description types.Description `json:"description"`
}

func createValidatorMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req createValidatorReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	var valAddr sdk.AccAddress
	valAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
	if err != nil {
		valAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.BaseReq.From)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse sender address or name: %s", req.BaseReq.From)
		}
	}

	req.BaseReq.From = valAddr.String()
	if !req.BaseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// Parameter touching
	consPubKey, err := sdk.GetConsPubKeyBech32(req.ConsPubKey)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	// create the message
	msg := types.NewMsgCreateValidator(valAddr, consPubKey, req.Description)
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

type editValidatorReq struct {
	BaseReq     rest.BaseReq      `json:"base_req"`
	Description types.Description `json:"description"`
}

func editValidatorMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req editValidatorReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	var valAddr sdk.AccAddress
	valAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
	if err != nil {
		valAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.BaseReq.From)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse sender address or name: %s", req.BaseReq.From)
		}
	}

	req.BaseReq.From = valAddr.String()
	if !req.BaseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// create the message
	msg := types.NewMsgEditValidator(valAddr, req.Description)
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

func getValidatorQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) ([]byte, error) {
	vars := r.URL.Query()
	strAddr := vars.Get("address")

	if strAddr == "" {
		return []byte{}, nil
	}

	addr, err := cliutil.GetAddress(cliCtx.Codec, cliCtx, strAddr)
	if err != nil {
		return nil, err
	}

	queryData := types.QueryValidatorParams{
		ValidatorAddr: addr,
	}

	bz := cliCtx.Codec.MustMarshalJSON(queryData)

	return bz, nil
}
