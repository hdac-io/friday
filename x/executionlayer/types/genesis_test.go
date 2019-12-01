package types

import (
	"os"
	"reflect"
	"testing"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	sdk "github.com/hdac-io/friday/types"
	"github.com/stretchr/testify/require"
)

const (
	mintCodePath = "$HOME/.fryd/contracts/mint_install.wasm"
	posCodePath  = "$HOME/.fryd/contracts/pos_install.wasm"
)

func TestToProtocolVersion(t *testing.T) {
	// empty string
	got, err := toProtocolVersion("")
	require.NotNil(t, err)
	require.Nil(t, got)

	// just a number
	got, err = toProtocolVersion("123")
	require.NotNil(t, err)
	require.Nil(t, got)

	// trailing dot
	got, err = toProtocolVersion("1.0.0.")
	require.NotNil(t, err)
	require.Nil(t, got)

	// too many digit
	got, err = toProtocolVersion("1.0.0.0")
	require.NotNil(t, err)
	require.Nil(t, got)

	// valid case
	got, err = toProtocolVersion("123.456.789")
	require.Nil(t, err)
	expected := state.ProtocolVersion{Major: 123, Minor: 456, Patch: 789}
	if !reflect.DeepEqual(expected, *got) {
		t.Errorf("Bad protocol version parsing. expected %v, got %v", expected, got)
	}
}

func TestToChainSpecGenesisAccount(t *testing.T) {
	// valid input
	addr, _ := sdk.AccAddressFromBech32("friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz")
	account := Account{
		PublicKey:           addr,
		InitialBalance:      "100000000",
		InitialBondedAmount: "10000",
	}
	_, err := toChainSpecGenesisAccount(account)
	require.Nil(t, err)

	// The reason of the blocked tests below: AccAddressFromBech32 type blocks malformed address

	// account.PublicKey = "invalid-public-key"
	// _, err = toChainSpecGenesisAccount(account)
	// require.NotNil(t, err)

	// account.PublicKey = base64.StdEncoding.EncodeToString([]byte("invalid-public-key"))
	// _, err = toChainSpecGenesisAccount(account)
	// require.NotNil(t, err)
}

func TestToChainSpecGenesisConfig(t *testing.T) {
	// valid input
	genesisState := DefaultGenesisState()
	genesisState.GenesisConf.Genesis.MintCodePath = os.ExpandEnv(mintCodePath)
	genesisState.GenesisConf.Genesis.PosCodePath = os.ExpandEnv(posCodePath)
	_, err := ToChainSpecGenesisConfig(genesisState.GenesisConf)
	require.Nil(t, err)

	// invalid system contract path
	genesisState.GenesisConf.Genesis.MintCodePath = "test-odd-path"
	_, err = ToChainSpecGenesisConfig(genesisState.GenesisConf)
	require.NotNil(t, err)

	genesisState.GenesisConf.Genesis.MintCodePath = os.ExpandEnv(mintCodePath)
	genesisState.GenesisConf.Genesis.PosCodePath = "test-odd-path"
	_, err = ToChainSpecGenesisConfig(genesisState.GenesisConf)
	require.NotNil(t, err)

	// revert to valid input
	genesisState.GenesisConf.Genesis.MintCodePath = os.ExpandEnv(mintCodePath)
	genesisState.GenesisConf.Genesis.PosCodePath = os.ExpandEnv(posCodePath)

	// invalid protocol version
	genesisState.GenesisConf.Genesis.ProtocolVersion = "1.0.0.0"
	_, err = ToChainSpecGenesisConfig(genesisState.GenesisConf)
	require.NotNil(t, err)

	// revert to valid input
	genesisState.GenesisConf.Genesis.ProtocolVersion = "1.0.0"

	// The reason of the blocked tests below: AccAddressFromBech32 type blocks malformed address

	// invalid account(not conformant to base64 encoded public key)
	// genesisState.GenesisConf.Genesis.Accounts = make([]Account, 1)
	// genesisState.GenesisConf.Genesis.Accounts[0] = Account{
	// 	PublicKey:           "invalid-public-key",
	// 	InitialBalance:      "100000000",
	// 	InitialBondedAmount: "10000",
	// }
	// _, err = ToChainSpecGenesisConfig(genesisState.GenesisConf)
	// require.NotNil(t, err)

	// genesisState.GenesisConf.Genesis.Accounts[0].PublicKey =
	// 	base64.StdEncoding.EncodeToString([]byte("invalid-public-key"))
	// _, err = ToChainSpecGenesisConfig(genesisState.GenesisConf)
	// require.NotNil(t, err)
}
