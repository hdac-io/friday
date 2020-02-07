package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	cliutil "github.com/hdac-io/friday/x/executionlayer/client/util"

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
		Use:   "getbalance --wallet|--address|--nickname <from> [--blockhash <blockhash_since>]",
		Short: "Get balance of address",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var addr sdk.AccAddress
			var err error

			// Extract "from" from flags
			if walletname := viper.GetString(FlagWallet); walletname != "" {
				kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
				if err != nil {
					return err
				}

				key, err := kb.Get(walletname)
				if err != nil {
					return err
				}

				addr = key.GetAddress()
			} else if straddr := viper.GetString(FlagAddress); straddr != "" {
				addr, err = sdk.AccAddressFromBech32(straddr)
				if err != nil {
					return fmt.Errorf("malformed address in --address: %s\n%s", straddr, err.Error())
				}
			} else if nickname := viper.GetString(FlagNickname); nickname != "" {
				addr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, nickname)
				if err != nil {
					return fmt.Errorf("no registered address of the given nickname '%s'", nickname)
				}
			} else {
				return fmt.Errorf("one of --address, --wallet, --nickname is essential")
			}
			cliCtx = cliCtx.WithFromAddress(addr)

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

			return cliCtx.PrintOutput(out)
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(FlagAddress, "", "Bech32 endocded address (fridayxxxxxx..)")
	cmd.Flags().String(FlagWallet, "", "Wallet alias in local")
	cmd.Flags().String(FlagNickname, "", "Nickname (Readable ID)")
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
		Use:   "validator --wallet|--address|--nickname <from>",
		Short: "Query a validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var addr sdk.AccAddress
			var err error

			// Extract "from" from flags
			if walletname := viper.GetString(FlagWallet); walletname != "" {
				kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
				if err != nil {
					return err
				}

				key, err := kb.Get(walletname)
				if err != nil {
					return err
				}

				addr = key.GetAddress()
			} else if straddr := viper.GetString(FlagAddress); straddr != "" {
				addr, err = sdk.AccAddressFromBech32(straddr)
				if err != nil {
					return fmt.Errorf("malformed address in --address: %s\n%s", straddr, err.Error())
				}
			} else if nickname := viper.GetString(FlagNickname); nickname != "" {
				addr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, nickname)
				if err != nil {
					return fmt.Errorf("no registered address of the given nickname '%s'", nickname)
				}
			} else {
				return fmt.Errorf("one of --address, --wallet, --nickname is essential")
			}

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
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(FlagAddress, "", "Bech32 endocded address (fridayxxxxxx..)")
	cmd.Flags().String(FlagWallet, "", "Wallet alias in local")
	cmd.Flags().String(FlagNickname, "", "Nickname (Readable ID)")

	return cmd
}

// // GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/queryallvalidator", types.ModuleName))
			if err != nil {
				fmt.Printf("could not resolve")
				return nil
			}

			var out types.Validators
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}
