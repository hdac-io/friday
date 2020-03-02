package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	cliutil "github.com/hdac-io/friday/x/executionlayer/client/util"

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
		Use:   "getbalance --from <from> [--blockhash <blockhash_since>]",
		Short: "Get balance of address",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			var addr sdk.AccAddress
			var err error
			addr, err = cliutil.GetAddress(cdc, cliCtx, valueFromFromFlag)
			if err != nil {
				kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
				if err != nil {
					return err
				}

				keyInfo, err := kb.Get(valueFromFromFlag)
				if err != nil {
					return err
				}

				addr = keyInfo.GetAddress()
			}

			var out types.QueryExecutionLayerResp
			if blockhashstr := viper.GetString(FlagBlockHash); blockhashstr != "" {
				blockHash, err := hex.DecodeString(blockhashstr)
				if err != nil || len(blockHash) != 32 {
					fmt.Println("malformed block hash - ", blockhashstr)
					fmt.Println(err)
					return nil
				}

				queryData := types.QueryGetBalanceDetail{
					Address:   addr,
					StateHash: blockHash,
				}
				bz := cdc.MustMarshalJSON(queryData)

				res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalancedetail", types.ModuleName), bz)
				if err != nil {
					fmt.Printf("no balance data of input")
					return nil
				}
				cdc.MustUnmarshalJSON(res, &out)

			} else {
				queryData := types.QueryGetBalance{
					Address: addr,
				}
				bz := cdc.MustMarshalJSON(queryData)

				res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalance", types.ModuleName), bz)
				if err != nil {
					fmt.Printf("no balance data of input")
					fmt.Println(err.Error())
					return nil
				}
				cdc.MustUnmarshalJSON(res, &out)
			}

			out.Value = string(cliutil.ToHdac(cliutil.Bigsun(out.Value)))

			return cliCtx.PrintOutput(out)
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")
	cmd.Flags().String(FlagBlockHash, "", "Block hash at the moment")

	return cmd
}

// GetCmdQuery is a EE query getter
func GetCmdQuery(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query address|uref|hash|local <data> <path> [--blockhash <blockhash_since>]",
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
					fmt.Println("malformed block hash - ", blockhashstr)
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

	cmd.Flags().String(FlagBlockHash, "", "Block hash at the moment")
	return cmd
}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator --from <from>",
		Short: "Query a validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			addr, err := cliutil.GetAddress(cdc, cliCtx, valueFromFromFlag)
			if err != nil {
				return err
			}

			if addr.Empty() {
				res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/queryallvalidator", types.ModuleName))
				if err != nil {
					fmt.Printf("could not resolve")
					return nil
				}

				var out types.Validators
				cdc.MustUnmarshalJSON(res, &out)

				return cliCtx.PrintOutput(out)
			} else {
				queryData := types.NewQueryValidatorParams(addr)
				bz := cdc.MustMarshalJSON(queryData)

				res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/queryvalidator", types.ModuleName), bz)
				if err != nil {
					fmt.Printf("could not resolve data - %s\n", addr.String())
					return nil
				}

				if len(res) == 0 {
					return fmt.Errorf("No validator found with address %s", addr)
				}

				var out types.Validator
				cdc.MustUnmarshalJSON(res, &out)

				return cliCtx.PrintOutput(out)
			}
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}
