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
		Use:   "getbalance [--wallet, --address, or --nickname] [--blockhash]",
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
					return fmt.Errorf("Malformed address in --address: %s\n%s", straddr, err.Error())
				}
			} else if nickname := viper.GetString(FlagNickname); nickname != "" {
				addr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, nickname)
				if err != nil {
					return fmt.Errorf("No registered address of the given nickname '%s'", nickname)
				}
			} else {
				return fmt.Errorf("One of --address, --wallet, --nickname is essential")
			}
			cliCtx = cliCtx.WithFromAddress(addr)

			var out types.QueryExecutionLayerResp
			if blockhashstr := viper.GetString(FlagBlockHash); blockhashstr != "" {
				blockHash, err := hex.DecodeString(blockhashstr)
				if err != nil || len(blockHash) != 32 {
					fmt.Println("Malformed block hash - ", blockhashstr)
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
					fmt.Printf("No balance data of input")
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
					fmt.Printf("No balance data of input")
					fmt.Println(err.Error())
					return nil
				}
				cdc.MustUnmarshalJSON(res, &out)
			}

			return cliCtx.PrintOutput(out)
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "flag for custom local path of client's home dir")
	cmd.Flags().String(FlagAddress, "", "flag for address")
	cmd.Flags().String(FlagWallet, "", "flag for wallet alias")
	cmd.Flags().String(FlagNickname, "", "flag for nickname")
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
