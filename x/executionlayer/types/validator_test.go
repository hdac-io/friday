package types

import (
	"testing"

	sdk "github.com/hdac-io/friday/types"
	"github.com/stretchr/testify/require"
)

func TestValidatorTestEquivalent(t *testing.T) {
	accAddr := "friday19rxdgfn3grqgwc6zhyeljmyas3tsawn6qe0quc"
	acc, _ := sdk.AccAddressFromBech32(accAddr)
	valAddr := sdk.ValAddress(acc)

	require.Equal(t, "fridayvaloper19rxdgfn3grqgwc6zhyeljmyas3tsawn64dsges", valAddr.String())
	eeAddress := acc.ToEEAddress()

	consPubKey, _ := sdk.GetConsPubKeyBech32("fridayvalconspub16jrl8jvqq98x7jjxfcm8252pwd4nv6fetpzk6nzx2ddyc3fn0p2rz4mwf44nqjtfga5k5at4xad82sjhx9r9zdfcwuc5uvt90934jjr4d4xk242909rxks28v9erv3jvwfcx2wp4fe8h54fsddu9zar5v3tyknrs8pykk2mw2p29j4n6w455c7j2d3x4ykft9akx6s24gsu8ys2nvayrykqst965z")
	val1 := NewValidator(eeAddress, consPubKey, Description{}, "0")
	val2 := NewValidator(eeAddress, consPubKey, Description{}, "0")

	ok := val1.TestEquivalent(val2)
	require.True(t, ok)
}

func TestUpdateDescription(t *testing.T) {
	d1 := Description{
		Website: "https://validator.friday",
		Details: "Test validator",
	}

	d2 := Description{
		Moniker:  DoNotModifyDesc,
		Identity: DoNotModifyDesc,
		Website:  DoNotModifyDesc,
		Details:  DoNotModifyDesc,
	}

	d3 := Description{
		Moniker:  "",
		Identity: "",
		Website:  "",
		Details:  "",
	}

	d, err := d1.UpdateDescription(d2)
	require.Nil(t, err)
	require.Equal(t, d, d1)

	d, err = d1.UpdateDescription(d3)
	require.Nil(t, err)
	require.Equal(t, d, d3)
}
