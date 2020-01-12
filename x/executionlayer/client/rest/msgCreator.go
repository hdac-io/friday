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
	"github.com/hdac-io/friday/x/executionlayer/client/util"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

type transferReq struct {
	ChainID               string `json:"chain_id"`
	TokenContractAddress  string `json:"token_contract_address"`
	SenderPubkeyOrName    string `json:"sender_pubkey_or_name"`
	RecipientPubkeyOrName string `json:"recipient_pubkey_or_name"`
	Amount                uint64 `json:"amount"`
	Fee                   uint64 `json:"fee"`
	GasPrice              uint64 `json:"gas_price"`
	Memo                  string `json:"memo"`
}

func transferMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req transferReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	senderPubkey, err := util.GetPubKey(cliCtx.Codec, cliCtx, req.SenderPubkeyOrName)
	if err != nil {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse sender public key or name: %s", req.SenderPubkeyOrName)
	}
	senderaddr := sdk.AccAddress(senderPubkey.Address().Bytes())

	baseReq := rest.BaseReq{
		From:    senderaddr.String(),
		ChainID: req.ChainID,
		Gas:     fmt.Sprint(req.GasPrice),
		Memo:    req.Memo,
	}

	if !baseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// Parameter touching
	recipientPubkey, err := util.GetPubKey(cliCtx.Codec, cliCtx, req.RecipientPubkeyOrName)
	if err != nil {
		return rest.BaseReq{}, nil, fmt.Errorf("wrong public key or no mapping rule in readable ID service: %s", req.RecipientPubkeyOrName)
	}

	// TODO: Change after WASM store feature merge
	transferCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/transfer_to_account.wasm"))
	transferAbi := grpc.MakeArgsTransferToAccount(recipientPubkey.Bytes(), req.Amount)
	paymentCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
	paymentAbi := grpc.MakeArgsStandardPayment(new(big.Int).SetUint64(req.GasPrice))

	// create the message
	eeMsg := types.NewMsgTransfer(req.TokenContractAddress, *senderPubkey, *recipientPubkey, transferCode, transferAbi, paymentCode, paymentAbi, req.GasPrice, senderaddr)
	err = eeMsg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return baseReq, []sdk.Msg{eeMsg}, nil
}

type bondReq struct {
	ChainID              string `json:"chain_id"`
	TokenContractAddress string `json:"token_contract_address"`
	PubkeyOrName         string `json:"pubkey_or_name"`
	Amount               uint64 `json:"amount"`
	GasPrice             uint64 `json:"gas_price"`
	Memo                 string `json:"memo"`
}

func bondUnbondMsgCreator(bondIsTrue bool, w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req bondReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	pubkey, err := util.GetPubKey(cliCtx.Codec, cliCtx, req.PubkeyOrName)
	if err != nil {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse sender public key or name: %s", req.PubkeyOrName)
	}
	addr := sdk.AccAddress(pubkey.Address().Bytes())

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
	paymentAbi := grpc.MakeArgsStandardPayment(new(big.Int).SetUint64(req.GasPrice))

	// create the message
	msg := types.NewMsgExecute([]byte{0}, req.TokenContractAddress, *pubkey, bondingUnbondingCode, bondingUnbondingAbi, paymentCode, paymentAbi, req.GasPrice, addr)
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

	pubkey, err := util.GetPubKey(cliCtx.Codec, cliCtx, straddr)
	if err != nil {
		return nil, err
	}

	var bz []byte

	if blockHashStr == "" {
		queryData := types.QueryGetBalance{
			PublicKey: *pubkey,
		}
		bz = cliCtx.Codec.MustMarshalJSON(queryData)
		//return bz, nil
	} else {
		blockHash, err := hex.DecodeString(blockHashStr)
		if err != nil {
			return nil, err
		}
		queryData := types.QueryGetBalanceDetail{
			PublicKey: *pubkey,
			StateHash: blockHash,
		}
		bz = cliCtx.Codec.MustMarshalJSON(queryData)
	}

	return bz, nil
}
