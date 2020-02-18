package util

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/crypto/keys"
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

// GetLocalWalletInfo takes wallet info from local
// Rules:
// 1) If --from exists, search according to:
//  (1) Local wallet alias
//  (2) Address
//  (3) Nickname
// 2) If --from doesn't exist,
//  (1) If only one wallet exists, take this
//  (2) If not, raise error
func GetLocalWalletInfo(valueFromFromFlag string, kb keys.Keybase, cdc *codec.Codec, cliCtx context.CLIContext) (keys.Info, error) {
	if valueFromFromFlag != "" {
		// Find in local wallet
		key, err := kb.Get(valueFromFromFlag)
		if err == nil {
			return key, nil
		}

		// If not exist, try parsing into address and find in nickname
		addr, err := GetAddress(cdc, cliCtx, valueFromFromFlag)
		if err != nil {
			return nil, fmt.Errorf("cannot parse into address, or no registered address of the given nickname '%s'", valueFromFromFlag)
		}
		key, err = kb.GetByAddress(addr)
		if err != nil {
			return nil, err
		}
		return key, nil
	}

	infoList, err := kb.List()
	if err != nil {
		return nil, err
	}

	if len(infoList) > 1 {
		return nil, fmt.Errorf("multiple wallets in local. Cannot specify one wallet")
	} else if len(infoList) == 0 {
		return nil, fmt.Errorf("no wallet in local")
	}

	return infoList[0], nil
}
