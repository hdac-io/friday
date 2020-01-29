package nickname

import (
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/nickname/types"
	abci "github.com/hdac-io/tendermint/abci/types"
)

// Query endpoints definition for GET request
const (
	QueryGetAccount = "getaddress"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k NicknameKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetAccount:
			return queryUnitAccount(ctx, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown readable name query endpoint")
		}
	}
}

func queryUnitAccount(ctx sdk.Context, path []string, req abci.RequestQuery, k NicknameKeeper) ([]byte, sdk.Error) {
	var param QueryReqUnitAccount
	err := ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, types.ErrBadQueryRequest(ModuleName)
	}

	value := k.GetUnitAccount(ctx, param.Nickname)
	if value.Nickname.MustToString() == "" {
		return nil, types.ErrNoRegisteredReadableID(ModuleName, param.Nickname)
	}

	qryvalue := QueryResUnitAccount{
		Nickname: value.Nickname.MustToString(),
		Address:  value.Address,
	}
	res, _ := codec.MarshalJSONIndent(k.cdc, qryvalue)
	return res, nil
}
