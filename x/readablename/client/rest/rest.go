package rest

import (
	"fmt"
	"net/http"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/types/rest"
	"github.com/hdac-io/friday/x/readablename/types"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth/client/utils"
	"github.com/gorilla/mux"
)

const (
	restName = "name"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/names", storeName), addressCheckHandler(cliCtx)).Methods("POST")      // Address check
	r.HandleFunc(fmt.Sprintf("/%s/newname", storeName), newNameHandler(cliCtx)).Methods("POST")         // New account
	r.HandleFunc(fmt.Sprintf("/%s/names", storeName), changeKeyHandler(cliCtx)).Methods("PUT")          // Change Key
	r.HandleFunc(fmt.Sprintf("/%s/names", storeName), getNameHandler(cliCtx, storeName)).Methods("GET") // Get UnitAccount
	//r.HandleFunc(fmt.Sprintf("/%s/names/{%s}", storeName, restName), resolveNameHandler(cliCtx, storeName)).Methods("GET")
	//r.HandleFunc(fmt.Sprintf("/%s/names/{%s}/new", storeName, restName), whoIsHandler(cliCtx, storeName)).Methods("GET")
}

// --------------------------------------------------------------------------------------
// Tx Handler

type addressCheckReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	ID      string       `json:"id"`
	Address string       `json:"address"`
}

func addressCheckHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addressCheckReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(req.Address)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgAddrCheck(req.ID, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type newNameReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	ID      string       `json:"id"`
	Address string       `json:"address"`
}

func newNameHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req newNameReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(req.Address)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgSetAccount(req.ID, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type changeKeyReq struct {
	BaseReq    rest.BaseReq `json:"base_req"`
	ID         string       `json:"id"`
	OldAddress string       `json:"oldaddress"`
	NewAddress string       `json:"newaddress"`
}

func changeKeyHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req changeKeyReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		oldaddr, err := sdk.AccAddressFromBech32(req.OldAddress)
		newaddr, err := sdk.AccAddressFromBech32(req.NewAddress)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgChangeKey(req.ID, oldaddr, newaddr)
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
		vars := mux.Vars(r)
		paramType := vars[restName]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaccount/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
