package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/rest"

	"github.com/hdac-io/friday/x/auth/client/utils"
	"github.com/hdac-io/friday/x/nickname/types"
)

const (
	restName = "nickname"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/new", restName), newNicknameHandler(cliCtx)).Methods("POST")         // New account
	r.HandleFunc(fmt.Sprintf("/%s/change", restName), changeKeyHandler(cliCtx)).Methods("PUT")         // Change Key
	r.HandleFunc(fmt.Sprintf("/%s/names", restName), getNameHandler(cliCtx, storeName)).Methods("GET") // Get UnitAccount
}

// --------------------------------------------------------------------------------------
// Tx Handler

type newNickname struct {
	BaseReq rest.BaseReq `json:"base_req"`

	Nickname string `json:"nickname"`
	Address  string `json:"address"`
}

func newNicknameHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req newNickname
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		addr, err := sdk.AccAddressFromBech32(req.Address)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse from given address")
		}

		req.BaseReq.From = req.Address
		if !req.BaseReq.ValidateBasic(w) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse base request")
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse base request")
			return
		}

		// create the message
		msg := types.NewMsgSetNickname(types.NewName(req.Nickname), addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

type changeKey struct {
	BaseReq rest.BaseReq `json:"base_req"`

	Nickname   string `json:"nickname"`
	OldAddress string `json:"old_address"`
	NewAddress string `json:"new_address"`
}

func changeKeyHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req changeKey
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		oldaddr, err := sdk.AccAddressFromBech32(req.OldAddress)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse 'old_address'")
		}

		newaddr, err := sdk.AccAddressFromBech32(req.NewAddress)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse 'new_address'")
		}

		req.BaseReq.From = req.OldAddress
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgChangeKey(req.Nickname, oldaddr, newaddr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

//--------------------------------------------------------------------------------------
// Query Handlers

func getNameHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		straddr := vars.Get("address")

		param := types.QueryReqUnitAccount{
			Nickname: straddr,
		}
		bz, err := types.ModuleCdc.MarshalJSON(param)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaddress", storeName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
