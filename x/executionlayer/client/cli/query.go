package cli

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
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

			queryData := types.QueryGetBalanceDetail{
				Address: addr,
			}
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querybalancedetail", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("no balance data of input")
				return nil
			}
			out := &state.Value{}
			err = jsonpb.Unmarshal(bytes.NewReader(res), out)
			if err != nil {
				fmt.Printf("Faild to json unmarshal, %s", err)
			}

			balance := string(cliutil.ToHdac(cliutil.Bigsun(out.GetStringValue())))

			_, err = fmt.Println(balance)
			return err
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

// GetCmdQuery is a EE query getter
func GetCmdQuery(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query address|uref|hash|local <data> <path> [--blockhash <blockhash_since>]",
		Short: "Get query of the data",
		Args:  cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			dataType := args[0]
			data := args[1]
			path := ""
			if len(args) == 3 {
				path = args[2]
			}

			queryData := types.QueryExecutionLayerDetail{
				KeyType: dataType,
				KeyData: data,
				Path:    path,
			}
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querydetail", types.ModuleName), bz)

			if err != nil {
				fmt.Printf("could not resolve data - %s %s %s\nerr : %s\n", dataType, data, path, err.Error())
				return nil
			}

			value := &state.Value{}
			err = jsonpb.Unmarshal(bytes.NewReader(res), value)

			marshaler := jsonpb.Marshaler{Indent: "  "}
			valueStr, err := marshaler.MarshalToString(value)
			if err != nil {
				fmt.Printf("could not resolve data - %s %s %s\nerr : %s\n", dataType, data, path, err.Error())
				return nil
			}

			valueStr = cliutil.ReplaceBase64HashToBech32(path, valueStr)

			_, err = fmt.Println(valueStr)
			return err
		},
	}
	return cmd
}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [--from <from>]",
		Short: "Query a validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			var addr sdk.AccAddress
			var err error
			if valueFromFromFlag != "" {
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

// GetCmdQueryDelegator implements the validator query command.
func GetCmdQueryDelegator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator [<vaidator-address>] [--from <from>]",
		Short: "Query a validator",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			var addr sdk.AccAddress
			var err error
			if valueFromFromFlag != "" {
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
			}

			var validator sdk.AccAddress
			if len(args) > 0 {
				validator, err = cliutil.GetAddress(cdc, cliCtx, args[0])
				if err != nil {
					return err
				}
			}

			if addr.Empty() && validator.Empty() {
				return fmt.Errorf("Requires validator or delegate address.")
			}

			queryData := types.NewQueryDelegatorParams(addr, validator)
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querydelegator", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("could not resolve data - %s\n", addr.String())
				return nil
			}

			if len(res) == 0 {
				errStr := "No delegator found with"
				if !addr.Empty() {
					errStr += (" address " + valueFromFromFlag)
				}
				if !validator.Empty() {
					if !addr.Empty() {
						errStr += " and "
					}
					errStr += ("validator " + args[0])
				}

				return fmt.Errorf(errStr)
			}

			var out types.Delegators
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

// GetCmdQueryVoter implements the validator query command.
func GetCmdQueryVoter(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voter [<contract_address>] [--from <from>]",
		Short: "Query a voter",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			var addr sdk.AccAddress
			var err error
			if valueFromFromFlag != "" {
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
			}

			var bz []byte
			var contractAddressUref types.ContractUrefAddress
			var contractAddressHash types.ContractHashAddress
			var contractAddress types.ContractAddress
			if len(args) > 0 {
				if strings.HasPrefix(args[0], sdk.Bech32PrefixContractURef) {
					contractAddressUref, err = sdk.ContractUrefAddressFromBech32(args[0])
					queryData := types.NewQueryVoterUrefParams(addr, contractAddressUref)
					bz = cdc.MustMarshalJSON(queryData)
					contractAddress = contractAddressUref
				} else if strings.HasPrefix(args[0], sdk.Bech32PrefixContractHash) {
					contractAddressHash, err = sdk.ContractHashAddressFromBech32(args[0])
					queryData := types.NewQueryVoterHashParams(addr, contractAddressHash)
					bz = cdc.MustMarshalJSON(queryData)
					contractAddress = contractAddressHash
				} else {
					err = fmt.Errorf("malformed contract address")
				}

				if err != nil {
					return err
				}
			} else {
				contractAddressUref = types.ContractUrefAddress{}
				queryData := types.NewQueryVoterUrefParams(addr, contractAddressUref)
				bz = cdc.MustMarshalJSON(queryData)
				contractAddress = contractAddressHash
			}

			if addr.Empty() && contractAddress.Empty() {
				return fmt.Errorf("requires voter address or dapp hash")
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/queryvoter", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("could not resolve data - %s %s\n", contractAddress.String(), addr.String())
				return nil
			}

			if len(res) == 0 {
				errStr := "No voter found with"
				if !addr.Empty() {
					errStr += (" address " + valueFromFromFlag)
				}
				if !contractAddress.Empty() {
					if !addr.Empty() {
						errStr += " and "
					}
					errStr += ("hash " + args[0])
				}

				return fmt.Errorf(errStr)
			}

			var out types.Voters
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

// GetCmdQueryReward is a getter of the reward of the address
func GetCmdQueryReward(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getreward --from <from>",
		Short: "Get reward of address",
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

			queryData := types.NewQueryGetReward(addr)
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/queryreward", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("no reward data of input")
				return nil
			}
			out := &state.Value{}
			err = jsonpb.Unmarshal(bytes.NewReader(res), out)
			if err != nil {
				fmt.Printf("Faild to json unmarshal, %s", err)
			}

			balance := string(cliutil.ToHdac(cliutil.Bigsun(out.GetStringValue())))

			_, err = fmt.Println(balance)
			return err
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

// GetCmdQueryCommission is a getter of the commission of the address
func GetCmdQueryCommission(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getcommission --from <from>",
		Short: "Get reward of address",
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

			queryData := types.NewQueryGetCommission(addr)
			bz := cdc.MustMarshalJSON(queryData)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/querycommission", types.ModuleName), bz)
			if err != nil {
				fmt.Printf("no reward data of input")
				return nil
			}
			out := &state.Value{}
			err = jsonpb.Unmarshal(bytes.NewReader(res), out)
			if err != nil {
				fmt.Printf("Faild to json unmarshal, %s", err)
			}

			balance := string(cliutil.ToHdac(cliutil.Bigsun(out.GetStringValue())))

			_, err = fmt.Println(balance)
			return err
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}
