package rest

import (
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	grpc "github.com/hdac-io/casperlabs-ee-grpc-go-util/util"

	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/rest"
	"github.com/hdac-io/friday/x/auth/client/utils"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/transfer", storeName), transferHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/bond", storeName), bondHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/unbond", storeName), unbondHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/balance", storeName), getBalanceHandler(cliCtx, storeName)).Methods("GET")
}

func transferHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := transferMsgCreator(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

type transferReq struct {
	BaseReq          rest.BaseReq `json:"base_req"`
	SenderAddress    string       `json:"sender_address"`
	PaymentAmt       uint64       `json:"payment_amount"`
	Fee              uint64       `json:"fee"`
	GasPrice         uint64       `json:"gas_price"`
	RecipientAddress string       `json:"recipient_address"`
}

func transferMsgCreator(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req transferReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	baseReq := req.BaseReq.Sanitize()
	if !baseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// Parameter touching
	senderaddr, err := sdk.AccAddressFromBech32(req.SenderAddress)
	if err != nil {
		return rest.BaseReq{}, nil, fmt.Errorf("Wrong address type")
	} else if !senderaddr.Equals(cliCtx.FromAddress) {
		return rest.BaseReq{}, nil, fmt.Errorf("Sender address should be equal to tx creator")
	}

	receipaddr, err := sdk.AccAddressFromBech32(req.RecipientAddress)
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	// TODO: Change after WASM store feature merge
	transferCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/transfer_to_account.wasm"))
	transferAbi := grpc.MakeArgsTransferToAccount(receipaddr, req.PaymentAmt)
	paymentCode := grpc.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
	paymentAbi := grpc.MakeArgsStandardPayment(new(big.Int).SetUint64(req.Fee))

	// create the message
	msg := types.NewMsgExecute([]byte{0}, senderaddr, cliCtx.FromAddress, transferCode, transferAbi, paymentCode, paymentAbi, req.GasPrice)
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return baseReq, []sdk.Msg{msg}, nil
}

type bondReq struct {
	BaseReq   rest.BaseReq `json:"base_req"`
	Address   string       `json:"address"`
	BondedAmt uint64       `json:"bonded_amount"`
	GasPrice  uint64       `json:"gas_price"`
}

func bondHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := bondUnbondMsgCreator(true, w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func unbondHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := bondUnbondMsgCreator(false, w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func bondUnbondMsgCreator(bondIsTrue bool, w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request) (rest.BaseReq, []sdk.Msg, error) {
	var req bondReq

	// Get body parameters
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse request")
	}

	baseReq := req.BaseReq.Sanitize()
	if !baseReq.ValidateBasic(w) {
		return rest.BaseReq{}, nil, fmt.Errorf("failed to parse base request")
	}

	// Parameter touching
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return rest.BaseReq{}, nil, err
	} else if !addr.Equals(cliCtx.FromAddress) {
		return rest.BaseReq{}, nil, fmt.Errorf("Sender address should be equal to tx creator")
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
	msg := types.NewMsgExecute([]byte{0}, addr, cliCtx.FromAddress, bondingCode, bondingAbi, paymentCode, paymentAbi, req.GasPrice)
	err = msg.ValidateBasic()
	if err != nil {
		return rest.BaseReq{}, nil, err
	}

	return baseReq, []sdk.Msg{msg}, nil
}

func getBalanceHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := getBalanceQuerying(w, cliCtx, r, storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getBalanceQuerying(w http.ResponseWriter, cliCtx context.CLIContext, r *http.Request, storeName string) ([]byte, error) {
	vars := mux.Vars(r)
	straddr, addrexist := vars["address"]
	if addrexist != true {
		rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprint("Need 'address' query string (?address=blahblah)"))
	}

	addr, err := sdk.AccAddressFromBech32(straddr)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
	}

	pubkey := types.ToPublicKey(addr)
	queryData := types.QueryGetBalance{
		Address: pubkey,
	}
	bz := cliCtx.Codec.MustMarshalJSON(queryData)

	res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalance", storeName), bz)

	if err != nil {
		return nil, err
	}
	return res, err
}
