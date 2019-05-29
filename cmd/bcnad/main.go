package main

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stanvoets/blockchain/app"
	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main(){
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	config.Seal()

	// Get context
	ctx := server.NewDefaultContext()

	// Let Cobra sort commands
	cobra.EnableCommandSorting = false

	// Construct root command
	rootCmd := &cobra.Command{
		Use:               "bcnad",
		Short:             "Bitcanna Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	// Add commands
	rootCmd.AddCommand(gaiaInit.InitCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.CollectGenTxsCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.TestnetFilesCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.GenTxCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.AddGenesisAccountCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.ValidateGenesisCmd(ctx, cdc))
	rootCmd.AddCommand(client.NewCompletionCmd(rootCmd, true))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// Prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "GA", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// Handle with #870
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewBcnaApp(
		logger, db, traceStore, true,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	if height != -1 {
		gApp := app.NewBcnaApp(logger, db, traceStore, false)
		err := gApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return gApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
	gApp := app.NewBcnaApp(logger, db, traceStore, true)
	return gApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
