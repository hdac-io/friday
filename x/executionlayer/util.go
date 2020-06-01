package executionlayer

import (
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

func ReplaceFromBech32ToHex(valueStr string) (string, error) {
	res := valueStr

	r := regexp.MustCompile(`\"hash\":\{\"hash\":\"(fridaycontracthash[a-zA-Z0-9+/]+)\"`)
	for _, matchedGroup := range r.FindAllStringSubmatch(valueStr, -1) {
		hashStr := matchedGroup[1]
		hashaddr, err := sdk.ContractHashAddressFromBech32(hashStr)
		if err != nil {
			return valueStr, err
		}
		hashaddrhex := hex.EncodeToString(hashaddr.Bytes())
		res = strings.Replace(res, hashStr, hashaddrhex, -1)
	}

	r = regexp.MustCompile(`\"uref\":\{\"uref\":\"(fridaycontracturef[a-zA-Z0-9+/]+)\"`)
	for _, matchedGroup := range r.FindAllStringSubmatch(valueStr, -1) {
		urefStr := matchedGroup[1]
		urefaddr, err := sdk.ContractUrefAddressFromBech32(urefStr)
		if err != nil {
			return valueStr, err
		}
		urefaddrhex := hex.EncodeToString(urefaddr.Bytes())
		res = strings.Replace(res, urefStr, urefaddrhex, -1)
	}

	r = regexp.MustCompile(`\"address\":\{\"account\":\"(friday[a-zA-Z0-9+/]+)\"`)
	for _, matchedGroup := range r.FindAllStringSubmatch(valueStr, -1) {
		accountStr := matchedGroup[1]
		accountAddr, err := sdk.AccAddressFromBech32(accountStr)
		if err != nil {
			return valueStr, err
		}
		accountHex := hex.EncodeToString(accountAddr.Bytes())
		res = strings.Replace(res, accountStr, accountHex, -1)
	}

	return res, nil
}
