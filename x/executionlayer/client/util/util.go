package util

import (
	"fmt"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	idtype "github.com/hdac-io/friday/x/readablename/types"
	"github.com/hdac-io/tendermint/crypto/secp256k1"
)

// GetPubKey searches public key in readable ID service mapping
func GetPubKey(cdc *codec.Codec, cliCtx context.CLIContext, pubkeyOrName string) (*secp256k1.PubKeySecp256k1, error) {
	fromPubkey, err := sdk.GetSecp256k1FromRawHexString(pubkeyOrName)
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
		fromPubkey = &out.PubKey
	}

	return fromPubkey, nil
}
