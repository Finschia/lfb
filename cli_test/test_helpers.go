package clitest

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/line/lfb/app"
	lfbcmd "github.com/line/lfb/cmd/lfb/cmd"
	lfbtypes "github.com/line/lfb/types"

	"github.com/line/lfb-sdk/baseapp"
	"github.com/line/lfb-sdk/client"
	clientkeys "github.com/line/lfb-sdk/client/keys"
	"github.com/line/lfb-sdk/client/rpc"
	"github.com/line/lfb-sdk/codec/legacy"
	"github.com/line/lfb-sdk/crypto/hd"
	"github.com/line/lfb-sdk/crypto/keyring"
	"github.com/line/lfb-sdk/server"
	srvconfig "github.com/line/lfb-sdk/server/config"
	servertypes "github.com/line/lfb-sdk/server/types"
	storetypes "github.com/line/lfb-sdk/store/types"
	"github.com/line/lfb-sdk/testutil"
	testcli "github.com/line/lfb-sdk/testutil/cli"
	testnet "github.com/line/lfb-sdk/testutil/network"
	sdk "github.com/line/lfb-sdk/types"
	"github.com/line/lfb-sdk/types/tx"
	authcli "github.com/line/lfb-sdk/x/auth/client/cli"
	authtypes "github.com/line/lfb-sdk/x/auth/types"
	bankcli "github.com/line/lfb-sdk/x/bank/client/cli"
	banktypes "github.com/line/lfb-sdk/x/bank/types"
	distcli "github.com/line/lfb-sdk/x/distribution/client/cli"
	disttypes "github.com/line/lfb-sdk/x/distribution/types"
	"github.com/line/lfb-sdk/x/genutil"
	genutilcli "github.com/line/lfb-sdk/x/genutil/client/cli"
	govcli "github.com/line/lfb-sdk/x/gov/client/cli"
	gov "github.com/line/lfb-sdk/x/gov/types"
	slashingcli "github.com/line/lfb-sdk/x/slashing/client/cli"
	slashing "github.com/line/lfb-sdk/x/slashing/types"
	stakingcli "github.com/line/lfb-sdk/x/staking/client/cli"
	staking "github.com/line/lfb-sdk/x/staking/types"
	wasmcli "github.com/line/lfb-sdk/x/wasm/client/cli"

	wasmtypes "github.com/line/lfb-sdk/x/wasm/types"
	ostcfg "github.com/line/ostracon/config"
	ostflags "github.com/line/ostracon/libs/cli/flags"
	"github.com/line/ostracon/libs/log"
	osthttp "github.com/line/ostracon/rpc/client/http"
	ostctypes "github.com/line/ostracon/rpc/core/types"
	osttypes "github.com/line/ostracon/types"
)

const (
	denom        = "stake"
	keyFoo       = "foo"
	keyBar       = "bar"
	fooDenom     = "foot"
	feeDenom     = "feet"
	fee2Denom    = "fee2t"
	keyBaz       = "baz"
	keyVesting   = "vesting"
	keyFooBarBaz = "foobarbaz"

	DenomStake = "stake2"
	DenomLink  = "link"
	UserTina   = "tina"
	UserKevin  = "kevin"
	UserRinah  = "rinah"
	UserBrian  = "brian"
	UserEvelyn = "evelyn"
	UserSam    = "sam"
)

const (
	namePrefix        = "node"
	networkNamePrefix = "line-lfbnode-testnet-"
)

var curPort int32 = 26656

