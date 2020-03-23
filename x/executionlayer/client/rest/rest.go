package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/types/rest"
	"github.com/hdac-io/friday/x/auth/client/utils"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	const (
		hdacSpecific = "hdac"
		general      = "contract"
	)

	r.HandleFunc(fmt.Sprintf("/%s/transfer", hdacSpecific), transferHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/bond", hdacSpecific), bondHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/unbond", hdacSpecific), unbondHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/delegate", hdacSpecific), delegateHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/undelegate", hdacSpecific), undelegateHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/redelegate", hdacSpecific), redelegateHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/balance", hdacSpecific), getBalanceHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/validators", hdacSpecific), getValidatorHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/validators", hdacSpecific), createValidatorHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/validators", hdacSpecific), editValidatorHandler(cliCtx)).Methods("PUT")
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

func delegateHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := delegateUndelegateMsgCreator(true, w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func undelegateHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := delegateUndelegateMsgCreator(false, w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func redelegateHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := redelegateMsgCreator(w, cliCtx, r)
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

func createValidatorHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := createValidatorMsgCreator(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func editValidatorHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := editValidatorMsgCreator(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func getValidatorHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err := getValidatorQuerying(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var res []byte
		if len(bz) == 0 {
			res, _, err = cliCtx.Query(fmt.Sprintf("custom/%s/queryallvalidator", storeName))
		} else {
			res, _, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/queryvalidator", types.ModuleName), bz)
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}
