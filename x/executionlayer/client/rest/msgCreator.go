package rest

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/rest"
	cliutil "github.com/hdac-io/friday/x/executionlayer/client/util"
	"github.com/hdac-io/friday/x/executionlayer/types"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
)

type contractRunReq struct {
	BaseReq                       rest.BaseReq `json:"base_req"`
	ExecutionType                 string       `json:"type"`
	TokenContractAddressOrKeyName string       `json:"token_contract_address_or_key_name"`
	Base64EncodedBinary           string       `json:"base64_encoded_binary"`
	Args                          string       `json:"args"`
	Fee                           string       `json:"fee"`
}

func contractRunMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req contractRunReq

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

	sessionType := cliutil.GetContractType(req.ExecutionType)
	var sessionCode []byte
	var contractAddress string

	switch sessionType {
	case util.WASM:
		contractAddress = "wasm_file_direct_execution"
		sessionCode, err = base64.StdEncoding.DecodeString(req.Base64EncodedBinary)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to decode WASM binary")
		}
	case util.HASH:
		contractAddress = req.TokenContractAddressOrKeyName
		contractHashAddr, err := sdk.ContractHashAddressFromBech32(req.TokenContractAddressOrKeyName)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to decode given contract hash address")
		}
		sessionCode = contractHashAddr.Bytes()
	case util.UREF:
		contractAddress = req.TokenContractAddressOrKeyName
		contractUrefAddr, err := sdk.ContractUrefAddressFromBech32(req.TokenContractAddressOrKeyName)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to decode given contract uref address")
		}
		sessionCode = contractUrefAddr.Bytes()
	case util.NAME:
		contractAddress = fmt.Sprintf("%s:%s", senderAddr.String(), req.TokenContractAddressOrKeyName)
		sessionCode = []byte(req.TokenContractAddressOrKeyName)
	default:
		return rest.BaseReq{}, nil, fmt.Errorf("type must be one of wasm, name, uref, or hash")
	}

	fee, err := cliutil.ToBigsun(cliutil.Hdac(req.Fee))
	if err != nil {
		return rest.BaseReq{}, nil, fmt.Errorf("error on conversion from bigsun to token")
	}

	gasPrice, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, fmt.Errorf("error to parse into gas")
	}

	// build and sign the transaction, then broadcast to Tendermint
	msg := types.NewMsgExecute(
		contractAddress,
		senderAddr,
		sessionType,
		sessionCode,
		req.Args,
		string(fee),
		gasPrice,
	)

	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

func getContractQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) ([]byte, string, error) {
	vars := r.URL.Query()
	dataType := vars.Get("data_type")
	data := vars.Get("data")
	path := vars.Get("path")
	blockhash := vars.Get("blockhash")

	queryData := types.QueryExecutionLayerDetail{
		KeyType:   dataType,
		KeyData:   data,
		Path:      path,
		BlockHash: blockhash,
	}
	bz := cliCtx.Codec.MustMarshalJSON(queryData)

	return bz, path, nil
}

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

	amount, err := cliutil.ToBigsun(cliutil.Hdac(req.Amount))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	fee, err := cliutil.ToBigsun(cliutil.Hdac(req.Fee))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	gas, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	// create the message
	eeMsg := types.NewMsgTransfer(req.TokenContractAddress, senderAddr, recipientAddr, string(amount), string(fee), gas)
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

	amount, err := cliutil.ToBigsun(cliutil.Hdac(req.Amount))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	fee, err := cliutil.ToBigsun(cliutil.Hdac(req.Fee))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	var msg sdk.Msg
	gas, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	if bondIsTrue == true {
		msg = types.NewMsgBond(req.TokenContractAddress, addr, string(amount), string(fee), gas)
	} else {
		msg = types.NewMsgUnBond(req.TokenContractAddress, addr, string(amount), string(fee), gas)
	}

	// create the message
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

type delegateReq struct {
	BaseReq              rest.BaseReq `json:"base_req"`
	TokenContractAddress string       `json:"token_contract_address"`
	ValidatorAddress     string       `json:"validator_address"`
	Amount               string       `json:"amount"`
	Fee                  string       `json:"fee"`
}

func delegateUndelegateMsgCreator(delegateIsTrue bool, w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req delegateReq

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

	valAddress, err := sdk.AccAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		valAddress, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.ValidatorAddress)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse address or name: %s", req.BaseReq.From)
		}
	}

	amount, err := cliutil.ToBigsun(cliutil.Hdac(req.Amount))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	fee, err := cliutil.ToBigsun(cliutil.Hdac(req.Fee))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	var msg sdk.Msg
	gas, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	if delegateIsTrue == true {
		msg = types.NewMsgDelegate(req.TokenContractAddress, addr, valAddress, string(amount), string(fee), gas)
	} else {
		msg = types.NewMsgUndelegate(req.TokenContractAddress, addr, valAddress, string(amount), string(fee), gas)
	}

	// create the message
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

type redelegateReq struct {
	BaseReq              rest.BaseReq `json:"base_req"`
	TokenContractAddress string       `json:"token_contract_address"`
	SrcValidatorAddress  string       `json:"src_validator_address"`
	DestValidatorAddress string       `json:"dest_validator_address"`
	Amount               string       `json:"amount"`
	Fee                  string       `json:"fee"`
}

func redelegateMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req redelegateReq

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

	srcValAddress, err := sdk.AccAddressFromBech32(req.SrcValidatorAddress)
	if err != nil {
		srcValAddress, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.SrcValidatorAddress)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse address or name: %s", req.BaseReq.From)
		}
	}

	destValAddress, err := sdk.AccAddressFromBech32(req.DestValidatorAddress)
	if err != nil {
		destValAddress, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, req.DestValidatorAddress)
		if err != nil {
			return rest.BaseReq{}, nil, fmt.Errorf("failed to parse address or name: %s", req.BaseReq.From)
		}
	}

	amount, err := cliutil.ToBigsun(cliutil.Hdac(req.Amount))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	fee, err := cliutil.ToBigsun(cliutil.Hdac(req.Fee))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	var msg sdk.Msg
	gas, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	msg = types.NewMsgRedelegate(req.TokenContractAddress, addr, srcValAddress, destValAddress, string(amount), string(fee), gas)

	// create the message
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

type voteReq struct {
	BaseReq              rest.BaseReq `json:"base_req"`
	TokenContractAddress string       `json:"token_contract_address"`
	Hash                 string       `json:"hash"`
	Amount               string       `json:"amount"`
	Fee                  string       `json:"fee"`
}

func voteUnvoteMsgCreator(voteIsTrue bool, w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req voteReq

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

	hash, err := hex.DecodeString(req.Hash)
	if err != nil {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse hash: %s", req.Hash)
	}

	amount, err := cliutil.ToBigsun(cliutil.Hdac(req.Amount))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	fee, err := cliutil.ToBigsun(cliutil.Hdac(req.Fee))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	var msg sdk.Msg
	gas, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	if voteIsTrue == true {
		msg = types.NewMsgVote(req.TokenContractAddress, addr, hash, string(amount), string(fee), gas)
	} else {
		msg = types.NewMsgUnvote(req.TokenContractAddress, addr, hash, string(amount), string(fee), gas)
	}

	// create the message
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

type claimReq struct {
	BaseReq              rest.BaseReq `json:"base_req"`
	TokenContractAddress string       `json:"token_contract_address"`
	RewardOrCommission   bool         `json:"reward_or_commission"`
	Amount               string       `json:"amount"`
	Fee                  string       `json:"fee"`
}

func claimMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req claimReq

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

	fee, err := cliutil.ToBigsun(cliutil.Hdac(req.Fee))
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	var msg sdk.Msg
	gas, err := strconv.ParseUint(req.BaseReq.Gas, 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	msg = types.NewMsgClaim(req.TokenContractAddress, addr, req.RewardOrCommission, string(fee), gas)

	// create the message
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return req.BaseReq, []sdk.Msg{msg}, nil
}

func getRewardQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request, storeName string) ([]byte, error) {
	vars := r.URL.Query()
	straddr := vars.Get("address")

	addr, err := cliutil.GetAddress(cliCtx.Codec, cliCtx, straddr)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	queryData := types.NewQueryGetReward(addr)
	bz := cliCtx.Codec.MustMarshalJSON(queryData)

	return bz, nil
}

func getCommissionQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request, storeName string) ([]byte, error) {
	vars := r.URL.Query()
	straddr := vars.Get("address")

	addr, err := cliutil.GetAddress(cliCtx.Codec, cliCtx, straddr)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	queryData := types.NewQueryGetCommission(addr)
	bz := cliCtx.Codec.MustMarshalJSON(queryData)

	return bz, nil
}

func getBalanceQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request, storeName string) ([]byte, error) {
	vars := r.URL.Query()
	straddr := vars.Get("address")
	blockHashStr := vars.Get("block")

	addr, err := cliutil.GetAddress(cliCtx.Codec, cliCtx, straddr)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	queryData := types.QueryGetBalanceDetail{
		Address:   addr,
		BlockHash: blockHashStr,
	}
	bz := cliCtx.Codec.MustMarshalJSON(queryData)

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

func getDelegatorQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) ([]byte, error) {
	vars := r.URL.Query()
	validatorAddressStr := vars.Get("validator")
	validatorAddress, err := cliutil.GetAddress(cliCtx.Codec, cliCtx, validatorAddressStr)
	if err != nil {
		return nil, err
	}

	delegatorAddressStr := vars.Get("delegator")
	delegatorAddress, err := cliutil.GetAddress(cliCtx.Codec, cliCtx, delegatorAddressStr)
	if err != nil {
		return nil, err
	}

	if validatorAddress.Empty() && delegatorAddress.Empty() {
		return nil, fmt.Errorf("Requires validator or delegate address.")
	}

	queryData := types.QueryDelegatorParams{
		ValidatorAddr: validatorAddress,
		DelegatorAddr: delegatorAddress,
	}

	bz := cliCtx.Codec.MustMarshalJSON(queryData)

	return bz, nil
}

func getVoterQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) ([]byte, error) {
	vars := r.URL.Query()
	hashStr := vars.Get("hash")
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		return nil, err
	}

	addressStr := vars.Get("address")
	address, err := cliutil.GetAddress(cliCtx.Codec, cliCtx, addressStr)
	if err != nil {
		return nil, err
	}

	if address.Empty() && len(hash) == 0 {
		return nil, fmt.Errorf("Requires hash or voter address.")
	}

	queryData := types.QueryVoterParams{
		Hash:    hash,
		Address: address,
	}

	bz := cliCtx.Codec.MustMarshalJSON(queryData)

	return bz, nil
}