var (
	TotalCoins = sdk.NewCoins(
		sdk.NewCoin(DenomLink, sdk.TokensFromConsensusPower(6000)),
		sdk.NewCoin(DenomStake, sdk.TokensFromConsensusPower(600000000)),
		sdk.NewCoin(fee2Denom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(fooDenom, sdk.TokensFromConsensusPower(2000)),
		sdk.NewCoin(denom, sdk.TokensFromConsensusPower(300)), // We don't use inflation
		// sdk.NewCoin(denom, sdk.TokensFromConsensusPower(300).Add(sdk.NewInt(12))), // add coins from inflation
	)

	startCoins = sdk.NewCoins(
		sdk.NewCoin(fee2Denom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(fooDenom, sdk.TokensFromConsensusPower(1000)),
		sdk.NewCoin(denom, sdk.TokensFromConsensusPower(150)),
	)

	vestingCoins = sdk.NewCoins(
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(500000)),
	)

	// coins we set during ./.initialize.sh
	defaultCoins = sdk.NewCoins(
		sdk.NewCoin(DenomLink, sdk.TokensFromConsensusPower(1000)),
		sdk.NewCoin(DenomStake, sdk.TokensFromConsensusPower(100000000)),
	)

	ostraconCmd = &cobra.Command{
		Use:   "ostracon",
		Short: "Ostracon subcommands",
	}
)

func init() {
	testnet := false
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(lfbtypes.Bech32PrefixAcc(testnet), lfbtypes.Bech32PrefixAccPub(testnet))
	config.SetBech32PrefixForValidator(lfbtypes.Bech32PrefixValAddr(testnet), lfbtypes.Bech32PrefixValPub(testnet))
	config.SetBech32PrefixForConsensusNode(lfbtypes.Bech32PrefixConsAddr(testnet), lfbtypes.Bech32PrefixConsPub(testnet))
	config.SetCoinType(lfbtypes.CoinType)
	config.SetFullFundraiserPath(lfbtypes.FullFundraiserPath)
	config.Seal()

	ostraconCmd.AddCommand(
		server.ShowNodeIDCmd(),
		server.ShowValidatorCmd(),
		server.ShowAddressCmd(),
		server.VersionCmd(),
	)
}

// ___________________________________________________________________________________
// Fixtures

// Fixtures is used to setup the testing environment
type Fixtures struct {
	ChainID  string
	RPCAddr  string
	Port     string
	Home     string
	P2PAddr  string
	P2PPort  string
	GRPCAddr string
	GRPCPort string
	TMAddr   string
	TMPort   string
	Moniker  string
	T        *testing.T
}

func getHomeDir(t *testing.T) string {
	tmpDir := path.Join(os.ExpandEnv("$HOME"), ".lfbtest")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(t, err)
	tmpDir, err = ioutil.TempDir(tmpDir, "link_integration_"+strings.Split(t.Name(), "/")[0]+"_")
	require.NoError(t, err)
	return tmpDir
}

// NewFixtures creates a new instance of Fixtures with many vars set
func NewFixtures(t *testing.T, homeDir string) *Fixtures {
	if err := os.MkdirAll(filepath.Join(homeDir, "config/"), os.ModePerm); err != nil {
		panic(err)
	}

	servAddr, servPort := newTCPAddr(t)
	p2pAddr, p2pPort := newTCPAddr(t)
	grpcAddr, grpcPort := newGRPCAddr(t)

	return &Fixtures{
		T:        t,
		Home:     homeDir,
		RPCAddr:  servAddr,
		P2PAddr:  p2pAddr,
		Port:     servPort,
		P2PPort:  p2pPort,
		GRPCAddr: grpcAddr,
		GRPCPort: grpcPort,
		Moniker:  "", // initialized by LFBInit
	}
}

func newTCPAddr(t *testing.T) (addr, port string) {
	portI := atomic.AddInt32(&curPort, 1)
	require.Less(t, portI, int32(32768), "A new port should be less than ip_local_port_range.min")

	port = fmt.Sprintf("%d", portI)
	addr = fmt.Sprintf("tcp://0.0.0.0:%s", port)
	return
}

func newGRPCAddr(t *testing.T) (addr, port string) {
	portI := atomic.AddInt32(&curPort, 1)
	require.Less(t, portI, int32(32768), "A new port should be less than ip_local_port_range.min")

	port = fmt.Sprintf("%d", portI)
	addr = fmt.Sprintf("0.0.0.0:%s", port)
	return
}

func (f *Fixtures) LogResult(isSuccess bool, stdOut, stdErr string) {
	if !isSuccess {
		f.T.Error(stdErr)
	} else {
		f.T.Log(stdOut)
	}
}

func (f Fixtures) Clone() *Fixtures {
	newF := NewFixtures(f.T, getHomeDir(f.T))
	newF.ChainID = f.ChainID

	if err := copyDirectory(f.Home, newF.Home); err != nil {
		os.Exit(0)
	}

	return newF
}

func copyDirectory(src, dest string) error {
	cmd := exec.Command("cp", "-r", src, dest)
	return cmd.Start()
}

// GenesisFile returns the path of the genesis file
func (f Fixtures) GenesisFile() string {
	return filepath.Join(f.Home, "config", "genesis.json")
}

func (f Fixtures) PrivValidatorKeyFile() string {
	return filepath.Join(f.Home, "config", "priv_validator_key.json")
}

// GenesisFile returns the application's genesis state
func (f Fixtures) GenesisState() app.GenesisState {
	genDoc, err := osttypes.GenesisDocFromFile(f.GenesisFile())
	require.NoError(f.T, err)

	var appState app.GenesisState

	require.NoError(f.T, legacy.Cdc.UnmarshalJSON(genDoc.AppState, &appState))
	return appState
}

// InitFixtures is called at the beginning of a test  and initializes a chain
// with 1 validator.
func InitFixtures(t *testing.T) (f *Fixtures) {
	f = NewFixtures(t, getHomeDir(t))

	// ensure keystore has foo and bar keys
	f.KeysDelete(keyFoo)
	f.KeysDelete(keyBar)
	f.KeysDelete(keyBaz)
	f.KeysDelete(keyVesting)
	f.KeysDelete(keyFooBarBaz)
	f.KeysAdd(keyFoo)
	f.KeysAdd(keyBar)
	f.KeysAdd(keyBaz)
	f.KeysAdd(keyVesting)
	f.KeysAdd(keyFooBarBaz, "--multisig-threshold=2", fmt.Sprintf(
		"--multisig=%s,%s,%s", keyFoo, keyBar, keyBaz))

	// ensure keystore to have user keys
	f.KeysDelete(UserTina)
	f.KeysDelete(UserKevin)
	f.KeysDelete(UserRinah)
	f.KeysDelete(UserBrian)
	f.KeysDelete(UserEvelyn)
	f.KeysDelete(UserSam)
	f.KeysAdd(UserTina)
	f.KeysAdd(UserKevin)
	f.KeysAdd(UserRinah)
	f.KeysAdd(UserBrian)
	f.KeysAdd(UserEvelyn)
	f.KeysAdd(UserSam)

	// NOTE: LFBInit sets the ChainID
	f.LFBInit(keyFoo)

	// start an account with tokens
	f.AddGenesisAccount(f.KeyAddress(keyFoo), startCoins)
	// f.AddGenesisAccount(f.KeyAddress(keyBar), startCoins)
	f.AddGenesisAccount(
		f.KeyAddress(keyVesting), startCoins,
		fmt.Sprintf("--vesting-amount=%s", vestingCoins),
		fmt.Sprintf("--vesting-start-time=%d", time.Now().UTC().UnixNano()),
		fmt.Sprintf("--vesting-end-time=%d", time.Now().Add(60*time.Second).UTC().UnixNano()),
	)

	// add genesis accounts for testing
	f.AddGenesisAccount(f.KeyAddress(UserTina), defaultCoins)
	f.AddGenesisAccount(f.KeyAddress(UserKevin), defaultCoins)
	f.AddGenesisAccount(f.KeyAddress(UserRinah), defaultCoins)
	f.AddGenesisAccount(f.KeyAddress(UserBrian), defaultCoins)
	f.AddGenesisAccount(f.KeyAddress(UserEvelyn), defaultCoins)
	f.AddGenesisAccount(f.KeyAddress(UserSam), defaultCoins)

	f.GenTx(keyFoo)
	f.CollectGenTxs()

	return f
}

// Cleanup is meant to be run at the end of a test to clean up an remaining test state
func (f *Fixtures) Cleanup(dirs ...string) {
	clean := append(dirs, f.Home) //nolint:gocritic
	for _, d := range clean {
		_ = os.RemoveAll(d)
	}
}

func getCliCtx(f *Fixtures) client.Context {
	kb, err := keyring.New(sdk.KeyringServiceName(), "test", f.Home, os.Stdin)
	require.NoError(f.T, err)

	httpClient, err := osthttp.New(f.RPCAddr, "/websocket")
	require.NoError(f.T, err)
	encodingConfig := app.MakeEncodingConfig()
	return client.Context{}.
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode("block").
		WithSkipConfirmation(true).
		WithHomeDir(f.Home).
		WithKeyringDir(f.Home).
		WithKeyring(kb).
		WithChainID(f.ChainID).
		WithNodeURI(f.RPCAddr).
		WithClient(httpClient)
}

// ___________________________________________________________________________________
// lfb

// UnsafeResetAll is lfb unsafe-reset-all
func (f *Fixtures) UnsafeResetAll(flags ...string) {
	cmd := server.UnsafeResetAllCmd()
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags("", flags...))
	require.NoError(f.T, err)

	err = os.RemoveAll(filepath.Join(f.Home, "config", "gentx"))
	require.NoError(f.T, err)
}

