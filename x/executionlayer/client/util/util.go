package util

import (
	"fmt"

	"github.com/hdac-io/tendermint/crypto"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	idtype "github.com/hdac-io/friday/x/readablename/types"
)

// GetPubKey searches public key in readable ID service mapping
func GetPubKey(cdc *codec.Codec, cliCtx context.CLIContext, pubkeyOrName string) (crypto.PubKey, error) {
	fromPubkey, err := sdk.GetAccPubKeyBech32(pubkeyOrName)
	if err != nil {
		queryData := idtype.QueryReqUnitAccount{
			Name: pubkeyOrName,
		}
		bz := cdc.MustMarshalJSON(queryData)

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaccount", idtype.StoreKey), bz)
		if err != nil {
			return nil, err
		}
		var out idtype.QueryResUnitAccount
		cdc.MustUnmarshalJSON(res, &out)
		fromPubkey = out.PubKey
	}

	return fromPubkey, nil
}
