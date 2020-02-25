package executionlayer

import (
	"encoding/hex"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/nickname"
)

// TODO: change KeyType from string to typed enum.
func toBytes(keyType string, key string,
	k nickname.NicknameKeeper, ctx sdk.Context) ([]byte, error) {
	var bytes []byte
	var err error

	switch keyType {
	case "address":
		// we have special key for system account
		if key == "system" {
			bytes = make([]byte, 32)
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
		bytes = addr.ToEEAddress().Bytes()

	case "uref", "local", "hash":
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
