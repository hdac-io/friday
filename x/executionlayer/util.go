package executionlayer

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
	"github.com/hdac-io/friday/x/nickname"
)

// TODO: change KeyType from string to typed enum.
func toBytes(keyType string, key string,
	k nickname.NicknameKeeper, ctx sdk.Context) ([]byte, error) {
	var bytes []byte
	var err error

	switch keyType {
	case types.ADDRESS:
		// we have special key for system account
		if key == types.SYSTEM {
			bytes = types.SYSTEM_ACCOUNT
			break
		}

		var addr sdk.AccAddress
		addr, err = sdk.AccAddressFromBech32(key)
		if err != nil {
			acc := k.GetUnitAccount(ctx, key)
			if acc.Nickname.MustToString() == "" {
				err = fmt.Errorf("no readable ID mapping of %s", key)
				break
			}
			addr = acc.Address
			err = nil
		}
		bytes = addr.Bytes()

	case types.UREF:
		urefbytes, err := sdk.ContractUrefAddressFromBech32(key)
		if err != nil {
			break
		}
		bytes = urefbytes.Bytes()

	case types.HASH:
		hashbytes, err := sdk.ContractHashAddressFromBech32(key)
		if err != nil {
			break
		}
		bytes = hashbytes.Bytes()

	case types.LOCAL:
		bytes, err = hex.DecodeString(key)

	default:
		err = fmt.Errorf("Unknown QueryKey type: %v", keyType)
	}

	if err != nil {
		bytes = nil
	}

	return bytes, err
}

func DeployArgsToJsonString(args []*consensus.Deploy_Arg) (string, error) {
	m := &jsonpb.Marshaler{}
	str := "["
	for idx, arg := range args {
		if idx != 0 {
			str += ","
		}
		s, err := m.MarshalToString(arg)
		if err != nil {
			return "", err
		}
		str += s
	}
	str += "]"

	return str, nil
}

func ReplaceFromBech32ToHex(isCustomContractRun bool, valueStr string) (string, []sdk.AccAddress, error) {
	res := valueStr
	addrList := []sdk.AccAddress{}
	if isCustomContractRun {
		r := regexp.MustCompile(fmt.Sprintf(`\"hash\":\{\"hash\":\"(%s[a-zA-Z0-9+/]+)\"`, sdk.Bech32PrefixContractHash))
		for _, matchedGroup := range r.FindAllStringSubmatch(valueStr, -1) {
			hashStr := matchedGroup[1]
			hashaddr, err := sdk.ContractHashAddressFromBech32(hashStr)
			if err != nil {
				return valueStr, []sdk.AccAddress{}, err
			}
			hashaddrhex := base64.StdEncoding.EncodeToString(hashaddr.Bytes())

			filterHashStr := `"hash":{"hash":"` + hashStr
			replaceStr := `"hash":{"hash":"` + hashaddrhex
			res = strings.Replace(res, filterHashStr, replaceStr, -1)
		}

		r = regexp.MustCompile(fmt.Sprintf(`\"uref\":\{\"uref\":\"(%s[a-zA-Z0-9+/]+)\"`, sdk.Bech32PrefixContractURef))
		for _, matchedGroup := range r.FindAllStringSubmatch(valueStr, -1) {
			urefStr := matchedGroup[1]
			urefaddr, err := sdk.ContractUrefAddressFromBech32(urefStr)
			if err != nil {
				return valueStr, []sdk.AccAddress{}, err
			}
			urefaddrhex := base64.StdEncoding.EncodeToString(urefaddr.Bytes())

			filterUrefStr := `"uref":{"uref":"` + urefStr
			replaceStr := `"uref":{"uref":"` + urefaddrhex
			res = strings.Replace(res, filterUrefStr, replaceStr, -1)
		}

		r = regexp.MustCompile(fmt.Sprintf(`{\"name\":\"address\",\"value\":{\"cl_type\":\{\"list\_type\":\{\"inner\":\{\"simple_type\":\"U8\"\}\}\},\"value\":\{\"bytes\_value\":\"(%s[a-zA-Z0-9+/]+)\"\}\}\}`, sdk.Bech32PrefixAccAddr))
		for _, matchedGroup := range r.FindAllStringSubmatch(valueStr, -1) {
			accountStr := matchedGroup[1]
			accountAddr, err := sdk.AccAddressFromBech32(accountStr)
			if err != nil {
				return valueStr, []sdk.AccAddress{}, err
			}
			addrList = append(addrList, accountAddr)
			accountHex := base64.StdEncoding.EncodeToString(accountAddr.Bytes())

			filterAccountStr := `{"name":"address","value":{"cl_type":{"list_type":{"inner":{"simple_type":"U8"}}},"value":{"bytes_value":"` + accountStr
			replaceStr := `{"name":"address","value":{"cl_type":{"list_type":{"inner":{"simple_type":"U8"}}},"value":{"bytes_value":"` + accountHex
			res = strings.Replace(res, filterAccountStr, replaceStr, -1)
		}
	}

	return res, addrList, nil
}
