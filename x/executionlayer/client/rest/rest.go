package rest

import (
	"github.com/hdac-io/friday/client/context"

	"github.com/gorilla/mux"
)

const (
	restName = "contract"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	//r.HandleFunc(fmt.Sprintf("/%s/names", storeName), addressCheckHandler(cliCtx)).Methods("POST")      // Address check
}

// The code below is a pattern of defining RESTFul API
// You may start from the example first

// // 1. POST/PUT request

// // Struct means HTTP body
// type addressCheckReq struct {
// 	BaseReq rest.BaseReq `json:"base_req"`
// 	ID      string       `json:"id"`
// 	Address string       `json:"address"`
// }

// func addressCheckHandler(cliCtx context.CLIContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var req addressCheckReq

//		// Get body parameters
// 		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
// 			return
// 		}

// 		baseReq := req.BaseReq.Sanitize()
// 		if !baseReq.ValidateBasic(w) {
// 			return
// 		}

//		// Parameter touching
// 		addr, err := sdk.AccAddressFromBech32(req.Address)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}

// 		// create the message
// 		msg := types.NewMsgAddrCheck(req.ID, addr)
// 		err = msg.ValidateBasic()
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}

// 		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
// 	}
// }

////////////////////////////////////////////////////////////////
//	//	2. Query

// func getNameHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		paramType := vars[restName]

// 		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaccount/%s", storeName, paramType), nil)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
// 			return
// 		}

// 		rest.PostProcessResponse(w, cliCtx, res)
// 	}
// }
