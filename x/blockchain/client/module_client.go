package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	blockchaincmd "github.com/stanvoets/blockchain/x/blockchain/client/cli"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)


// Exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc *amino.Codec
}


func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}


// Returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	// Group blockchain queries under a subcommand
	namesvcQueryCmd := &cobra.Command{
		Use:   "blockchain",
		Short: "Querying commands for the blockchain module",
	}

	namesvcQueryCmd.AddCommand(client.GetCommands(
		blockchaincmd.GetCmdResolveName(mc.storeKey, mc.cdc),
		blockchaincmd.GetCmdWhois(mc.storeKey, mc.cdc),
	)...)

	return namesvcQueryCmd
}


// Returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	namesvcTxCmd := &cobra.Command{
		Use:   "blockchain",
		Short: "Blockchain transactions subcommands",
	}

	namesvcTxCmd.AddCommand(client.PostCommands(
		blockchaincmd.GetCmdBuyName(mc.cdc),
		blockchaincmd.GetCmdSetName(mc.cdc),
	)...)

	return namesvcTxCmd
}

