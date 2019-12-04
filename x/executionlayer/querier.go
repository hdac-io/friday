package executionlayer

import (
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/hdac-io/friday/x/executionlayer/types"
)

const (
	QueryEE              = "query"
	QueryEEDetail        = "querydetail"
	QueryEEBalance       = "querybalance"
	QueryEEBalanceDetail = "querybalancedetail"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper ExecutionLayerKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryEEDetail:
			return queryEEDetail(ctx, path[1:], req, keeper)
		case QueryEE:
			return queryEE(ctx, path[1:], req, keeper)
		case QueryEEBalance:
			return queryBalance(ctx, path[1:], req, keeper)
		case QueryEEBalanceDetail:
			return queryBalanceDetail(ctx, path[1:], req, keeper)

		default:
			return nil, sdk.ErrUnknownRequest("unknown ee query")
		}
	}
}

func queryEEDetail(ctx sdk.Context, path []string, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryExecutionLayerDetail
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	val, errmsg := keeper.GetQueryResult(ctx, param.StateHash, param.KeyType, param.KeyData, param.Path)
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
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	val, errmsg := keeper.GetQueryResultSimple(ctx, param.KeyType, param.KeyData, param.Path)
	if errmsg != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, errmsg.Error())
	}

	qryvalue := QueryExecutionLayerResp{
		Value: val.String(),
	}

	res, _ := codec.MarshalJSONIndent(keeper.cdc, qryvalue)
	return res, nil
}

func queryBalanceDetail(ctx sdk.Context, path []string, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryGetBalanceDetail
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	address, err := sdk.AccAddressFromBech32(param.Address)
	if err != nil {
		return nil, sdk.NewError(sdk.Bech32PrefixValAddr, sdk.CodeInvalidAddress, err.Error())
	}
	val, err := keeper.GetQueryBalanceResult(ctx, param.StateHash, types.ToPublicKey(address))
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, err.Error())
	}

	queryvalue := QueryExecutionLayerResp{
		Value: val,
	}

	res, _ := codec.MarshalJSONIndent(keeper.cdc, queryvalue)
	return res, nil
}

func queryBalance(ctx sdk.Context, path []string, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryGetBalance
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	address, err := sdk.AccAddressFromBech32(param.Address)
	if err != nil {
		return nil, sdk.NewError(sdk.Bech32PrefixValAddr, sdk.CodeInvalidAddress, err.Error())
	}
	val, err := keeper.GetQueryBalanceResultSimple(ctx, types.ToPublicKey(address))
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, err.Error())
	}

	queryvalue := QueryExecutionLayerResp{
		Value: val,
	}

	res, _ := codec.MarshalJSONIndent(keeper.cdc, queryvalue)
	return res, nil
}
