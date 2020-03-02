package util

import (
	"fmt"
	"regexp"
	"strings"

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

type Hdac string
type Bigsun string

func ToBigsun(hdac Hdac) (Bigsun, error) {
	src := string(hdac)
	// validation
	resRegexp, err := regexp.MatchString("^[0-9|.]*$", src)
	if !resRegexp {
		errStr := "Must be number or '.'"
		if err != nil {
			errStr += err.Error()
		}
		return Bigsun("0"), fmt.Errorf(errStr)
	}

	// convert
	ress := strings.Split(src, ".")

	if len(ress) > 2 {
		return Bigsun("0"), fmt.Errorf("'.' must be less than or equal to 1, But %d", len(ress))
	}

	res := strings.Join(ress, "")

	if strings.Count(res, "0") == len(res) {
		return Bigsun("0"), nil
	}

	paddingCount := 18
	if len(res) != len(ress[0]) {
		paddingCount -= len(ress[1])
		if paddingCount < 0 {
			return Bigsun("0"), fmt.Errorf("The decimal place must be 18 digits or less, But %d", len(ress[1]))
		}
	}
	res += strings.Repeat("0", paddingCount)

	for i := 0; i < len(res); i++ {
		if res[i] != '0' {
			res = res[i:]
			break
		}
	}

	return Bigsun(res), nil
}

func ToHdac(Bigsun Bigsun) Hdac {
	src := string(Bigsun)

	res := []string{"0", ""}
	if len(src) > 18 {
		res[0] = src[:len(src)-18]
		res[1] = src[len(src)-18:]
	} else {
		if src == "0" {
			return Hdac(src)
		}
		res[1] = strings.Repeat("0", 18-len(src)) + src
	}

	i := len(res[1]) - 1
	for ; i >= 0; i-- {
		if res[1][i] != '0' {
			res[1] = res[1][:i+1]
			break
		}
	}

	if i < 0 {
		return Hdac(res[0])
	}

	return Hdac(strings.Join(res, "."))
}
