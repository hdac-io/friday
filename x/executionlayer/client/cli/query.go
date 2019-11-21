package cli

import (
	"fmt"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/x/executionlayer/types"

	"github.com/spf13/cobra"
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
		GetCmdQueryBalanceWithBlockHash(cdc),
		GetCmdQuery(cdc),
		GetCmdQueryWithHash(cdc),
	)...)
	return executionlayerQueryCmd
}

// GetCmdQueryBalance is a getter of the balance of the address
func GetCmdQueryBalance(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "getbalance [address]",
		Short: "Get balance of address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			name := args[0]

			queryData := types.QueryGetBalance{
				Address: name,
			}
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalance", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("could not resolve address - %s \n", name)
				return nil
			}

			var out types.QueryExecutionLayerResp
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryBalanceWithBlockHash is a getter of the balance of the address
func GetCmdQueryBalanceWithBlockHash(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "getbalancewithhash [address] [block_hash]",
		Short: "Get balance of address with block hash",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			name := args[0]
			blockHash := []byte(args[1])

			queryData := types.QueryGetBalanceDetail{
				Address:   name,
				StateHash: blockHash,
			}
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalancedetail", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("could not resolve address - %s \n", name)
				return nil
			}

			var out types.QueryExecutionLayerResp
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQuery is a EE query getter
func GetCmdQuery(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "query [type:=address,uref,hash,local] [data] [path]",
		Short: "Get query of the data",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			dataType := args[0]
			data := []byte(args[1])
			path := args[2]

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

			var out types.QueryExecutionLayerResp
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryWithHash is a EE query getter with block hash
func GetCmdQueryWithHash(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "querywithhash [type:=address,uref,hash,local] [data] [path] [block_hash]",
		Short: "Get query of the data with block hash",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			dataType := args[0]
			data := []byte(args[1])
			path := args[2]
			blockHash := []byte(args[3])

			queryData := types.QueryExecutionLayerDetail{
				KeyType:   dataType,
				KeyData:   data,
				Path:      path,
				StateHash: blockHash,
			}
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querydetail", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("could not resolve data - %s %s %s\n", dataType, data, path)
				return nil
			}

			var out types.QueryExecutionLayerResp
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
