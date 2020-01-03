package executionlayer

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/readablename"
)

// TODO: change KeyType from string to typed enum.
func toBytes(keyType string, key string,
	k readablename.ReadableNameKeeper, ctx sdk.Context) ([]byte, error) {
	var bytes []byte
	var err error = nil

	switch keyType {
	case "address":
		pubkeybytes, err := sdk.GetAccPubKeyBech32(key)
		if err != nil {
			pubkeybytes = k.GetUnitAccount(ctx, key).PubKey
			return nil, err
		}
		bytes = pubkeybytes.Bytes()
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
