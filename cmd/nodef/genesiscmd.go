package main

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/hdac-io/friday/client/keys"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/server"
	sdk "github.com/hdac-io/friday/types"
	elconfig "github.com/hdac-io/friday/x/executionlayer/configuration"
	eltypes "github.com/hdac-io/friday/x/executionlayer/types"
	"github.com/hdac-io/friday/x/genutil"
)

const (
	flagClientHome = "home-client"
)

// AddElGenesisAccountCmd returns add-genesis-account cobra Command.
func AddElGenesisAccountCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {

	cmd := &cobra.Command{
		Use:   `add-el-genesis-account [bech32 string] [initial_balance] [initial_bonded_amount]`,
		Short: "Add a genesis account to genesis.json",
		Long: `Add a genesis account to genesis.json. The provided account must specify
the base64 encoded publickey and a list of initial coins.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			accAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				kb, err := keys.NewKeyBaseFromDir(viper.GetString(flagClientHome))
				if err != nil {
					return err
				}

				info, err := kb.Get(args[0])
				if err != nil {
					return err
				}

				accAddr = info.GetAddress()
			}

			// Use sdk.AccAddress as public key for PoC.
			// It should be replaced with a raw public key later.
			account := eltypes.Account{
				PublicKey:           eltypes.ToPublicKey(accAddr),
				InitialBalance:      args[1],
				InitialBondedAmount: args[2],
			}

			// get genesis file
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// retrieve genesis state for execution layer
			var genesisState eltypes.GenesisState
			if appState[eltypes.ModuleName] != nil {
				cdc.MustUnmarshalJSON(appState[eltypes.ModuleName], &genesisState)
			}

			// check already existing address
			for _, v := range genesisState.Accounts {
				if bytes.Compare([]byte(v.PublicKey), []byte(account.PublicKey)) == 0 {
					return fmt.Errorf("already existing address: %v", args[0])
				}
			}

			// append an account
			genesisState.Accounts = append(genesisState.Accounts, account)
			genesisStateBytes, err := cdc.MarshalJSON(genesisState)
			if err != nil {
				return fmt.Errorf("failed to marshal executionlayer genesis state: %w", err)
			}
			appState[eltypes.ModuleName] = genesisStateBytes

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")

	return cmd
}

// LoadChainspecCmd returns load-chainspec cobra Command.
func LoadChainspecCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   `load-chainspec [chainspec toml file path]`,
		Short: "Load a executionlayer genesis config to genesis.json",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// get genesis file
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			// retrieve genesis state for executionlayer
			var genesisState eltypes.GenesisState
			if appState[eltypes.ModuleName] != nil {
				cdc.MustUnmarshalJSON(appState[eltypes.ModuleName], &genesisState)
			}

			// parse chainspec toml
			genesisConf, err := elconfig.ParseGenesisChainSpec(args[0])
			if err != nil {
				return err
			}
			genesisState.GenesisConf = *genesisConf

			// Marshall GenesisState
			genesisStateBytes, err := cdc.MarshalJSON(genesisState)
			if err != nil {
				return fmt.Errorf("failed to marshal executionlayer genesis state: %w", err)
			}
			appState[eltypes.ModuleName] = genesisStateBytes

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")

	return cmd
}