// LFBInit is lfb init
// NOTE: LFBInit sets the ChainID for the Fixtures instance
func (f *Fixtures) LFBInit(moniker string, flags ...string) {
	f.Moniker = moniker
	args := fmt.Sprintf("-o %s", moniker)
	cmd := genutilcli.InitCmd(app.ModuleBasics, f.Home)
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)

	genesisPath := filepath.Join(f.Home, "config", "genesis.json")
	genDoc, err := osttypes.GenesisDocFromFile(genesisPath)
	require.NoError(f.T, err)

	f.ChainID = genDoc.ChainID
}

// AddGenesisAccount is lfb add-genesis-account
func (f *Fixtures) AddGenesisAccount(address sdk.AccAddress, coins sdk.Coins, flags ...string) {
	args := fmt.Sprintf("%s %s --keyring-backend=test", address, coins)
	cmd := lfbcmd.AddGenesisAccountCmd(f.Home)
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
}

// GenTx is lfb gentx
func (f *Fixtures) GenTx(name string, flags ...string) {
	args := fmt.Sprintf("%s 100000000stake --chain-id=%s", name, f.ChainID)
	encodingConfig := app.MakeEncodingConfig()
	cmd := genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, f.Home)
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
}

// CollectGenTxs is lfb collect-gentxs
func (f *Fixtures) CollectGenTxs(flags ...string) {
	cmd := genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, f.Home)
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags("", flags...))
	require.NoError(f.T, err)
}

func (f *Fixtures) LFBStart(minGasPrices string) *testnet.Network {
	cfg := newTestnetConfig(f.T, f.GenesisState(), f.ChainID, minGasPrices)
	n := testnet.NewWithoutInit(f.T, cfg, f.Home, []*testnet.Validator{newValidator(f, cfg, srvconfig.DefaultConfig(), server.NewDefaultContext())})
	err := n.WaitForNextBlock()
	require.NoError(f.T, err)
	return n
}

// LFBOstracon returns the results of lfb ostracon [query]
func (f *Fixtures) LFBOstracon(query string) string {
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), ostraconCmd, strings.Split(query, " "))
	require.NoError(f.T, err)

	return out.String()
}

// ValidateGenesis runs lfb validate-genesis
func (f *Fixtures) ValidateGenesis(genFile string, flags ...string) {
	cmd := genutilcli.ValidateGenesisCmd(app.ModuleBasics)
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(genFile, flags...))
	require.NoError(f.T, err)
}

// ___________________________________________________________________________________
// lfb keys

// KeysDelete is lfb keys delete
func (f *Fixtures) KeysDelete(name string, flags ...string) {
	args := fmt.Sprintf("delete --keyring-backend=test -f -y %s", name)
	cmd := clientkeys.Commands(f.Home)
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))

	if strings.Contains(fmt.Sprintf("%v", err), "The specified item could not be found in the keyring") {
		return
	}
	require.NoError(f.T, err)
}

// KeysAdd is lfb keys add
func (f *Fixtures) KeysAdd(name string, flags ...string) {
	args := fmt.Sprintf("add --keyring-backend=test --output=json %s", name)
	cmd := clientkeys.Commands(f.Home)
	stdout, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, append(addFlags(args, flags...), flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, stdout)
}

