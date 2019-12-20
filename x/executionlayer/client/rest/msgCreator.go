package rest

import (
	"fmt"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strconv"

	grpc "github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/rest"
	bank "github.com/hdac-io/friday/x/bank"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

type transferReq struct {
	ChainID          string `json:"chain_id"`
	SenderAddress    string `json:"sender_address"`
	Amount           string `json:"amount"`
	Fee              uint64 `json:"fee"`
	GasPrice         uint64 `json:"gas_price"`
	RecipientAddress string `json:"recipient_address"`
	Memo             string `json:"memo"`
}

func transferMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req transferReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	baseReq := rest.BaseReq{
		From:    req.SenderAddress,
		ChainID: req.ChainID,
		Gas:     fmt.Sprint(req.GasPrice),
		Memo:    req.Memo,
	}

	if !baseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// Parameter touching
	senderaddr, err := sdk.AccAddressFromBech32(req.SenderAddress)
	if err != nil {
		return rest.BaseReq{}, nil, fmt.Errorf("Wrong address type")
	}

	receipaddr, err := sdk.AccAddressFromBech32(req.RecipientAddress)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	// parse conis trying to be sent
	coins, err := sdk.ParseCoins(req.Amount)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	re := regexp.MustCompile("[0-9]+")
	amount, err := strconv.ParseUint(re.FindAllString(req.Amount, -1)[0], 10, 64)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	// TODO: Change after WASM store feature merge
	transferCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/transfer_to_account.wasm"))
	transferAbi := grpc.MakeArgsTransferToAccount(receipaddr, amount)
	paymentCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
	paymentAbi := grpc.MakeArgsStandardPayment(new(big.Int).SetUint64(req.GasPrice))

	// create the message
	bankMsg := bank.NewMsgSend(senderaddr, receipaddr, coins)
	eeMsg := types.NewMsgExecute([]byte{0}, senderaddr, senderaddr, transferCode, transferAbi, paymentCode, paymentAbi, req.GasPrice)
	err = eeMsg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return baseReq, []sdk.Msg{bankMsg, eeMsg}, nil
}

type bondReq struct {
	ChainID   string `json:"chain_id"`
	Address   string `json:"address"`
	BondedAmt uint64 `json:"bonded_amount"`
	GasPrice  uint64 `json:"gas_price"`
	Memo      string `json:"memo"`
}

func bondUnbondMsgCreator(bondIsTrue bool, w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req bondReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	baseReq := rest.BaseReq{
		From:    req.Address,
		ChainID: req.ChainID,
		Gas:     fmt.Sprint(req.GasPrice),
		Memo:    req.Memo,
	}

	if !baseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// Parameter touching
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	// TODO: Change after WASM store feature merge
	var bondingCode []byte
	if bondIsTrue == true {
		bondingCode = grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/bonding.wasm"))
	} else {
		bondingCode = grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/unbonding.wasm"))
	}
	bondingAbi := grpc.MakeArgsTransferToAccount(addr, req.BondedAmt)
	paymentCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
	paymentAbi := grpc.MakeArgsStandardPayment(new(big.Int).SetUint64(req.GasPrice))

	// create the message
	msg := types.NewMsgExecute([]byte{0}, addr, addr, bondingCode, bondingAbi, paymentCode, paymentAbi, req.GasPrice)
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return baseReq, []sdk.Msg{msg}, nil
}

func getBalanceQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request, storeName string) ([]byte, error) {
	vars := r.URL.Query()
	straddr := vars.Get("address")

	addr, err := sdk.AccAddressFromBech32(straddr)
	if err != nil {
		return nil, err
	}

	pubkey := types.ToPublicKey(addr)
	queryData := types.QueryGetBalance{
		Address: pubkey,
	}
	bz := cliCtx.Codec.MustMarshalJSON(queryData)
	return bz, nil
}
