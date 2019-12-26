package executionlayer

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// TODO: change KeyType from string to typed enum.
func toBytes(keyType string, key string) ([]byte, error) {
	var bytes []byte
	var err error = nil

	switch keyType {
	case "address":
		bech32addr, err := sdk.AccAddressFromBech32(key)
		if err != nil {
			return nil, err
		}
		bytes = types.ToPublicKey(bech32addr)
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
