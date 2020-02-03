package rest

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"os"

	grpc "github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
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

	// TODO: Change after WASM store feature merge
	transferCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/transfer_to_account.wasm"))
	transferAbi := grpc.MakeArgsTransferToAccount(recipientAddr.ToEEAddress(), req.Amount)
	paymentCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
	paymentAbi := grpc.MakeArgsStandardPayment(new(big.Int).SetUint64(req.Fee))

	// create the message
	eeMsg := types.NewMsgTransfer(req.TokenContractAddress, senderAddr, recipientAddr, transferCode, transferAbi, paymentCode, paymentAbi, req.GasPrice)
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

	// TODO: Change after WASM store feature merge
	var bondingUnbondingCode []byte
	var bondingUnbondingAbi []byte
	if bondIsTrue == true {
		bondingUnbondingCode = grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/bonding.wasm"))
		bondingUnbondingAbi = grpc.MakeArgsBonding(req.Amount)
	} else {
		bondingUnbondingCode = grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/unbonding.wasm"))
		bondingUnbondingAbi = grpc.MakeArgsUnBonding(req.Amount)
	}
	paymentCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
	paymentAbi := grpc.MakeArgsStandardPayment(new(big.Int).SetUint64(req.Fee))

	// create the message
	msg := types.NewMsgExecute(req.TokenContractAddress, addr, bondingUnbondingCode, bondingUnbondingAbi, paymentCode, paymentAbi, req.GasPrice)
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
