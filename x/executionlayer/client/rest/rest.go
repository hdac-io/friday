package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/types/rest"
	"github.com/hdac-io/friday/x/auth/client/utils"
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

func getBalanceHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err := getBalanceQuerying(w, cliCtx, r, storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalance", storeName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}
