package readablename

import (
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"
)

// Query endpoints definition for GET request
const (
	QueryGetAccount = "getaccount"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k ReadableNameKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetAccount:
			return queryUnitAccount(ctx, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown readable name query endpoint")
		}
	}
}

func queryUnitAccount(ctx sdk.Context, path []string, req abci.RequestQuery, k ReadableNameKeeper) ([]byte, sdk.Error) {
	var param QueryReqUnitAccount
	err := ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, err.Error())
	}

	value := k.GetUnitAccount(ctx, param.Name)
	qryvalue := QueryResUnitAccount{
		Name:    value.Name.MustToString(),
		Address: value.Address,
		PubKey:  value.PubKey,
	}
	res, _ := codec.MarshalJSONIndent(k.cdc, qryvalue)
	return res, nil
}
