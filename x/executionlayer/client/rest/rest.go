package rest

import (
	"fmt"
	"net/http"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/mux"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/storedvalue"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/types/rest"
	"github.com/hdac-io/friday/x/auth/client/utils"
	cliutil "github.com/hdac-io/friday/x/executionlayer/client/util"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	const (
		hdacSpecific = "hdac"
		general      = "contract"
	)

	r.HandleFunc(fmt.Sprintf("/%s", general), contractRunHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s", general), contractQueryHandler(cliCtx, storeName)).Methods("GET")

	r.HandleFunc(fmt.Sprintf("/%s/transfer", hdacSpecific), transferHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/bond", hdacSpecific), bondHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/unbond", hdacSpecific), unbondHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/delegate", hdacSpecific), delegateHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/undelegate", hdacSpecific), undelegateHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/redelegate", hdacSpecific), redelegateHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/delegator", hdacSpecific), getDelegatorHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/vote", hdacSpecific), voteHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/vote", hdacSpecific), getVoteHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/unvote", hdacSpecific), unvoteHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/claim", hdacSpecific), claimHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/reward", hdacSpecific), getRewardHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/commission", hdacSpecific), getCommissionHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/balance", hdacSpecific), getBalanceHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/stake", hdacSpecific), getStakeHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/validators", hdacSpecific), getValidatorHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/validators", hdacSpecific), createValidatorHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/validators", hdacSpecific), editValidatorHandler(cliCtx)).Methods("PUT")
}

func contractRunHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := contractRunMsgCreator(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func contractQueryHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, path, err := getContractQuerying(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querydetail", types.ModuleName), bz)
		var storedValue storedvalue.StoredValue
		storedValue, err, _ = storedValue.FromBytes(res)
		if err != nil {
			fmt.Printf("could not resolve data - %s\nerr : %s\n", path, err.Error())
			return
		}
		marshaler := jsonpb.Marshaler{Indent: "  "}

		valueStr := ""

		switch storedValue.Type {
		case storedvalue.TYPE_ACCOUNT:
			value := &state.Value{Value: &state.Value_Account{Account: storedValue.Account.ToStateValue()}}
			valueStr, err = marshaler.MarshalToString(value)
		case storedvalue.TYPE_CONTRACT:
			value := &state.Value{Value: &state.Value_Contract{Contract: storedValue.Contract.ToStateValue()}}
			valueStr, err = marshaler.MarshalToString(value)
		case storedvalue.TYPE_CL_VALUE:
			value := storedValue.ClValue.ToCLInstanceValue()
			valueStr, err = marshaler.MarshalToString(value)
		}

		if err != nil {
			fmt.Printf("could not resolve data - %s\nerr : %s\n", path, err.Error())
			return
		}

		valueStr = cliutil.ReplaceBase64HashToBech32(path, valueStr)

		rest.PostProcessResponseBare(w, cliCtx, []byte(valueStr))
	}
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

func voteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := voteUnvoteMsgCreator(true, w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func unvoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := voteUnvoteMsgCreator(false, w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func claimHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseReq, msgs, err := claimMsgCreator(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}

func getBalanceHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err, cliCtx := getBalanceQuerying(w, cliCtx, r, storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalancedetail", storeName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}

func getStakeHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err, cliCtx := getStakeQuerying(w, cliCtx, r, storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querystakedetail", storeName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}

func getVoteHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err, cliCtx := getVoteQuerying(w, cliCtx, r, storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/queryvotedetail", storeName), bz)
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
		bz, err, cliCtx := getValidatorQuerying(w, cliCtx, r)
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

func getDelegatorHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err, cliCtx := getDelegatorQuerying(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querydelegator", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}

func getVoterHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err, cliCtx := getVoterQuerying(w, cliCtx, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/queryvoter", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}

func getRewardHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err, cliCtx := getRewardQuerying(w, cliCtx, r, storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/queryreward", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}

func getCommissionHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bz, err, cliCtx := getCommissionQuerying(w, cliCtx, r, storeName)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querycommission", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}
