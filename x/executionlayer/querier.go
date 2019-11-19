package executionlayer

import (
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/hdac-io/friday/x/executionlayer/types"
)

const (
	QueryEE       = "query"
	QueryEEDetail = "querydetail"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper ExecutionLayerKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryEEDetail:
			return queryEEDetail(ctx, path[1:], req, keeper)
		case QueryEE:
			return queryEE(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown ee query")
		}
	}
}

func queryEEDetail(ctx sdk.Context, path []string, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryExecutionLayerDetail
	err := types.ModuleCdc.UnmarshalJSON(req.Data, param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	val, errmsg := keeper.GetQueryResult(ctx, param.StateHash, param.KeyType, string(param.KeyData), param.Path)
	if errmsg != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, errmsg.Error())
	}

	qryvalue := QueryExecutionLayerResp{
		Value: val.String(),
	}

	res, _ := codec.MarshalJSONIndent(keeper.cdc, qryvalue)
	return res, nil
}

func queryEE(ctx sdk.Context, path []string, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryExecutionLayer
	err := types.ModuleCdc.UnmarshalJSON(req.Data, param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	val, errmsg := keeper.GetQueryResultSimple(ctx, param.KeyType, string(param.KeyData), param.Path)
	if errmsg != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, errmsg.Error())
	}

	qryvalue := QueryExecutionLayerResp{
		Value: val.String(),
	}

	res, _ := codec.MarshalJSONIndent(keeper.cdc, qryvalue)
	return res, nil
}
