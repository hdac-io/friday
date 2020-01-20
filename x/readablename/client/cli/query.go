package cli

import (
	"fmt"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/x/readablename/types"

	"github.com/spf13/cobra"
)

// GetDataQueryCmd controls GET type CLI controller
func GetDataQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	nameserverGetDataQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Readable name service query commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	nameserverGetDataQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryUnitAccount(cdc),
	)...)
	return nameserverGetDataQueryCmd
}

// GetCmdQueryUnitAccount handles to get accounts list
func GetCmdQueryUnitAccount(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "getaccount [readable_name]",
		Short: "Get account information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			queryData := types.QueryReqUnitAccount{
				Name: args[0],
			}
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaccount", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("could not resolve account - %s \n", args[0])
				return nil
			}

			var out types.QueryResUnitAccount
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
