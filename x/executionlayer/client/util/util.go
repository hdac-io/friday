package util

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	idtype "github.com/hdac-io/friday/x/nickname/types"
)

// GetAddress searches address in nickname mapping
func GetAddress(cdc *codec.Codec, cliCtx context.CLIContext, addressOrName string) (sdk.AccAddress, error) {
	var address sdk.AccAddress
	var err error
	address, err = sdk.AccAddressFromBech32(addressOrName)
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

func GetContractType(strContractType string) util.ContractType {
	var contractType util.ContractType
	switch strContractType {
	case "wasm":
		contractType = util.WASM
	case "uref":
		contractType = util.UREF
	case "hash":
		contractType = util.HASH
	case "name":
		contractType = util.NAME
	}
	return contractType
}

// ProtobufSafeDecodeingHexString is decode to string.
// The encoding and decoding data for nil is different. Use it to avoid that part.
func ProtobufSafeDecodeingHexString(str string) ([]byte, error) {
	res, err := hex.DecodeString(str)
	if err != nil {
		return []byte{}, err
	}

	if bytes.Equal(res, []byte{}) {
		res = []byte("empty")
	}

	return res, nil
}