// KeysAddRecover prepares lfb keys add --recover
func (f *Fixtures) KeysAddRecover(name, mnemonic string, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("add --keyring-backend=test --recover %s", name)
	cmd := clientkeys.Commands(f.Home)
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// KeysAddRecoverHDPath prepares lfb keys add --recover --account --index
func (f *Fixtures) KeysAddRecoverHDPath(name, mnemonic string, account uint32, index uint32, flags ...string) {
	args := fmt.Sprintf("add --keyring-backend=test --recover %s --account=%d --index=%d", name, account, index)
	cmd := clientkeys.Commands(f.Home)
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
}

// KeysShow is lfb keys show
func (f *Fixtures) KeysShow(name string, flags ...string) keyring.KeyOutput {
	args := fmt.Sprintf("show --keyring-backend=test --keyring-dir=%s --output json %s", f.Home, name)
	cmd := clientkeys.Commands(f.Home)
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var ko keyring.KeyOutput
	err = clientkeys.UnmarshalJSON(out.Bytes(), &ko)
	require.NoError(f.T, err)
	return ko
}

// KeyAddress returns the SDK account address from the key
func (f *Fixtures) KeyAddress(name string) sdk.AccAddress {
	ko := f.KeysShow(name)
	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(f.T, err)
	return accAddr
}

// ___________________________________________________________________________________
// lfb tx send/sign/broadcast

// TxSend is lfb tx send
func (f *Fixtures) TxSend(from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (testutil.BufferWriter, error) {
	node := f.RPCAddr
	for i, flag := range flags {
		if strings.Contains(flag, "node") {
			node = strings.Split(flag, "=")[1]
			flags = append(flags[:i], flags[i+1:]...)
		}
	}

	args := fmt.Sprintf("--keyring-backend=test %s %s %s --node=%s", from, to, amount, node)
	cmd := bankcli.NewSendTxCmd()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// TxSign is lfb tx sign
func (f *Fixtures) TxSign(signer, fileName string, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("--keyring-backend=test --from=%s --chain-id=%s %v --node=%s", signer, f.ChainID, fileName, f.RPCAddr)
	cmd := authcli.GetSignCommand()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// TxBroadcast is lfb tx broadcast
func (f *Fixtures) TxBroadcast(fileName string, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("%v --node=%s", fileName, f.RPCAddr)
	arr := addFlags(args, flags...)
	arr = append(arr, flags...)
	cmd := authcli.GetBroadcastCommand()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, arr)
}

// TxEncode is lfb tx encode
func (f *Fixtures) TxEncode(fileName string, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("%v --node=%s", fileName, f.RPCAddr)
	cmd := authcli.GetEncodeCommand()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// TxMultisign is lfb tx multisign
func (f *Fixtures) TxMultisign(fileName, name string, signaturesFiles []string,
	flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("--keyring-backend=test %s %s %s --node=%s", fileName, name, strings.Join(signaturesFiles, " "), f.RPCAddr)
	cmd := authcli.GetMultiSignCommand()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// ___________________________________________________________________________________
// lfb tx staking

// TxStakingCreateValidator is lfb tx staking create-validator
func (f *Fixtures) TxStakingCreateValidator(from, consPubKey string, amount sdk.Coin, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("--keyring-backend=test --from=%s --pubkey=%s", from, consPubKey)
	args += fmt.Sprintf(" --amount=%v --moniker=%v --commission-rate=%v", amount, from, "0.05")
	args += fmt.Sprintf(" --commission-max-rate=%v --commission-max-change-rate=%v", "0.20", "0.10")
	args += fmt.Sprintf(" --min-self-delegation=%v", "1")
	args += fmt.Sprintf(" --node=%s", f.RPCAddr)
	cmd := stakingcli.NewCreateValidatorCmd()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// TxStakingUnbond is lfb tx staking unbond
func (f *Fixtures) TxStakingUnbond(from, shares string, validator sdk.ValAddress, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("--keyring-backend=test %s %v --from=%s --node=%s", validator, shares, from, f.RPCAddr)
	cmd := stakingcli.NewUnbondCmd()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// ___________________________________________________________________________________
// lfb tx gov

// TxGovSubmitProposal is lfb tx gov submit-proposal
func (f *Fixtures) TxGovSubmitProposal(from, typ, title, description string, deposit sdk.Coin, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("--keyring-backend=test --from=%s --type=%s", from, typ)
	args += fmt.Sprintf(" --title=%s --description=%s --deposit=%s", title, description, deposit)
	args += fmt.Sprintf(" --node=%s", f.RPCAddr)
	cmd := govcli.NewCmdSubmitProposal()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// TxGovDeposit is lfb tx gov deposit
func (f *Fixtures) TxGovDeposit(proposalID int, from string, amount sdk.Coin, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("%d %s --keyring-backend=test --from=%s --node=%s", proposalID, amount, from, f.RPCAddr)
	cmd := govcli.NewCmdDeposit()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// TxGovVote is lfb tx gov vote
func (f *Fixtures) TxGovVote(proposalID int, option gov.VoteOption, from string, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("%d %s --keyring-backend=test --from=%s --node=%s", proposalID, option, from, f.RPCAddr)
	cmd := govcli.NewCmdVote()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// TxGovSubmitParamChangeProposal executes a CLI parameter change proposal
// submission.
func (f *Fixtures) TxGovSubmitParamChangeProposal(
	from, proposalPath string, flags ...string,
) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("param-change --proposal=%s --keyring-backend=test --from=%s --node=%s", proposalPath, from, f.RPCAddr)
	cmd := govcli.NewCmdSubmitProposal()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// TxGovSubmitCommunityPoolSpendProposal executes a CLI community pool spend proposal
// submission.
func (f *Fixtures) TxGovSubmitCommunityPoolSpendProposal(
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("%s --keyring-backend=test --from=%s --node=%s", proposalPath, from, f.RPCAddr)
	cmd := govcli.NewCmdSubmitProposal()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

// ___________________________________________________________________________________
// lfb query account

// QueryAccount is lfb query account
func (f *Fixtures) QueryAccount(address sdk.AccAddress, flags ...string) authtypes.BaseAccount {
	args := fmt.Sprintf("%s -o=json", address)
	cmd := authcli.GetAccountCmd()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)

	var acc authtypes.BaseAccount
	err = legacy.Cdc.UnmarshalJSON(out.Bytes(), &acc)
	require.NoError(f.T, err)
	return acc
}

// ___________________________________________________________________________________
// lfb query bank

// QueryBalances is lfb query bank balances
func (f *Fixtures) QueryBalances(address sdk.AccAddress, flags ...string) banktypes.QueryAllBalancesResponse {
	args := fmt.Sprintf("%s -o=json", address)
	cmd := bankcli.GetBalancesCmd()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)

	var bal banktypes.QueryAllBalancesResponse
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &bal)
	require.NoError(f.T, err)

	return bal
}

// ___________________________________________________________________________________
// lfb query tx

// QueryTx is lfb query tx
func (f *Fixtures) QueryTx(hash string, flags ...string) *sdk.TxResponse {
	args := fmt.Sprintf("%s -o=json --node=%s", hash, f.RPCAddr)
	cmd := authcli.QueryTxCmd()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var result sdk.TxResponse
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &result)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return &result
}

// QueryTxInvalid query tx with wrong hash and compare expected error
func (f *Fixtures) QueryTxInvalid(expectedErr error, hash string, flags ...string) {
	args := fmt.Sprintf("%s -o=json --node=%s", hash, f.RPCAddr)
	cmd := authcli.QueryTxCmd()
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	require.EqualError(f.T, expectedErr, errStr)
}

// ___________________________________________________________________________________
// lfb query txs

// QueryTxs is lfb query txs
func (f *Fixtures) QueryTxs(page, limit int, flags ...string) *sdk.SearchTxsResult {
	args := fmt.Sprintf("--page=%d --limit=%d --node=%s", page, limit, f.RPCAddr)
	cmd := authcli.QueryTxsByEventsCmd()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var result sdk.SearchTxsResult
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &result)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return &result
}

// QueryTxsInvalid query txs with wrong parameters and compare expected error
func (f *Fixtures) QueryTxsInvalid(expectedErr error, page, limit int, flags ...string) {
	args := fmt.Sprintf("--page=%d --limit=%d --node=%s", page, limit, f.RPCAddr)
	cmd := authcli.QueryTxsByEventsCmd()
	_, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	require.EqualError(f.T, expectedErr, errStr)
}

// ___________________________________________________________________________________
// lfb query block

func (f *Fixtures) QueryLatestBlock(flags ...string) *ostctypes.ResultBlock {
	args := fmt.Sprintf("--node=%s", f.RPCAddr)
	cmd := rpc.BlockCommand()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var result ostctypes.ResultBlock
	err = legacy.Cdc.UnmarshalJSON(out.Bytes(), &result)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return &result
}

func (f *Fixtures) QueryBlockWithHeight(height int, flags ...string) *ostctypes.ResultBlock {
	args := fmt.Sprintf("%d --node=%s", height, f.RPCAddr)
	cmd := rpc.BlockCommand()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var result ostctypes.ResultBlock
	err = legacy.Cdc.UnmarshalJSON(out.Bytes(), &result)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return &result
}

// ___________________________________________________________________________________
// lfb query staking

// QueryStakingValidator is lfb query staking validator
func (f *Fixtures) QueryStakingValidator(valAddr sdk.ValAddress, flags ...string) staking.Validator {
	args := fmt.Sprintf("%s --node=%s -o=json", valAddr, f.RPCAddr)
	cmd := stakingcli.GetCmdQueryValidator()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var validator staking.Validator
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &validator)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return validator
}

// QueryStakingUnbondingDelegationsFrom is lfb query staking unbonding-delegations-from
func (f *Fixtures) QueryStakingUnbondingDelegationsFrom(valAddr sdk.ValAddress, flags ...string) staking.QueryValidatorUnbondingDelegationsResponse {
	args := fmt.Sprintf("%s --node=%s -o=json", valAddr, f.RPCAddr)
	cmd := stakingcli.GetCmdQueryValidatorUnbondingDelegations()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var ubds staking.QueryValidatorUnbondingDelegationsResponse
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &ubds)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return ubds
}

// QueryStakingDelegationsTo is lfb query staking delegations-to
func (f *Fixtures) QueryStakingDelegationsTo(valAddr sdk.ValAddress, flags ...string) staking.QueryValidatorDelegationsResponse {
	args := fmt.Sprintf("%s --node=%s -o=json", valAddr, f.RPCAddr)
	cmd := stakingcli.GetCmdQueryValidatorDelegations()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var delegations staking.QueryValidatorDelegationsResponse
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &delegations)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return delegations
}

// QueryStakingPool is lfb query staking pool
func (f *Fixtures) QueryStakingPool(flags ...string) staking.Pool {
	args := fmt.Sprintf("--node=%s", f.RPCAddr)
	cmd := stakingcli.GetCmdQueryPool()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var pool staking.Pool
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &pool)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return pool
}

// QueryStakingParameters is lfb query staking parameters
func (f *Fixtures) QueryStakingParameters(flags ...string) staking.Params {
	args := fmt.Sprintf("--node=%s", f.RPCAddr)
	cmd := stakingcli.GetCmdQueryParams()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var params staking.Params
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &params)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return params
}

// ___________________________________________________________________________________
// lfb query gov

// QueryGovParamDeposit is lfb query gov param deposit
func (f *Fixtures) QueryGovParamDeposit(flags ...string) gov.DepositParams {
	args := fmt.Sprintf("deposit --node=%s -o=json", f.RPCAddr)
	cmd := govcli.GetCmdQueryParam()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var depositParam gov.DepositParams
	err = legacy.Cdc.UnmarshalJSON(out.Bytes(), &depositParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return depositParam
}

// QueryGovParamVoting is lfb query gov param voting
func (f *Fixtures) QueryGovParamVoting(flags ...string) gov.VotingParams {
	args := fmt.Sprintf("voting --node=%s -o=json", f.RPCAddr)
	cmd := govcli.GetCmdQueryParam()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var votingParam gov.VotingParams
	err = legacy.Cdc.UnmarshalJSON(out.Bytes(), &votingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votingParam
}

// QueryGovParamTallying is lfb query gov param tallying
func (f *Fixtures) QueryGovParamTallying(flags ...string) gov.TallyParams {
	args := fmt.Sprintf("tallying --node=%s -o=json", f.RPCAddr)
	cmd := govcli.GetCmdQueryParam()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var tallyingParam gov.TallyParams
	err = legacy.Cdc.UnmarshalJSON(out.Bytes(), &tallyingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return tallyingParam
}

// QueryGovProposals is lfb query gov proposals
func (f *Fixtures) QueryGovProposals(flags ...string) gov.QueryProposalsResponse {
	cmd := govcli.GetCmdQueryProposals()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags("-o=json", flags...))
	if err != nil && strings.Contains(err.Error(), "no proposals found") {
		return gov.QueryProposalsResponse{}
	}
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var proposals gov.QueryProposalsResponse
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &proposals)
	require.NoError(f.T, err)
	return proposals
}

// QueryGovProposal is lfb query gov proposal
func (f *Fixtures) QueryGovProposal(proposalID int, flags ...string) gov.Proposal {
	args := fmt.Sprintf("%d --node=%s -o=json", proposalID, f.RPCAddr)
	cmd := govcli.GetCmdQueryProposal()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var proposal gov.Proposal
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &proposal)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return proposal
}

// QueryGovVote is lfb query gov vote
func (f *Fixtures) QueryGovVote(proposalID int, voter sdk.AccAddress, flags ...string) gov.Vote {
	args := fmt.Sprintf("%d %s --node=%s -o=json", proposalID, voter, f.RPCAddr)
	cmd := govcli.GetCmdQueryVote()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var vote gov.Vote
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &vote)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return vote
}

