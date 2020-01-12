package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	cliutil "github.com/hdac-io/friday/x/executionlayer/client/util"
	"github.com/hdac-io/tendermint/crypto/secp256k1"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetExecutionLayerQueryCmd controls GET type CLI controller
func GetExecutionLayerQueryCmd(cdc *codec.Codec) *cobra.Command {
	executionlayerQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for execution layer",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	executionlayerQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryBalance(cdc),
		GetCmdQuery(cdc),
	)...)
	return executionlayerQueryCmd
}

// GetCmdQueryBalance is a getter of the balance of the address
func GetCmdQueryBalance(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("getbalance [--name \"readable_id\" or --pubkey \"secp256k1_pubkey\" or --%spub \"bech32\"] [--blockhash]", sdk.Bech32MainPrefix),
		Short: "Get balance of address",
		Long: fmt.Sprintf("Get balance of address.\nIt needs at least one of \"--name\", \"--pubkey\", or \"--%[1]spub\" parameter.\n"+
			"\t--name: readabld ID\n"+
			"\t--pubkey: Compressed Secp256k1 public key\n"+
			"\t--%[1]spub: Bech32 encoded public key starting from '%[1]s'\n"+
			"If you need the value of specific time, use \"--blockhash\" option.\n"+
			"\t--blockhash: Hex blockhash value which represents of the specific time. (Default: the latest state)", sdk.Bech32MainPrefix),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Get public key with 3 different types
			var pubkey secp256k1.PubKeySecp256k1

			if name := viper.GetString(client.FlagName); name != "" {
				// --name: from readable id
				pubkeyPtr, err := cliutil.GetPubKey(cdc, cliCtx, name)
				if err != nil {
					return err
				}
				pubkey = *pubkeyPtr
			} else if rawPubkey := viper.GetString(client.FlagName); rawPubkey != "" {
				// --pubkey: from raw secp256k1 public key
				pubkeyPtr, err := sdk.GetSecp256k1FromRawHexString(rawPubkey)
				if err != nil {
					return err
				}
				pubkey = *pubkeyPtr
			} else if bech32Pubkey := viper.GetString(client.FlagName); bech32Pubkey != "" {
				// --[bech32_prefix]pub: from bech32 public key (fridaypubxxxxxx...)
				rawPubkey, err := sdk.GetSecp256k1FromBech32AccPubKey(bech32Pubkey)
				if err != nil {
					return err
				}
				pubkey = *rawPubkey
			} else {
				return fmt.Errorf("at least one of --name, --pubkey, or --%[1]spub is essential", sdk.Bech32MainPrefix)
			}

			var out types.QueryExecutionLayerResp
			if blockhashstr := viper.GetString(FlagBlockHash); blockhashstr != "" {
				blockHash, err := hex.DecodeString(blockhashstr)
				if err != nil || len(blockHash) != 32 {
					fmt.Println("Malformed block hash - ", blockhashstr)
					fmt.Println(err)
					return nil
				}

				queryData := types.QueryGetBalanceDetail{
					PublicKey: pubkey,
					StateHash: blockHash,
				}
				bz := cdc.MustMarshalJSON(queryData)

				res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalancedetail", types.ModuleName), bz)
				if err != nil {
					fmt.Printf("No balance data of input")
					return nil
				}
				cdc.MustUnmarshalJSON(res, &out)

			} else {
				queryData := types.QueryGetBalance{
					PublicKey: pubkey,
				}
				bz := cdc.MustMarshalJSON(queryData)

				res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalance", types.ModuleName), bz)
				if err != nil {
					fmt.Printf("No balance data of input")
					fmt.Println(err.Error())
					return nil
				}
				cdc.MustUnmarshalJSON(res, &out)
			}

			return cliCtx.PrintOutput(out)
		},
	}

	cmd.Flags().String(client.FlagName, "", "flag for readable name input")
	cmd.Flags().String(FlagPubKey, "", "flag for secp256k1 public key input")
	cmd.Flags().String(FlagBech32PubKey, "", "flag for bech32 public key input")
	cmd.Flags().String(FlagBlockHash, "", "flag for block hash input")

	return cmd
}

// GetCmdQuery is a EE query getter
func GetCmdQuery(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query [type:=address,uref,hash,local] [data] [path] [--blockhash]",
		Short: "Get query of the data",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			dataType := args[0]
			data := args[1]
			path := args[2]

			var out types.QueryExecutionLayerResp
			if blockhashstr := viper.GetString(FlagBlockHash); blockhashstr != "" {
				blockhash, err := hex.DecodeString(blockhashstr)
				if err != nil || len(blockhash) != 32 {
					fmt.Println("Malformed block hash - ", blockhashstr)
					fmt.Println(err)
					return nil
				}
				queryData := types.QueryExecutionLayerDetail{
					KeyType:   dataType,
					KeyData:   data,
					Path:      path,
					StateHash: blockhash,
				}
				bz := cdc.MustMarshalJSON(queryData)

				res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querydetail", types.ModuleName), bz)
				if err != nil {
					fmt.Printf("could not resolve data - %s %s %s\n", dataType, data, path)
					return nil
				}

				cdc.MustUnmarshalJSON(res, &out)
			} else {
				queryData := types.QueryExecutionLayer{
					KeyType: dataType,
					KeyData: data,
					Path:    path,
				}
				bz := cdc.MustMarshalJSON(queryData)

				res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/query", types.ModuleName), bz)
				if err != nil {
					fmt.Printf("could not resolve data - %s %s %s\n", dataType, data, path)
					return nil
				}

				cdc.MustUnmarshalJSON(res, &out)
			}
			return cliCtx.PrintOutput(out)
		},
	}

	cmd.Flags().String(FlagBlockHash, "", "flag for block hash input")
	return cmd
}
