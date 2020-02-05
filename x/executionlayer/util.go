package executionlayer

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/nickname"
)

// TODO: change KeyType from string to typed enum.
func toBytes(keyType string, key string,
	k nickname.NicknameKeeper, ctx sdk.Context) ([]byte, error) {
	var bytes []byte
	var err error = nil

	switch keyType {
	case "address":
		var addr sdk.AccAddress
		addr, err := sdk.AccAddressFromBech32(key)
		if err != nil {
			acc := k.GetUnitAccount(ctx, key)
			if acc.Nickname.MustToString() == "" {
				return nil, fmt.Errorf("no readable ID mapping of %s", key)
			}
			addr = acc.Address
		}
		bytes = addr.ToEEAddress().Bytes()

	case "uref", "local", "hash":
		bytes, err = hex.DecodeString(key)

	default:
		err = fmt.Errorf("Unknown QueryKey type: %v", keyType)
	}

	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func isEmptyBytes(src []byte) bool {
	return bytes.Equal([]byte{}, src)
}

func ProtobufSafeEncodeBytes(src []byte) []byte {
	if bytes.Equal(src, []byte("empty")) {
		src = []byte{}
	}

	return src
}
