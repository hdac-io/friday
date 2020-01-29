package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"

	"github.com/hdac-io/friday/x/nickname/types"
)

// GetDataQueryCmd controls GET type CLI controller
func GetDataQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	nameserverGetDataQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Nickname query commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	nameserverGetDataQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryAddress(cdc),
	)...)
	return nameserverGetDataQueryCmd
}

// GetCmdQueryAddress handles to get accounts list
func GetCmdQueryAddress(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-address [nickname]",
		Short: "Get address of given nickname",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			queryData := types.QueryReqUnitAccount{
				Nickname: args[0],
			}
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getaddress", types.ModuleName), bz)
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