// QueryGovVotes is lfb query gov votes
func (f *Fixtures) QueryGovVotes(proposalID int, flags ...string) gov.QueryVotesResponse {
	args := fmt.Sprintf("%d -o=json", proposalID)
	cmd := govcli.GetCmdQueryVotes()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var votes gov.QueryVotesResponse
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &votes)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votes
}

// QueryGovDeposit is lfb query gov deposit
func (f *Fixtures) QueryGovDeposit(proposalID int, depositor sdk.AccAddress, flags ...string) gov.Deposit {
	args := fmt.Sprintf("%d %s --node=%s -o=json", proposalID, depositor, f.RPCAddr)
	cmd := govcli.GetCmdQueryDeposit()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var deposit gov.Deposit
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &deposit)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposit
}

// QueryGovDeposits is lfb query gov deposits
func (f *Fixtures) QueryGovDeposits(propsalID int, flags ...string) gov.QueryDepositsResponse {
	args := fmt.Sprintf("%d -o=json", propsalID)
	cmd := govcli.GetCmdQueryDeposits()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var deposits gov.QueryDepositsResponse
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &deposits)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposits
}

// ___________________________________________________________________________________
// query slashing

// QuerySigningInfo returns the signing info for a validator
func (f *Fixtures) QuerySigningInfo(val string, flags ...string) slashing.ValidatorSigningInfo {
	args := fmt.Sprintf("%s --node=%s -o=json", val, f.RPCAddr)
	cmd := slashingcli.GetCmdQuerySigningInfo()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	cdc, _ := app.MakeCodecs()
	var sinfo slashing.ValidatorSigningInfo
	err = cdc.UnmarshalJSON(out.Bytes(), &sinfo)
	require.NoError(f.T, err)
	return sinfo
}

