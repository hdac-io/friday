package readablename

import (
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"
)

// Query endpoints definition for GET request
const (
	//QueryAccountsList = "accountslist"
	//QueryGetAddress   = "getaddress"
	QueryGetAccount = "getaccount"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k AccountKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetAccount:
			return queryUnitAccount(ctx, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown nameservice query endpoint")
		}
	}
}

func queryUnitAccount(ctx sdk.Context, path []string, req abci.RequestQuery, k AccountKeeper) ([]byte, sdk.Error) {
	value := k.GetUnitAccount(ctx, path[0])
	strname, _ := value.ID.ToString()
	qryvalue := QueryResUnitAccount{
		ID:      strname,
		Address: value.Address,
	}
	res, _ := codec.MarshalJSONIndent(k.cdc, qryvalue)
	return res, nil
}
