package util

import (
	"fmt"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	idtype "github.com/hdac-io/friday/x/nickname/types"
)

// GetAddress searches address in nickname mapping
func GetAddress(cdc *codec.Codec, cliCtx context.CLIContext, addressOrName string) (sdk.AccAddress, error) {
	var address sdk.AccAddress
	address, err := sdk.AccAddressFromBech32(addressOrName)
	if err != nil {
		queryData := idtype.QueryReqUnitAccount{
			Nickname: addressOrName,
		}
		bz := cdc.MustMarshalJSON(queryData)

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaddress", idtype.StoreKey), bz)
		if err != nil {
			return nil, err
		}
		var out idtype.QueryResUnitAccount
		cdc.MustUnmarshalJSON(res, &out)
		address = out.Address
	}

	return address, nil
}