// QuerySlashingParams is lfb query slashing params
func (f *Fixtures) QuerySlashingParams(flags ...string) slashing.Params {
	args := fmt.Sprintf("--node=%s -o=json", f.RPCAddr)
	cmd := slashingcli.GetCmdQueryParams()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	cdc, _ := app.MakeCodecs()
	var params slashing.Params
	err = cdc.UnmarshalJSON(out.Bytes(), &params)
	require.NoError(f.T, err)
	return params
}

// ___________________________________________________________________________________
// query distribution

// QueryRewards returns the rewards of a delegator
func (f *Fixtures) QueryRewards(delAddr sdk.AccAddress, flags ...string) disttypes.QueryDelegatorTotalRewardsResponse {
	args := fmt.Sprintf("%s", delAddr)
	cmd := distcli.GetCmdQueryDelegatorRewards()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	var rewards disttypes.QueryDelegatorTotalRewardsResponse
	err = legacy.Cdc.UnmarshalJSON(out.Bytes(), &rewards)
	require.NoError(f.T, err)
	return rewards
}

// ___________________________________________________________________________________
// query supply

// QueryTotalSupply returns the total supply of coins
func (f *Fixtures) QueryTotalSupply(flags ...string) (totalSupply banktypes.QueryTotalSupplyResponse) {
	cmd := bankcli.GetCmdQueryTotalSupply()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags("-o=json", flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)
	cdc, _ := app.MakeCodecs()
	err = cdc.UnmarshalJSON(out.Bytes(), &totalSupply)
	require.NoError(f.T, err)
	return totalSupply
}

// QueryTotalSupplyOf returns the total supply of a given coin denom
func (f *Fixtures) QueryTotalSupplyOf(denom string, flags ...string) sdk.Coin {
	args := fmt.Sprintf("--denom=%s -o=json", denom)
	cmd := bankcli.GetCmdQueryTotalSupply()
	out, err := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.NoError(f.T, err)
	require.NotNil(f.T, out)

	_, cdc := app.MakeCodecs()
	var supplyOf sdk.Coin
	err = cdc.UnmarshalJSON(out.Bytes(), &supplyOf)
	require.NoError(f.T, err)
	return supplyOf
}

// ___________________________________________________________________________________
// tendermint rpc
func (f *Fixtures) NetInfo(flags ...string) *ostctypes.ResultNetInfo {
	ostc, err := osthttp.New(fmt.Sprintf("tcp://0.0.0.0:%s", f.Port), "/websocket")
	if err != nil {
		panic(fmt.Sprintf("failed to create Tendermint HTTP client: %s", err))
	}

	err = ostc.Start()
	require.NoError(f.T, err)
	defer func() {
		err := ostc.Stop()
		require.NoError(f.T, err)
	}()

	netInfo, err := ostc.NetInfo(context.Background())
	require.NoError(f.T, err)
	return netInfo
}

func (f *Fixtures) Status() *ostctypes.ResultStatus {
	ostc, err := osthttp.New(fmt.Sprintf("tcp://0.0.0.0:%s", f.Port), "/websocket")
	if err != nil {
		panic(fmt.Sprintf("failed to create Tendermint HTTP client: %s", err))
	}

	err = ostc.Start()
	require.NoError(f.T, err)
	defer func() {
		err := ostc.Stop()
		require.NoError(f.T, err)
	}()

	netInfo, err := ostc.Status(context.Background())
	require.NoError(f.T, err)
	return netInfo
}

// ___________________________________________________________________________________
// linkcli mempool

// MempoolNumUnconfirmedTxs is linkcli mempool num-unconfirmed-txs
func (f *Fixtures) MempoolNumUnconfirmedTxs(flags ...string) *ostctypes.ResultUnconfirmedTxs {
	ostc, err := osthttp.New(fmt.Sprintf("tcp://0.0.0.0:%s", f.Port), "/websocket")
	if err != nil {
		panic(fmt.Sprintf("failed to create Tendermint HTTP client: %s", err))
	}

	err = ostc.Start()
	require.NoError(f.T, err)
	defer func() {
		err := ostc.Stop()
		require.NoError(f.T, err)
	}()

	out, err := ostc.NumUnconfirmedTxs(context.Background())
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return out
}

// ___________________________________________________________________________________
// Fixture Group

type FixtureGroup struct {
	T                  *testing.T
	fixturesMap        map[string]*Fixtures
	Network            *testnet.Network
	networkName        string
	genesisFileContent []byte
	BaseDir            string
}

func NewFixtureGroup(t *testing.T) *FixtureGroup {
	fg := &FixtureGroup{
		T:           t,
		fixturesMap: make(map[string]*Fixtures),
		BaseDir:     getHomeDir(t),
		Network:     &testnet.Network{},
	}
	fg.networkName = networkNamePrefix + t.Name()

	return fg
}

func InitFixturesGroup(t *testing.T, numOfNodes ...int) *FixtureGroup {
	nodeNumber := 4
	if len(numOfNodes) == 1 {
		nodeNumber = numOfNodes[0]
	}
	fg := NewFixtureGroup(t)
	fg.initNodes(nodeNumber)
	return fg
}

func (fg *FixtureGroup) initNodes(numberOfNodes int) {
	t := fg.T

	// Initialize temporary directories
	gentxDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(gentxDir))
	}()

	for idx := 0; idx < numberOfNodes; idx++ {
		name := fmt.Sprintf("%s-%s%d", t.Name(), namePrefix, idx)
		f := NewFixtures(t, filepath.Join(fg.BaseDir, name))
		f.LFBInit(name, fmt.Sprintf("--chain-id=%s", fg.T.Name()))
		fg.fixturesMap[name] = f
		require.NoError(fg.T, err)
	}
	for name, f := range fg.fixturesMap {
		f.KeysDelete(name)
		f.KeysAdd(name)
	}

	for _, f := range fg.fixturesMap {
		for nameInner, fInner := range fg.fixturesMap {
			f.AddGenesisAccount(fInner.KeyAddress(nameInner), startCoins)
		}
	}

	for name, f := range fg.fixturesMap {
		gentxDoc := filepath.Join(gentxDir, fmt.Sprintf("%s.json", name))
		f.GenTx(name, fmt.Sprintf("--output-document=%s", gentxDoc))
	}

	for _, f := range fg.fixturesMap {
		f.CollectGenTxs(fmt.Sprintf("--gentx-dir=%s", gentxDir))
		f.ValidateGenesis(filepath.Join(f.Home, "config", "genesis.json"))
		if len(fg.genesisFileContent) == 0 {
			fg.genesisFileContent, err = ioutil.ReadFile(f.GenesisFile())
			require.NoError(t, err)
		}
	}

	for _, f := range fg.fixturesMap {
		err := ioutil.WriteFile(f.GenesisFile(), fg.genesisFileContent, os.ModePerm)
		require.NoError(t, err)
	}
}
func (fg *FixtureGroup) LFBStartCluster(flags ...string) {
	genDoc, err := osttypes.GenesisDocFromJSON(fg.genesisFileContent)
	require.NoError(fg.T, err)

	var appState app.GenesisState
	require.NoError(fg.T, legacy.Cdc.UnmarshalJSON(genDoc.AppState, &appState))

	cfg := newTestnetConfig(fg.T, appState, fg.T.Name(), "")

	validators := make([]*testnet.Validator, 0)
	for _, f := range fg.fixturesMap {
		validators = append(validators, newValidator(f, cfg, srvconfig.DefaultConfig(), server.NewDefaultContext()))
	}

	fg.Network = testnet.NewWithoutInit(fg.T, cfg, fg.BaseDir, validators)
}

