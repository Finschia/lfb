package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/line/lfb-sdk/baseapp"
	"github.com/line/lfb-sdk/client"
	"github.com/line/lfb-sdk/client/debug"
	"github.com/line/lfb-sdk/client/flags"
	"github.com/line/lfb-sdk/client/keys"
	"github.com/line/lfb-sdk/client/rpc"
	"github.com/line/lfb-sdk/codec"
	"github.com/line/lfb-sdk/server"
	servertypes "github.com/line/lfb-sdk/server/types"
	"github.com/line/lfb-sdk/snapshots"
	"github.com/line/lfb-sdk/store"
	sdk "github.com/line/lfb-sdk/types"
	authclient "github.com/line/lfb-sdk/x/auth/client"
	authcmd "github.com/line/lfb-sdk/x/auth/client/cli"
	"github.com/line/lfb-sdk/x/auth/types"
	vestingcli "github.com/line/lfb-sdk/x/auth/vesting/client/cli"
	banktypes "github.com/line/lfb-sdk/x/bank/types"
	"github.com/line/lfb-sdk/x/crisis"
	genutilcli "github.com/line/lfb-sdk/x/genutil/client/cli"
	lfbtypes "github.com/line/lfb/types"
	ostcli "github.com/line/ostracon/libs/cli"
	"github.com/line/ostracon/libs/log"
	dbm "github.com/line/tm-db/v2"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/line/lfb/app"
	"github.com/line/lfb/app/params"
)

const (
	flagTestnet = "testnet"
)

// NewRootCmd creates a new root command for simd. It is called once in the
// main function.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	encodingConfig := app.MakeEncodingConfig()
	initClientCtx := client.Context{}.
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(app.DefaultNodeHome)

	rootCmd := &cobra.Command{
		Use:   "lfb",
		Short: "LINE Financial Blockchain (LFB) App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}
			return lfbPreRunE(cmd)
		},
	}
	rootCmd.PersistentFlags().Bool(flagTestnet, false, "Run with testnet mode. The address prefix becomes tlink if this flag is set.")

	initRootCmd(rootCmd, encodingConfig)

	return rootCmd, encodingConfig
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig) {
	authclient.Codec = encodingConfig.Marshaler

	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		ostcli.NewCompletionCmd(rootCmd, true),
		testnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, createSimappAndExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(app.DefaultNodeHome),
	)
}
func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	app.ModuleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		flags.LineBreak,
		vestingcli.GetTxCmd(),
	)

	app.ModuleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func NewApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	return newApp(logger, db, traceStore, appOpts)
}

// newApp is an AppCreator
func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	ibCacheMetricsProvider, iavlCacheMetricsProvider :=
		baseapp.MetricsProvider(cast.ToBool(viper.GetBool(server.FlagPrometheus)))
	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager(
			cast.ToInt(appOpts.Get(server.FlagInterBlockCacheSize)), ibCacheMetricsProvider)
	}

	bech32CacheSize := cast.ToInt(appOpts.Get(server.FlagBech32CacheSize))
	if bech32CacheSize > 0 {
		sdk.SetBech32Cache(int64(bech32CacheSize))
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		panic(err)
	}
	snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	return app.NewLinkApp(
		logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		app.MakeEncodingConfig(), // Ideally, we would reuse the one created by NewRootCmd.
		appOpts,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetIAVLCacheManager(
			cast.ToInt(appOpts.Get(server.FlagIAVLCacheSize)), iavlCacheMetricsProvider),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshotStore(snapshotStore),
		baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval))),
		baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent))),
	)
}

func createSimappAndExport(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string,
	appOpts servertypes.AppOptions) (servertypes.ExportedApp, error) {
	encCfg := app.MakeEncodingConfig() // Ideally, we would reuse the one created by NewRootCmd.
	encCfg.Marshaler = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	var linkApp *app.LinkApp
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}
	if height != -1 {
		linkApp = app.NewLinkApp(logger, db, traceStore, false, map[int64]bool{}, homePath, cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)), encCfg, appOpts)

		if err := linkApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		linkApp = app.NewLinkApp(logger, db, traceStore, true, map[int64]bool{}, homePath, cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)), encCfg, appOpts)
	}

	return linkApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
}

func initConfig(testnet bool) {
	config := sdk.GetConfig()
	config.SetCoinType(lfbtypes.CoinType)
	config.SetFullFundraiserPath(lfbtypes.FullFundraiserPath)
	config.SetBech32PrefixForAccount(lfbtypes.Bech32PrefixAcc(testnet), lfbtypes.Bech32PrefixAccPub(testnet))
	config.SetBech32PrefixForConsensusNode(lfbtypes.Bech32PrefixConsAddr(testnet), lfbtypes.Bech32PrefixConsPub(testnet))
	config.SetBech32PrefixForValidator(lfbtypes.Bech32PrefixValAddr(testnet), lfbtypes.Bech32PrefixValPub(testnet))
	config.GetCoinType()
	config.Seal()
}

func lfbPreRunE(cmd *cobra.Command) (err error) {
	err = server.InterceptConfigsPreRunHandler(cmd)

	testnet := viper.GetBool(flagTestnet) // this should be called after initializing cmd
	initConfig(testnet)
	ctx := server.GetServerContextFromCmd(cmd)
	if cmd.Name() == server.StartCmd(nil, "").Name() {
		var networkMode string
		if testnet {
			networkMode = "testnet"
		} else {
			networkMode = "mainnet"
		}
		ctx.Logger.Info(fmt.Sprintf("Network mode is %s", networkMode))
	}
	return
}
