package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/rest"
	cliutil "github.com/hdac-io/friday/x/executionlayer/client/util"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

type transferReq struct {
	ChainID                    string `json:"chain_id"`
	TokenContractAddress       string `json:"token_contract_address"`
	SenderAddressOrNickname    string `json:"sender_address_or_nickname"`
	RecipientAddressOrNickname string `json:"recipient_address_or_nickname"`
	Amount                     uint64 `json:"amount"`
	Fee                        uint64 `json:"fee"`
	GasPrice                   uint64 `json:"gas_price"`
	Memo                       string `json:"memo"`
}

func transferMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req transferReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	var senderAddr sdk.AccAddress
	senderAddr, err := sdk.AccAddressFromBech32(req.SenderAddressOrNickname)
	if err != nil {
		senderAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.SenderAddressOrNickname)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse sender address or name: %s", req.SenderAddressOrNickname)
		}
	}

	baseReq := rest.BaseReq{
		From:    senderAddr.String(),
		ChainID: req.ChainID,
		Gas:     fmt.Sprint(req.GasPrice),
		Memo:    req.Memo,
	}

	if !baseReq.ValidateBasic(w) {
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

	// create the message
	eeMsg := types.NewMsgTransfer(req.TokenContractAddress, senderAddr, recipientAddr, req.Amount, req.Fee, req.GasPrice)
	err = eeMsg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return baseReq, []sdk.Msg{eeMsg}, nil
}

type bondReq struct {
	ChainID              string `json:"chain_id"`
	TokenContractAddress string `json:"token_contract_address"`
	AddressOrNickname    string `json:"address_or_nickname"`
	Amount               uint64 `json:"amount"`
	GasPrice             uint64 `json:"gas_price"`
	Fee                  uint64 `json:"fee"`
	Memo                 string `json:"memo"`
}

func bondUnbondMsgCreator(bondIsTrue bool, w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req bondReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	// Parameter touching
	var addr sdk.AccAddress
	addr, err := sdk.AccAddressFromBech32(req.AddressOrNickname)
	if err != nil {
		addr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.AddressOrNickname)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse address or name: %s", req.AddressOrNickname)
		}
	}

	baseReq := rest.BaseReq{
		From:    addr.String(),
		ChainID: req.ChainID,
		Gas:     fmt.Sprint(req.GasPrice),
		Memo:    req.Memo,
	}

	if !baseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	var msg sdk.Msg
	if bondIsTrue == true {
		msg = types.NewMsgBond(req.TokenContractAddress, addr, req.Amount, req.Fee, req.GasPrice)
	} else {
		msg = types.NewMsgUnBond(req.TokenContractAddress, addr, req.Amount, req.Fee, req.GasPrice)
	}

	// create the message
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return baseReq, []sdk.Msg{msg}, nil
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
	ChainID                    string            `json:"chain_id"`
	ValidatorAddressOrNickName string            `json:"validator_address_or_nickname"`
	ConsPubKey                 string            `json:"cons_pub_key"`
	Description                types.Description `json:"description"`
	Gas                        uint64            `json:"gas"`
	Memo                       string            `json:"memo"`
}

func createValidatorMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req createValidatorReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	var valAddr sdk.AccAddress
	valAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddressOrNickName)
	if err != nil {
		valAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.ValidatorAddressOrNickName)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse sender address or name: %s", req.ValidatorAddressOrNickName)
		}
	}

	baseReq := rest.BaseReq{
		From:    valAddr.String(),
		ChainID: req.ChainID,
		Gas:     fmt.Sprint(req.Gas),
		Memo:    req.Memo,
	}

	if !baseReq.ValidateBasic(w) {
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

	return baseReq, []sdk.Msg{msg}, nil
}

type editValidatorReq struct {
	ChainID                    string            `json:"chain_id"`
	ValidatorAddressOrNickName string            `json:"validator_address_or_nickname"`
	Description                types.Description `json:"description"`
	Gas                        uint64            `json:"gas"`
	Memo                       string            `json:"memo"`
}

func editValidatorMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req editValidatorReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	var valAddr sdk.AccAddress
	valAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddressOrNickName)
	if err != nil {
		valAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.ValidatorAddressOrNickName)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse sender address or name: %s", req.ValidatorAddressOrNickName)
		}
	}

	baseReq := rest.BaseReq{
		From:    valAddr.String(),
		ChainID: req.ChainID,
		Gas:     fmt.Sprint(req.Gas),
		Memo:    req.Memo,
	}

	if !baseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// create the message
	msg := types.NewMsgEditValidator(valAddr, req.Description)
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return baseReq, []sdk.Msg{msg}, nil
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