func (fg *FixtureGroup) AddFullNode(flags ...string) *Fixtures {
	t := fg.T
	idx := len(fg.fixturesMap)
	chainID := fg.T.Name()

	name := fmt.Sprintf("%s-%s%d", t.Name(), namePrefix, idx)
	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	appCfg := srvconfig.DefaultConfig()
	ctx := server.NewDefaultContext()

	f := NewFixtures(t, tmpDir)

	// Initialize lfb
	{
		f.LFBInit(name, fmt.Sprintf("--chain-id=%s", chainID))
		f.KeysDelete(name)
		f.KeysAdd(name)
	}

	// Copy the genesis.json
	{
		if len(fg.genesisFileContent) == 0 {
			panic("genesis file is not loaded")
		}
		err := ioutil.WriteFile(f.GenesisFile(), fg.genesisFileContent, os.ModePerm)
		require.NoError(t, err)
	}

	// Configure for invisible options
	{
		if len(flags) > 0 {
			for _, flag := range flags {
				if flag == "--mempool.broadcast=false" {
					ctx.Config.Mempool.Broadcast = false
				}
			}
		}
	}

	// Add persistent peers
	{
		var persistentPeers []string

		for _, val := range fg.Network.Validators {
			p2pAddr, err := url.Parse(val.P2PAddress)
			require.NoError(t, err)
			persistentPeers = append(persistentPeers, fmt.Sprintf("%s@%s:%s", val.NodeID, p2pAddr.Hostname(), p2pAddr.Port()))
		}

		ctx.Config.P2P.PersistentPeers = strings.Join(persistentPeers, ",")
	}

	// Add fixture to the group
	fg.fixturesMap[name] = f

	// Start lfb
	validator := newValidator(f, fg.Network.Config, appCfg, ctx)
	testnet.AddNewValidator(t, fg.Network, validator)
	WaitForTMStart(f.Port)
	require.NoError(t, fg.Network.WaitForNextBlock())
	return f
}

func (fg *FixtureGroup) Fixture(index int) *Fixtures {
	name := fmt.Sprintf("%s-%s%d", fg.T.Name(), namePrefix, index)
	if f, ok := fg.fixturesMap[name]; ok {
		return f
	}
	return nil
}

func (fg *FixtureGroup) Cleanup() {
	fg.Network.Cleanup()
	for _, f := range fg.fixturesMap {
		f.Cleanup()
	}
}

// ___________________________________________________________________________________
// utils

// wait for tendermint to start by querying tendermint
func WaitForTMStart(port string) {
	url := fmt.Sprintf("http://localhost:%v/block", port)
	WaitForStart(url)
}

// WaitForStart waits for the node to start by pinging the url
// every 100ms for 10s until it returns 200. If it takes longer than 5s,
// it panics.
func WaitForStart(url string) {
	var err error

	// ping the status endpoint a few times a second
	// for a few seconds until we get a good response.
	// otherwise something probably went wrong
	// 2 ^ 7 = 128 --> approximately 10 secs
	wait := 1
	for i := 0; i < 7; i++ {
		// 0.1, 0.2, 0.4, 0.8, 1.6, 3.2, 6.4, 12.8, 25.6, 51.2, 102.4
		time.Sleep(time.Millisecond * 100 * time.Duration(wait))
		wait *= 2

		var res *http.Response
		/* #nosec */
		res, err = http.Get(url) // Error is arising in testing files
		if err != nil || res == nil {
			continue
		}
		err = res.Body.Close()
		if err != nil {
			panic(err)
		}

		if res.StatusCode == http.StatusOK {
			// good!
			return
		}
	}
	// still haven't started up?! panic!
	panic(err)
}

func addFlags(args string, flags ...string) []string {
	return append(strings.Split(args, " "), flags...)
}

// Write the given string to a new temporary file
func WriteToNewTempFile(t *testing.T, s string) *os.File {
	fp, err := ioutil.TempFile(os.TempDir(), "cosmos_cli_test_")
	require.Nil(t, err)
	_, err = fp.WriteString(s)
	require.Nil(t, err)
	return fp
}

func MarshalTx(t *testing.T, stdTx tx.Tx) []byte {
	cdc, _ := app.MakeCodecs()
	bz, err := cdc.MarshalBinaryBare(&stdTx)
	require.NoError(t, err)
	return bz
}

func UnmarshalTx(t *testing.T, s []byte) (stdTx tx.Tx) {
	cdc, _ := app.MakeCodecs()
	require.Nil(t, cdc.UnmarshalJSON(s, &stdTx))
	return
}

func UnmarshalTxResponse(t *testing.T, s []byte) (txResp sdk.TxResponse) {
	cdc, _ := app.MakeCodecs()
	require.Nil(t, cdc.UnmarshalJSON(s, &txResp))
	return
}

