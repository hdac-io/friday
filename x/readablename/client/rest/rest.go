package rest

import (
	"fmt"
	"net/http"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/types/rest"
	"github.com/hdac-io/friday/x/readablename/types"

	"github.com/gorilla/mux"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth/client/utils"
)

const (
	restName = "readablename"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/newname", restName), newNameHandler(cliCtx)).Methods("POST")         // New account
	r.HandleFunc(fmt.Sprintf("/%s/names", restName), changeKeyHandler(cliCtx)).Methods("PUT")          // Change Key
	r.HandleFunc(fmt.Sprintf("/%s/names", restName), getNameHandler(cliCtx, storeName)).Methods("GET") // Get UnitAccount
}

// --------------------------------------------------------------------------------------
// Tx Handler

type newNameReq struct {
	ChainID  string `json:"chain_id"`
	GasPrice uint64 `json:"gas_price"`
	Memo     string `json:"memo"`
	Name     string `json:"name"`
	PubKey   string `json:"pubkey"`
}

func newNameHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req newNameReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		// Try to parse public key
		pubkey, err := sdk.GetAccPubKeyBech32(req.PubKey)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "fail to parse public key")
			return
		}
		addr := sdk.AccAddress(pubkey.Address())

		// Force organizing base request
		baseReq := rest.BaseReq{
			From:    addr.String(),
			ChainID: req.ChainID,
			Gas:     fmt.Sprint(req.GasPrice),
			Memo:    req.Memo,
		}

		if !baseReq.ValidateBasic(w) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse base request")
			return
		}

		baseReq = baseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse base request")
			return
		}

		// create the message
		msg := types.NewMsgSetAccount(types.NewName(req.Name), addr, pubkey)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type changeKeyReq struct {
	ChainID   string `json:"chain_id"`
	GasPrice  uint64 `json:"gas_price"`
	Memo      string `json:"memo"`
	Name      string `json:"name"`
	OldPubKey string `json:"old_pubkey"`
	NewPubKey string `json:"new_pubkey"`
}

func changeKeyHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req changeKeyReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		// Try to parse public key
		oldpubkey, err := sdk.GetAccPubKeyBech32(req.OldPubKey)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "fail to parse old public key")
			return
		}
		oldaddr := sdk.AccAddress(oldpubkey.Address())

		newpubkey, err := sdk.GetAccPubKeyBech32(req.NewPubKey)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "fail to parse new public key")
			return
		}
		newaddr := sdk.AccAddress(newpubkey.Address())

		// Force organizing base request
		baseReq := rest.BaseReq{
			From:    oldaddr.String(),
			ChainID: req.ChainID,
			Gas:     fmt.Sprint(req.GasPrice),
			Memo:    req.Memo,
		}

		baseReq = baseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgChangeKey(req.Name, oldaddr, newaddr, oldpubkey, newpubkey)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

//--------------------------------------------------------------------------------------
// Query Handlers

func getNameHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		straddr := vars.Get("address")

		param := types.QueryReqUnitAccount{
			Name: straddr,
		}
		bz, err := types.ModuleCdc.MarshalJSON(param)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaccount", storeName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
