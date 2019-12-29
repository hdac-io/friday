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
		Short:                      "Querying commands for the nameserver",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	nameserverGetDataQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryUnitAccount(storeKey, cdc),
	)...)
	return nameserverGetDataQueryCmd
}

// GetCmdQueryUnitAccount handles to get accounts list
func GetCmdQueryUnitAccount(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "getaccount [ID]",
		Short: "Get account information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			name := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaccount/%s", queryRoute, name), nil)
			if err != nil {
				fmt.Printf("could not resolve account - %s \n", name)
				return nil
			}

			var out types.QueryResUnitAccount
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