func newTestnetConfig(t *testing.T, genesisState map[string]json.RawMessage, chainID, minGasPrices string) testnet.Config {
	encodingCfg := app.MakeEncodingConfig()
	cfg := testnet.Config{
		Codec:             encodingCfg.Marshaler,
		TxConfig:          encodingCfg.TxConfig,
		LegacyAmino:       encodingCfg.Amino,
		InterfaceRegistry: encodingCfg.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		GenesisState:      genesisState,
		TimeoutCommit:     2 * time.Second,
		ChainID:           chainID,
		NumValidators:     1,
		BondDenom:         sdk.DefaultBondDenom,
		MinGasPrices:      minGasPrices,
		AccountTokens:     sdk.TokensFromConsensusPower(1000),
		StakingTokens:     sdk.TokensFromConsensusPower(500),
		BondedTokens:      sdk.TokensFromConsensusPower(100),
		PruningStrategy:   storetypes.PruningOptionNothing,
		EnableLogging:     true,
		CleanupDir:        true,
		SigningAlgo:       string(hd.Secp256k1Type),
		KeyringOptions:    []keyring.Option{},
	}

	cfg.AppConstructor = func(val testnet.Validator) servertypes.Application {
		db, err := sdk.NewLevelDB("application", filepath.Join(val.Dir, "data"))
		require.NoError(t, err)
		return app.NewLinkApp(val.Ctx.Logger, db, nil, true, make(map[int64]bool), val.Dir, 0,
			encodingCfg,
			val.Ctx.Viper,
			nil,
			baseapp.SetPruning(storetypes.NewPruningOptionsFromString(storetypes.PruningOptionNothing)),
			baseapp.SetMinGasPrices(minGasPrices),
		)
	}

	return cfg
}

func newValidator(f *Fixtures, cfg testnet.Config, appCfg *srvconfig.Config, ctx *server.Context) *testnet.Validator {
	buf := bufio.NewReader(os.Stdin)

	appCfg.Pruning = cfg.PruningStrategy
	appCfg.MinGasPrices = cfg.MinGasPrices
	appCfg.API.Enable = true
	appCfg.API.Swagger = false
	appCfg.Telemetry.Enabled = false

	tmCfg := ctx.Config
	tmCfg.Consensus.TimeoutCommit = cfg.TimeoutCommit

	appCfg.API.Address = f.P2PAddr
	tmCfg.P2P.ListenAddress = f.P2PAddr
	tmCfg.RPC.ListenAddress = f.RPCAddr
	appCfg.GRPC.Address = f.GRPCAddr
	appCfg.GRPC.Enable = true

	logger := log.NewNopLogger()
	var err error
	if cfg.EnableLogging {
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		logger, err = ostflags.ParseLogLevel("info", logger, ostcfg.DefaultLogLevel)
		require.NoError(f.T, err)
	}

	ctx.Logger = logger

	require.NoError(f.T, os.MkdirAll(filepath.Join(f.Home, "config"), 0755))

	tmCfg.SetRoot(f.Home)
	tmCfg.Moniker = f.Moniker

	tmCfg.P2P.AddrBookStrict = false
	tmCfg.P2P.AllowDuplicateIP = true

	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, f.Home, buf, cfg.KeyringOptions...)
	require.NoError(f.T, err)

	srvconfig.WriteConfigFile(filepath.Join(f.Home, "config/app.toml"), appCfg)

	id, pubKey, err := genutil.InitializeNodeValidatorFiles(tmCfg)
	require.NoError(f.T, err)

	clientCtx := client.Context{}.
		WithKeyring(kb).
		WithHomeDir(tmCfg.RootDir).
		WithChainID(cfg.ChainID).
		WithInterfaceRegistry(cfg.InterfaceRegistry).
		WithJSONMarshaler(cfg.Codec).
		WithLegacyAmino(cfg.LegacyAmino).
		WithTxConfig(cfg.TxConfig).
		WithAccountRetriever(cfg.AccountRetriever)

	return &testnet.Validator{
		AppConfig:  appCfg,
		ClientCtx:  clientCtx,
		Ctx:        ctx,
		Dir:        f.Home,
		Moniker:    f.Moniker,
		RPCAddress: tmCfg.RPC.ListenAddress,
		P2PAddress: tmCfg.P2P.ListenAddress,
		NodeID:     id,
		PubKey:     pubKey,
	}
}

func (f *Fixtures) TxStoreWasm(wasmFilePath string, flags ...string) (testutil.BufferWriter, error) {
	cmd := wasmcli.StoreCodeCmd()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(wasmFilePath, flags...))
}

func (f *Fixtures) TxInstantiateWasm(codeID uint64, msgJSON string, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("%d %s", codeID, msgJSON)
	cmd := wasmcli.InstantiateContractCmd()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

func (f *Fixtures) TxExecuteWasm(contractAddress string, msgJSON string, flags ...string) (testutil.BufferWriter, error) {
	args := fmt.Sprintf("%s %s", contractAddress, msgJSON)
	cmd := wasmcli.ExecuteContractCmd()
	return testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
}

func (f *Fixtures) QueryListCodeWasm(flags ...string) wasmtypes.QueryCodesResponse {
	cmd := wasmcli.GetCmdListCode()
	res, errStr := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags("-o=json", flags...))

	require.Empty(f.T, errStr)
	cdc, _ := app.MakeCodecs()
	var queryCodesResponse wasmtypes.QueryCodesResponse

	err := cdc.UnmarshalJSON(res.Bytes(), &queryCodesResponse)
	require.NoError(f.T, err)
	return queryCodesResponse
}

func (f *Fixtures) QueryCodeWasm(codeID uint64, outputPath string, flags ...string) {
	args := fmt.Sprintf("%d %s -o=json", codeID, outputPath)
	cmd := wasmcli.GetCmdQueryCode()
	_, errStr := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.Empty(f.T, errStr)
}

func (f *Fixtures) QueryListContractByCodeWasm(codeID uint64, flags ...string) wasmtypes.QueryContractsByCodeResponse {
	args := fmt.Sprintf("%d -o=json", codeID)
	cmd := wasmcli.GetCmdListContractByCode()
	res, errStr := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))

	require.Empty(f.T, errStr)
	cdc, _ := app.MakeCodecs()
	var queryContractsByCodeResponse wasmtypes.QueryContractsByCodeResponse

	err := cdc.UnmarshalJSON(res.Bytes(), &queryContractsByCodeResponse)
	require.NoError(f.T, err)
	return queryContractsByCodeResponse
}

func (f *Fixtures) QueryContractStateSmartWasm(contractAddress string, reqJSON string, flags ...string) string {
	args := fmt.Sprintf("%s %s -o=json", contractAddress, reqJSON)
	cmd := wasmcli.GetCmdGetContractStateSmart()
	res, errStr := testcli.ExecTestCLICmd(getCliCtx(f), cmd, addFlags(args, flags...))
	require.Empty(f.T, errStr)
	return res.String()
}
