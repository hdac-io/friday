package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/hdac-io/friday/client/keys"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/server"
	sdk "github.com/hdac-io/friday/types"
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
			publicKey := eltypes.ToPublicKey(accAddr)
			balance := args[1]
			bondedAmount := args[2]

			// create concrete account type based on input parameters
			account := eltypes.Account{
				PublicKey:           publicKey,
				InitialBalance:      balance,
				InitialBondedAmount: bondedAmount,
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

			genesisState.GenesisConf.Genesis.Accounts = append(genesisState.GenesisConf.Genesis.Accounts, account)
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