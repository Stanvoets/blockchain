package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/common"
	"io"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stanvoets/blockchain/app"
)

const (
	flagInvCheckPeriod 	= "inv-check-period"
	flagOverwrite      	= "overwrite"
	flagClientHome   	= "home-client"
	flagVestingStart 	= "vesting-start-time"
	flagVestingEnd   	= "vesting-end-time"
	flagVestingAmt   	= "vesting-amount"
)

var invCheckPeriod uint

type printInfo struct {
	Moniker    string          `json:"moniker"`
	ChainID    string          `json:"chain_id"`
	NodeID     string          `json:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message"`
}

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
		Use:               "stand",
		Short:             "Bitcanna Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	// Add commands
	rootCmd.AddCommand(
		InitCmd(ctx, cdc),
		CollectGenTxsCmd(ctx, cdc),
		gaiaInit.TestnetFilesCmd(ctx, cdc),
		GenTxCmd(ctx, cdc),
		AddGenesisAccountCmd(ctx, cdc),
		gaiaInit.ValidateGenesisCmd(ctx, cdc),
		client.NewCompletionCmd(rootCmd, true),
	)

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// Prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "BC", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}

}

// Returns command that initializes all files needed for Tendermint
func InitCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			chainID := viper.GetString(client.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}

			nodeID, _, err := gaiaInit.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			var appState json.RawMessage
			genFile := config.GenesisFile()

			if appState, err = initializeEmptyGenesis(cdc, genFile, chainID, viper.GetBool("overwrite")); err != nil {
				return err
			}

			if err = gaiaInit.ExportGenesisFile(genFile, chainID, nil, appState); err != nil {
				return err
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)

			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			return displayInfo(cdc, toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(client.FlagChainID, "StanChain", "genesis file chain-id, if left blank will be randomly created")

	return cmd
}

func displayInfo(cdc *codec.Codec, info printInfo) error {
	out, err := codec.MarshalJSONIndent(cdc, info)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", string(out))
	if err != nil {
		panic(err)
	}

	return nil
}

func initializeEmptyGenesis(cdc *codec.Codec, genFile string, chainID string, overwrite bool, ) (appState json.RawMessage, err error) {

	if !overwrite && common.FileExists(genFile) {
		return nil, fmt.Errorf("genesis.json file already exists: %v", genFile)
	}

	return codec.MarshalJSONIndent(cdc, app.NewDefaultGenesisState())
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewStanApp(
		logger, db, traceStore, true, invCheckPeriod,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
	)
}
func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		bApp := app.NewStanApp(logger, db, traceStore, false, uint(1))
		err := bApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return bApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
	bApp := app.NewStanApp(logger, db, traceStore, true, uint(1))
	return bApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}