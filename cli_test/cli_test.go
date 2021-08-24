// +build cli_test

package clitest

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/line/lfb-sdk/client/flags"
	"github.com/line/lfb-sdk/crypto/keys/ed25519"
	sdk "github.com/line/lfb-sdk/types"
	"github.com/line/lfb-sdk/types/tx"
	gov "github.com/line/lfb-sdk/x/gov/types"
	minttypes "github.com/line/lfb-sdk/x/mint/types"
	"github.com/line/lfb/app"
	osttypes "github.com/line/ostracon/types"
	"github.com/stretchr/testify/require"
)

func TestLFBKeysAddMultisig(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// key names order does not matter
	f.KeysAdd("msig1", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", keyBar, keyBaz))

	// err : trying to save same pubkey
	msig1Addr := f.KeysShow("msig1").Address
	f.KeysDelete("msig1")

	f.KeysAdd("msig2", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", keyBaz, keyBar))
	require.Equal(t, msig1Addr, f.KeysShow("msig2").Address)

	f.KeysAdd("msig3", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", keyBar, keyBaz),
		"--nosort")
	f.KeysAdd("msig4", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", keyBaz, keyBar),
		"--nosort")
	require.NotEqual(t, f.KeysShow("msig3").Address, f.KeysShow("msig4").Address)
}

func TestLFBMinimumFees(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server with minimum fees
	minGasPrice, err := sdk.NewDecFromStr("0.000006")
	require.NoError(t, err)

	fees := fmt.Sprintf(
		"%s,%s",
		sdk.NewDecCoinFromDec(feeDenom, minGasPrice),
		sdk.NewDecCoinFromDec(fee2Denom, minGasPrice),
	)

	n := f.LFBStart(fees)
	defer n.Cleanup()

	barAddr := f.KeyAddress(keyBar)

	// Send a transaction that will get rejected
	out, err := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(fee2Denom, 10), "-y")
	require.NoError(t, err)
	require.Contains(t, out.String(), "insufficient fees")

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure tx w/ correct fees pass
	txFees := fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 2))
	_, err = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(fee2Denom, 10), txFees, "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure tx w/ improper fees fails
	txFees = fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 1))
	out, err = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 10), txFees, "-y")
	require.NoError(t, err)
	require.Contains(t, out.String(), "insufficient fees")
}

func TestLFBGasPrices(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server with minimum fees
	minGasPrice, err := sdk.NewDecFromStr("0.000006")
	require.NoError(t, err)

	n := f.LFBStart(sdk.NewDecCoinFromDec(feeDenom, minGasPrice).String())
	defer n.Cleanup()

	barAddr := f.KeyAddress(keyBar)

	// insufficient gas prices (tx fails)
	badGasPrice, err := sdk.NewDecFromStr("0.000003")
	require.NoError(t, err)

	out, err := f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 50),
		fmt.Sprintf("--gas-prices=%s", sdk.NewDecCoinFromDec(feeDenom, badGasPrice)), "-y")
	require.NoError(t, err)
	require.Contains(t, out.String(), "insufficient fees")

	// wait for a block confirmation
	require.NoError(t, n.WaitForNextBlock())

	// sufficient gas prices (tx passes)
	_, err = f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 50),
		fmt.Sprintf("--gas-prices=%s", sdk.NewDecCoinFromDec(feeDenom, minGasPrice)), "-y")
	require.NoError(t, err)

	// wait for a block confirmation
	require.NoError(t, n.WaitForNextBlock())
}

func TestLFBFeesDeduction(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server with minimum fees
	minGasPrice, err := sdk.NewDecFromStr("0.000006")
	require.NoError(t, err)

	n := f.LFBStart(sdk.NewDecCoinFromDec(feeDenom, minGasPrice).String())
	defer n.Cleanup()

	// Save key addresses for later use
	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	fooBal := f.QueryBalances(fooAddr)
	fooAmt := fooBal.GetBalances().AmountOf(fooDenom)

	// test simulation
	_, err = f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 1000),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 2)), "--dry-run")
	require.NoError(t, err)

	// Wait for a block
	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// ensure state didn't change
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, fooAmt.Int64(), fooBal.GetBalances().AmountOf(fooDenom).Int64())

	// insufficient funds (coins + fees) tx fails
	largeCoins := sdk.TokensFromConsensusPower(10000000)
	out, err := f.TxSend(
		keyFoo, barAddr, sdk.NewCoin(fooDenom, largeCoins),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 2)), "-y")
	require.NoError(t, err)
	require.Contains(t, out.String(), "insufficient funds")

	// Wait for a block
	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// ensure state didn't change
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, fooAmt.Int64(), fooBal.GetBalances().AmountOf(fooDenom).Int64())

	// test success (transfer = coins + fees)
	_, err = f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 500),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 2)), "-y")
	require.NoError(t, err)
}

func TestLFBSend(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	// Save key addresses for later use
	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	fooBal := f.QueryBalances(fooAddr)

	fmt.Println(fooBal)

	startTokens := sdk.TokensFromConsensusPower(50)
	fmt.Println(startTokens.Uint64())
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(denom))

	// Send some tokens from one account to the other
	sendTokens := sdk.TokensFromConsensusPower(10)
	_, err := f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure account balances match expected
	barBal := f.QueryBalances(barAddr)
	require.Equal(t, sendTokens, barBal.GetBalances().AmountOf(denom))
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(sendTokens), fooBal.GetBalances().AmountOf(denom))

	// Test --dry-run
	_, err = f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "--dry-run")
	require.NoError(t, err)

	// Test --generate-only
	out, err := f.TxSend(
		fooAddr.String(), barAddr, sdk.NewCoin(denom, sendTokens), "--generate-only=true",
	)
	require.NoError(t, err)
	msg := UnmarshalTx(t, out.Bytes())
	require.NotZero(t, msg.GetAuthInfo().GetFee().GetGasLimit())
	require.Len(t, msg.GetMsgs(), 1)
	require.Len(t, msg.GetSignatures(), 0)

	// Check state didn't change
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(sendTokens), fooBal.GetBalances().AmountOf(denom))

	// test autosequencing
	_, err = f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure account balances match expected
	barBal = f.QueryBalances(barAddr)
	require.Equal(t, sendTokens.MulRaw(2), barBal.GetBalances().AmountOf(denom))
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(sendTokens.MulRaw(2)), fooBal.GetBalances().AmountOf(denom))

	// test memo
	_, err = f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "--memo='testmemo'", "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure account balances match expected
	barBal = f.QueryBalances(barAddr)
	require.Equal(t, sendTokens.MulRaw(3), barBal.GetBalances().AmountOf(denom))
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(sendTokens.MulRaw(3)), fooBal.GetBalances().AmountOf(denom))
}

func TestLFBGasAuto(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	fooBal := f.QueryBalances(fooAddr)
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(denom))

	// Test failure with auto gas disabled and very little gas set by hand
	sendTokens := sdk.TokensFromConsensusPower(10)
	out, err := f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "--gas=10", "-y")
	require.NoError(t, err)
	require.Contains(t, out.String(), "out of gas in location")

	// Check state didn't change
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(denom))

	// Test failure with negative gas
	_, err = f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "--gas=-100", "-y")
	require.NoError(t, err)

	// Check state didn't change
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(denom))

	// Test failure with 0 gas
	out, err = f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "--gas=0", "-y")
	require.NoError(t, err)
	require.Contains(t, out.String(), "out of gas in location")

	// Check state didn't change
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(denom))

	// Enable auto gas
	out, err = f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "--gas=auto", "-y")
	require.NoError(t, err)
	cdc, _ := app.MakeCodecs()
	sendResp := sdk.TxResponse{}
	err = cdc.UnmarshalJSON(out.Bytes(), &sendResp)
	require.Nil(t, err)
	require.True(t, sendResp.GasWanted >= sendResp.GasUsed)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Check state has changed accordingly
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(sendTokens), fooBal.GetBalances().AmountOf(denom))
}

func TestLFBCreateValidator(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	barAddr := f.KeyAddress(keyBar)
	barVal := sdk.ValAddress(barAddr)

	consPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())

	sendTokens := sdk.TokensFromConsensusPower(10)
	_, err := f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	barBal := f.QueryBalances(barAddr)
	require.Equal(t, sendTokens, barBal.GetBalances().AmountOf(denom))

	// Generate a create validator transaction and ensure correctness
	out, err := f.TxStakingCreateValidator(barAddr.String(), consPubKey, sdk.NewInt64Coin(denom, 2), "--generate-only")

	require.NoError(t, err)
	fmt.Println(out.String())
	msg := UnmarshalTx(t, out.Bytes())
	require.NotZero(t, msg.GetAuthInfo().GetFee().GetGasLimit())
	require.Equal(t, len(msg.GetMsgs()), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	newValTokens := sdk.TokensFromConsensusPower(2)
	_, err = f.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewCoin(denom, newValTokens), "--dry-run")
	require.NoError(t, err)

	// Create the validator
	_, err = f.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewCoin(denom, newValTokens), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure funds were deducted properly
	barBal = f.QueryBalances(barAddr)
	require.Equal(t, sendTokens.Sub(newValTokens), barBal.GetBalances().AmountOf(denom))

	// Ensure that validator state is as expected
	validator := f.QueryStakingValidator(barVal)
	require.Equal(t, validator.OperatorAddress, barVal.String())
	require.True(sdk.IntEq(t, newValTokens, validator.Tokens))

	// Query delegations to the validator
	validatorDelegations := f.QueryStakingDelegationsTo(barVal)
	require.Len(t, validatorDelegations.DelegationResponses, 1)
	require.NotZero(t, validatorDelegations.GetDelegationResponses()[0].GetDelegation().GetShares())

	// unbond a single share
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(1))
	_, err = f.TxStakingUnbond(keyBar, unbondAmt.String(), barVal, "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure bonded staking is correct
	remainingTokens := newValTokens.Sub(unbondAmt.Amount)
	validator = f.QueryStakingValidator(barVal)
	require.Equal(t, remainingTokens, validator.Tokens)

	// Get unbonding delegations from the validator
	validatorUbds := f.QueryStakingUnbondingDelegationsFrom(barVal)
	require.Len(t, validatorUbds.GetUnbondingResponses(), 1)
	require.Len(t, validatorUbds.GetUnbondingResponses()[0].Entries, 1)
	require.Equal(t, remainingTokens.String(), validatorUbds.GetUnbondingResponses()[0].Entries[0].Balance.String())
}

func TestLFBQuerySupply(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	totalSupplyOf := f.QueryTotalSupplyOf(fooDenom)

	require.True(sdk.IntEq(t, TotalCoins.AmountOf(fooDenom), totalSupplyOf.Amount))
}

func TestLFBSubmitProposal(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	f.QueryGovParamDeposit()
	f.QueryGovParamVoting()
	f.QueryGovParamTallying()

	fooAddr := f.KeyAddress(keyFoo)

	fooBal := f.QueryBalances(fooAddr)
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(sdk.DefaultBondDenom))

	proposalsQuery := f.QueryGovProposals()
	require.Empty(t, proposalsQuery)

	// Test submit generate only for submit proposal
	proposalTokens := sdk.TokensFromConsensusPower(5)
	out, err := f.TxGovSubmitProposal(
		fooAddr.String(), "Text", "Test", "test", sdk.NewCoin(denom, proposalTokens), "--generate-only", "-y")
	require.NoError(t, err)
	msg := UnmarshalTx(t, out.Bytes())
	require.NotZero(t, msg.GetAuthInfo().GetFee().GetGasLimit())
	require.Equal(t, len(msg.GetMsgs()), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	_, err = f.TxGovSubmitProposal(keyFoo, "Text", "Test", "test", sdk.NewCoin(denom, proposalTokens), "--dry-run")
	require.NoError(t, err)

	// Create the proposal
	_, err = f.TxGovSubmitProposal(keyFoo, "Text", "Test", "test", sdk.NewCoin(denom, proposalTokens), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure transaction tags can be queried
	searchResult := f.QueryTxs(1, 50, fmt.Sprintf("--events='message.action=%s&message.sender=%s'", "submit_proposal", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure deposit was deducted
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(proposalTokens), fooBal.GetBalances().AmountOf(denom))

	// Ensure propsal is directly queryable
	proposal1 := f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.ProposalId)
	require.Equal(t, gov.StatusDepositPeriod, proposal1.Status)

	// Ensure query proposals returns properly
	proposalsQuery = f.QueryGovProposals()
	require.Equal(t, uint64(1), proposalsQuery.GetProposals()[0].ProposalId)

	// Query the deposits on the proposal
	deposit := f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, proposalTokens, deposit.Amount.AmountOf(denom))

	// Test deposit generate only
	depositTokens := sdk.TokensFromConsensusPower(10)
	out, err = f.TxGovDeposit(1, fooAddr.String(), sdk.NewCoin(denom, depositTokens), "--generate-only")
	require.NoError(t, err)
	msg = UnmarshalTx(t, out.Bytes())
	require.NotZero(t, msg.GetAuthInfo().GetFee().GetGasLimit())
	require.Equal(t, len(msg.GetMsgs()), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Run the deposit transaction
	_, err = f.TxGovDeposit(1, keyFoo, sdk.NewCoin(denom, depositTokens), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// test query deposit
	deposits := f.QueryGovDeposits(1)
	require.Len(t, deposits.GetDeposits(), 1)
	require.Equal(t, proposalTokens.Add(depositTokens), deposits.GetDeposits()[0].Amount.AmountOf(denom))

	// Ensure querying the deposit returns the proper amount
	deposit = f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, proposalTokens.Add(depositTokens), deposit.Amount.AmountOf(denom))

	// Ensure tags are set on the transaction
	searchResult = f.QueryTxs(1, 50, fmt.Sprintf("--events='message.action=%s&message.sender=%s'", "deposit", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure account has expected amount of funds
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(proposalTokens.Add(depositTokens)), fooBal.GetBalances().AmountOf(denom))

	// Fetch the proposal and ensure it is now in the voting period
	proposal1 = f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.ProposalId)
	require.Equal(t, gov.StatusVotingPeriod, proposal1.Status)

	// Test vote generate only
	out, err = f.TxGovVote(1, gov.OptionYes, fooAddr.String(), "--generate-only")
	require.NoError(t, err)
	msg = UnmarshalTx(t, out.Bytes())
	require.NotZero(t, msg.GetAuthInfo().GetFee().GetGasLimit())
	require.Equal(t, len(msg.GetMsgs()), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Vote on the proposal
	_, err = f.TxGovVote(1, gov.OptionYes, keyFoo, "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Query the vote
	vote := f.QueryGovVote(1, fooAddr)
	require.Equal(t, uint64(1), vote.ProposalId)
	require.Equal(t, gov.OptionYes, vote.Option)

	// Query the votes
	votes := f.QueryGovVotes(1)
	require.Len(t, votes.GetVotes(), 1)
	require.Equal(t, uint64(1), votes.GetVotes()[0].ProposalId)
	require.Equal(t, gov.OptionYes, votes.GetVotes()[0].Option)

	// Ensure tags are applied to voting transaction properly
	searchResult = f.QueryTxs(1, 50, fmt.Sprintf("--events='message.action=%s&message.sender=%s'", "vote", fooAddr))
	require.Len(t, searchResult.Txs, 1)

	// Ensure no proposals in deposit period
	proposalsQuery = f.QueryGovProposals("--status=DepositPeriod")
	require.Empty(t, proposalsQuery)

	// Ensure the proposal returns as in the voting period
	proposalsQuery = f.QueryGovProposals("--status=VotingPeriod")
	require.Equal(t, uint64(1), proposalsQuery.GetProposals()[0].ProposalId)

	// submit a second test proposal
	_, err = f.TxGovSubmitProposal(keyFoo, "Text", "Apples", "test", sdk.NewCoin(denom, proposalTokens), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Test limit on proposals query
	proposalsQuery = f.QueryGovProposals("--limit=2")
	require.Equal(t, uint64(2), proposalsQuery.GetProposals()[1].ProposalId)
}

func TestLFBSubmitParamChangeProposal(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	n := f.LFBStart("")
	defer n.Cleanup()

	fooAddr := f.KeyAddress(keyFoo)
	fooBal := f.QueryBalances(fooAddr)
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(sdk.DefaultBondDenom))

	// write proposal to file
	proposalTokens := sdk.TokensFromConsensusPower(5)

	proposal := fmt.Sprintf(`{
"title": "Param Change",
"description": "Update max validators",
"type": "Text",
"deposit": "%sstake"
}`, proposalTokens.String())

	proposalFile := WriteToNewTempFile(t, proposal)

	// create the param change proposal
	out, err := f.TxGovSubmitParamChangeProposal(keyFoo, proposalFile.Name(), "-y")
	fmt.Println(out.String())
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// ensure transaction tags can be queried
	txsPage := f.QueryTxs(1, 50, fmt.Sprintf("--events='message.action=%s&message.sender=%s'", "submit_proposal", fooAddr))
	require.Len(t, txsPage.Txs, 1)

	// ensure deposit was deducted
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(proposalTokens).String(), fooBal.GetBalances().AmountOf(sdk.DefaultBondDenom).String())

	// ensure proposal is directly queryable
	proposal1 := f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.ProposalId)
	require.Equal(t, gov.StatusDepositPeriod, proposal1.Status)

	// ensure correct query proposals result
	proposalsQuery := f.QueryGovProposals()
	require.Equal(t, uint64(1), proposalsQuery.GetProposals()[0].ProposalId)

	// ensure the correct deposit amount on the proposal
	deposit := f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, proposalTokens, deposit.Amount.AmountOf(denom))
}

func TestLFBSubmitCommunityPoolSpendProposal(t *testing.T) {
	t.Skip("Due to removing mint module")
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// create some inflation
	var cdc, amino = app.MakeCodecs()
	genesisState := f.GenesisState()
	inflationMin := sdk.MustNewDecFromStr("10000.0")
	var mintData minttypes.GenesisState
	err := cdc.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData)
	require.NoError(t, err)

	mintData.Minter.Inflation = inflationMin
	mintData.Params.InflationMin = inflationMin
	mintData.Params.InflationMax = sdk.MustNewDecFromStr("15000.0")
	mintDataBz, err := cdc.MarshalJSON(&mintData)
	require.NoError(t, err)
	genesisState[minttypes.ModuleName] = mintDataBz

	genFile := filepath.Join(f.Home, "config", "genesis.json")
	genDoc, err := osttypes.GenesisDocFromFile(genFile)
	require.NoError(t, err)
	genDoc.AppState, err = amino.MarshalJSON(genesisState)
	require.NoError(t, err)
	require.NoError(t, genDoc.SaveAs(genFile))

	n := f.LFBStart("")
	defer n.Cleanup()

	fooAddr := f.KeyAddress(keyFoo)
	fooBal := f.QueryBalances(fooAddr)
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(sdk.DefaultBondDenom))

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// write proposal to file
	proposalTokens := sdk.TokensFromConsensusPower(5)

	proposal := fmt.Sprintf(`{
"title": "Community Pool Spend",
"description": "Spend from community pool",
"type": "Text"
"deposit": "%s%s"
}
`, sdk.DefaultBondDenom, proposalTokens.String())

	proposalFile := WriteToNewTempFile(t, proposal)

	// create the param change proposal
	_, err = f.TxGovSubmitCommunityPoolSpendProposal(keyFoo, proposalFile.Name(), sdk.NewCoin(denom, proposalTokens), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// ensure transaction tags can be queried
	txsPage := f.QueryTxs(1, 50, fmt.Sprintf("--events='message.action=%s&message.sender=%s'", "submit_proposal", fooAddr))
	require.Len(t, txsPage.Txs, 1)

	// ensure deposit was deducted
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, startTokens.Sub(proposalTokens).String(), fooBal.GetBalances().AmountOf(sdk.DefaultBondDenom).String())

	// ensure proposal is directly queryable
	proposal1 := f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.ProposalId)
	require.Equal(t, gov.StatusDepositPeriod, proposal1.Status)

	// ensure correct query proposals result
	proposalsQuery := f.QueryGovProposals()
	require.Equal(t, uint64(1), proposalsQuery.GetProposals()[0].ProposalId)

	// ensure the correct deposit amount on the proposal
	deposit := f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, proposalTokens, deposit.Amount.AmountOf(denom))
}

func TestLFBQueryTxPagination(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	accFoo := f.QueryAccount(fooAddr)
	seq := accFoo.GetSequence()

	for i := 1; i <= 4; i++ {
		_, err := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, int64(i)), fmt.Sprintf("--sequence=%d", seq), "-y")
		require.NoError(t, err)
		seq++
	}

	// perPage = 15, 2 pages
	txsPage1 := f.QueryTxs(1, 2, fmt.Sprintf("--events='message.sender=%s'", fooAddr))
	require.Len(t, txsPage1.Txs, 2)
	require.Equal(t, txsPage1.Count, uint64(2))
	txsPage2 := f.QueryTxs(2, 2, fmt.Sprintf("--events='message.sender=%s'", fooAddr))
	require.Len(t, txsPage2.Txs, 2)
	require.NotEqual(t, txsPage1.Txs, txsPage2.Txs)

	// perPage = 16, 2 pages
	txsPage1 = f.QueryTxs(1, 3, fmt.Sprintf("--events='message.sender=%s'", fooAddr))
	require.Len(t, txsPage1.Txs, 3)
	txsPage2 = f.QueryTxs(2, 3, fmt.Sprintf("--events='message.sender=%s'", fooAddr))
	require.Len(t, txsPage2.Txs, 1)
	require.NotEqual(t, txsPage1.Txs, txsPage2.Txs)

	// perPage = 50
	txsPageFull := f.QueryTxs(1, 50, fmt.Sprintf("--events='message.sender=%s'", fooAddr))
	require.Len(t, txsPageFull.Txs, 4)
	require.Equal(t, txsPageFull.Txs, append(txsPage1.Txs, txsPage2.Txs...))

	// perPage = 0
	f.QueryTxsInvalid(errors.New("page must greater than 0"), 0, 50, fmt.Sprintf("--events='message.sender=%s'", fooAddr))

	// limit = 0
	f.QueryTxsInvalid(errors.New("limit must greater than 0"), 1, 0, fmt.Sprintf("--events='message.sender=%s'", fooAddr))

	// no events
	f.QueryTxsInvalid(errors.New("required flag(s) \"events\" not set"), 1, 30)
}

func TestLFBValidateSignatures(t *testing.T) {
	t.Skip("no flag validate-signatures")
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	// generate sendTx with default gas
	out, err := f.TxSend(fooAddr.String(), barAddr, sdk.NewInt64Coin(denom, 10), "--generate-only")
	require.NoError(t, err)

	// write  unsigned tx to file
	unsignedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(unsignedTxFile.Name())

	// validate we can successfully sign
	out, err = f.TxSign(keyFoo, unsignedTxFile.Name())
	require.NoError(t, err)
	stdTx := UnmarshalTx(t, out.Bytes())
	require.Equal(t, len(stdTx.GetMsgs()), 1)
	require.Equal(t, 1, len(stdTx.GetSignatures()))
	require.Equal(t, fooAddr.String(), stdTx.GetSigners()[0].String())

	// write signed tx to file
	signedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(signedTxFile.Name())

	// validate signatures
	_, err = f.TxSign(keyFoo, signedTxFile.Name(), "--validate-signatures")
	require.NoError(t, err)

	// modify the transaction
	stdTx.Body.Memo = "MODIFIED-ORIGINAL-TX-BAD"
	bz := MarshalTx(t, stdTx)
	modSignedTxFile := WriteToNewTempFile(t, string(bz))
	defer os.Remove(modSignedTxFile.Name())

	// validate signature validation failure due to different transaction sig bytes
	_, err = f.TxSign(keyFoo, modSignedTxFile.Name(), "--validate-signatures")
	require.Error(t, err)
}

func TestLFBSendGenerateSignAndBroadcast(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	// Test generate sendTx with default gas
	sendTokens := sdk.TokensFromConsensusPower(10)
	out, err := f.TxSend(fooAddr.String(), barAddr, sdk.NewCoin(denom, sendTokens), "--generate-only")
	require.NoError(t, err)
	msg := UnmarshalTx(t, out.Bytes())
	require.Equal(t, msg.GetAuthInfo().GetFee().GetGasLimit(), uint64(flags.DefaultGasLimit))
	require.Equal(t, len(msg.GetMsgs()), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx with --gas=$amount
	out, err = f.TxSend(fooAddr.String(), barAddr, sdk.NewCoin(denom, sendTokens), "--gas=100", "--generate-only")
	require.NoError(t, err)
	msg = UnmarshalTx(t, out.Bytes())
	require.Equal(t, msg.GetAuthInfo().GetFee().GetGasLimit(), uint64(100))
	require.Equal(t, len(msg.GetMsgs()), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx, estimate gas
	out, err = f.TxSend(fooAddr.String(), barAddr, sdk.NewCoin(denom, sendTokens), "--generate-only")
	require.NoError(t, err)
	msg = UnmarshalTx(t, out.Bytes())
	require.True(t, msg.GetAuthInfo().GetFee().GetGasLimit() > 0)
	require.Equal(t, len(msg.GetMsgs()), 1)

	// Write the output to disk
	unsignedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(unsignedTxFile.Name())

	// Test sign
	out, err = f.TxSign(keyFoo, unsignedTxFile.Name())
	require.NoError(t, err)
	msg = UnmarshalTx(t, out.Bytes())
	require.Equal(t, len(msg.GetMsgs()), 1)
	require.Equal(t, 1, len(msg.GetSignatures()))
	require.Equal(t, fooAddr.String(), msg.GetSigners()[0].String())

	// Write the output to disk
	signedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(signedTxFile.Name())

	// Ensure foo has right amount of funds
	fooBal := f.QueryBalances(fooAddr)
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(denom))

	// Test broadcast
	_, err = f.TxBroadcast(signedTxFile.Name())
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure account state
	barBal := f.QueryBalances(barAddr)
	fooBal = f.QueryBalances(fooAddr)
	require.Equal(t, sendTokens, barBal.GetBalances().AmountOf(denom))
	require.Equal(t, startTokens.Sub(sendTokens), fooBal.GetBalances().AmountOf(denom))
}

func TestLFBMultisignInsufficientCosigners(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server with minimum fees
	n := f.LFBStart("")
	defer n.Cleanup()

	fooBarBazAddr := f.KeyAddress(keyFooBarBaz)
	barAddr := f.KeyAddress(keyBar)

	// Send some tokens from one account to the other
	_, err := f.TxSend(keyFoo, fooBarBazAddr, sdk.NewInt64Coin(denom, 10), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Test generate sendTx with multisig
	out, err := f.TxSend(fooBarBazAddr.String(), barAddr, sdk.NewInt64Coin(denom, 5), "--generate-only")
	require.NoError(t, err)

	// Write the output to disk
	unsignedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	out, err = f.TxSign(keyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.NoError(t, err)

	// Write the output to disk
	fooSignatureFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(fooSignatureFile.Name())

	// Multisign, not enough signatures
	out, err = f.TxMultisign(unsignedTxFile.Name(), keyFooBarBaz, []string{fooSignatureFile.Name()})
	require.NoError(t, err)

	// Write the output to disk
	signedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(signedTxFile.Name())

	// Validate the multisignature
	_, err = f.TxSign(keyFooBarBaz, signedTxFile.Name(), "--validate-signatures")
	require.Error(t, err)

	// Broadcast the transaction
	out, err = f.TxBroadcast(signedTxFile.Name())
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Contains(t, out.String(), "signature verification failed")
}

func TestLFBEncode(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	cdc, _ := app.MakeCodecs()

	// Build a testing transaction and write it to disk
	barAddr := f.KeyAddress(keyBar)
	keyAddr := f.KeyAddress(keyFoo)

	sendTokens := sdk.TokensFromConsensusPower(10)
	out, err := f.TxSend(keyAddr.String(), barAddr, sdk.NewCoin(denom, sendTokens), "--generate-only", "--memo", "deadbeef")
	require.NoError(t, err)

	// Write it to disk
	jsonTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(jsonTxFile.Name())

	// Run the encode command, and trim the extras from the stdout capture
	base64Encoded, err := f.TxEncode(jsonTxFile.Name())
	require.NoError(t, err)
	trimmedBase64 := strings.Trim(base64Encoded.String(), "\"\n")
	// Decode the base64
	decodedBytes, err := base64.StdEncoding.DecodeString(trimmedBase64)
	require.Nil(t, err)

	// Check that the transaction decodes as expected
	var decodedTx tx.Tx
	require.Nil(t, cdc.UnmarshalBinaryBare(decodedBytes, &decodedTx))
	require.Equal(t, "deadbeef", decodedTx.GetBody().GetMemo())
}

func TestLFBMultisignSortSignatures(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server with minimum fees
	n := f.LFBStart("")
	defer n.Cleanup()

	fooBarBazAddr := f.KeyAddress(keyFooBarBaz)
	barAddr := f.KeyAddress(keyBar)

	// Send some tokens from one account to the other
	_, err := f.TxSend(keyFoo, fooBarBazAddr, sdk.NewInt64Coin(denom, 10), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure account balances match expected
	fooBarBazBal := f.QueryBalances(fooBarBazAddr)
	require.Equal(t, int64(10), fooBarBazBal.GetBalances().AmountOf(denom).Int64())

	// Test generate sendTx with multisig
	out, err := f.TxSend(fooBarBazAddr.String(), barAddr, sdk.NewInt64Coin(denom, 5), "--generate-only")
	require.NoError(t, err)

	// Write the output to disk
	unsignedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	out, err = f.TxSign(keyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.NoError(t, err)

	// Write the output to disk
	fooSignatureFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(fooSignatureFile.Name())

	// Sign with baz's key
	out, err = f.TxSign(keyBaz, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.NoError(t, err)

	// Write the output to disk
	bazSignatureFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(bazSignatureFile.Name())

	// Multisign, keys in different order
	out, err = f.TxMultisign(unsignedTxFile.Name(), keyFooBarBaz, []string{
		bazSignatureFile.Name(), fooSignatureFile.Name()})
	require.NoError(t, err)

	// Write the output to disk
	signedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(signedTxFile.Name())

	// Broadcast the transaction
	_, err = f.TxBroadcast(signedTxFile.Name())
	require.NoError(t, err)
}

func TestLFBMultisign(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server with minimum fees
	n := f.LFBStart("")
	defer n.Cleanup()

	fooBarBazAddr := f.KeyAddress(keyFooBarBaz)
	bazAddr := f.KeyAddress(keyBaz)

	// Send some tokens from one account to the other
	_, err := f.TxSend(keyFoo, fooBarBazAddr, sdk.NewInt64Coin(denom, 10), "-y")
	require.NoError(t, err)

	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// Ensure account balances match expected
	fooBarBazBal := f.QueryBalances(fooBarBazAddr)
	require.Equal(t, int64(10), fooBarBazBal.GetBalances().AmountOf(denom).Int64())

	// Test generate sendTx with multisig
	out, err := f.TxSend(fooBarBazAddr.String(), bazAddr, sdk.NewInt64Coin(denom, 10), "--generate-only")
	require.NoError(t, err)

	// Write the output to disk
	unsignedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	out, err = f.TxSign(keyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.NoError(t, err)

	// Write the output to disk
	fooSignatureFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(fooSignatureFile.Name())

	// Sign with bar's key
	out, err = f.TxSign(keyBar, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.NoError(t, err)

	// Write the output to disk
	barSignatureFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(barSignatureFile.Name())

	// Multisign
	out, err = f.TxMultisign(unsignedTxFile.Name(), keyFooBarBaz, []string{
		fooSignatureFile.Name(), barSignatureFile.Name()})
	require.NoError(t, err)

	// Write the output to disk
	signedTxFile := WriteToNewTempFile(t, out.String())
	defer os.Remove(signedTxFile.Name())

	// Broadcast the transaction
	_, err = f.TxBroadcast(signedTxFile.Name())
	require.NoError(t, err)
}

func TestLFBCollectGentxs(t *testing.T) {
	t.Parallel()
	var customMaxBytes, customMaxGas int64 = 99999999, 1234567
	f := NewFixtures(t, getHomeDir(t))

	// Initialize temporary directories
	gentxDir, err := ioutil.TempDir("", "")
	gentxDoc := filepath.Join(gentxDir, "gentx.json")
	require.NoError(t, err)

	defer f.Cleanup(gentxDir)

	// Initialize keys
	f.KeysAdd(keyFoo)

	// Run init
	f.LFBInit(keyFoo)

	// Customize genesis.json
	genFile := f.GenesisFile()
	genDoc, err := osttypes.GenesisDocFromFile(genFile)
	require.NoError(t, err)
	genDoc.ConsensusParams.Block.MaxBytes = customMaxBytes
	genDoc.ConsensusParams.Block.MaxGas = customMaxGas
	err = genDoc.SaveAs(genFile)
	require.NoError(t, err)

	// Add account to genesis.json
	f.AddGenesisAccount(f.KeyAddress(keyFoo), startCoins)

	// Write gentx file
	f.GenTx(keyFoo, fmt.Sprintf("--output-document=%s", gentxDoc))

	// Collect gentxs from a custom directory
	f.CollectGenTxs(fmt.Sprintf("--gentx-dir=%s", gentxDir))

	genDoc, err = osttypes.GenesisDocFromFile(genFile)
	require.NoError(t, err)
	require.Equal(t, genDoc.ConsensusParams.Block.MaxBytes, customMaxBytes)
	require.Equal(t, genDoc.ConsensusParams.Block.MaxGas, customMaxGas)
}

func TestValidateGenesis(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	f.ValidateGenesis(filepath.Join(f.Home, "config", "genesis.json"))
}

func TestLFBIncrementSequenceDecorator(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server
	n := f.LFBStart("")
	defer n.Cleanup()

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	sendTokens := sdk.TokensFromConsensusPower(1)

	time.Sleep(3 * time.Second)

	fooAcc := f.QueryAccount(fooAddr)

	// Prepare signed Tx
	signedTxFiles := make([]*os.File, 0)
	for idx := 0; idx < 3; idx++ {
		// Test generate sendTx, estimate gas
		out, err := f.TxSend(fooAddr.String(), barAddr, sdk.NewCoin(denom, sendTokens), "--generate-only")
		require.NoError(t, err)

		// Write the output to disk
		unsignedTxFile := WriteToNewTempFile(t, out.String())
		defer os.Remove(unsignedTxFile.Name())

		// Test sign
		out, err = f.TxSign(keyFoo, unsignedTxFile.Name(), "--offline", "--sig-block-height", strconv.Itoa(1), "--sequence", strconv.Itoa(int(fooAcc.Sequence)+idx))
		require.NoError(t, err)

		// Write the output to disk
		signedTxFile := WriteToNewTempFile(t, out.String())
		signedTxFiles = append(signedTxFiles, signedTxFile)
		defer os.Remove(signedTxFile.Name())
	}
	// Wait for a new block
	err := n.WaitForNextBlock()
	require.NoError(t, err)

	txHashes := make([]string, 0)
	// Broadcast the signed Txs
	for _, signedTxFile := range signedTxFiles {
		// Test broadcast
		out, err := f.TxBroadcast(signedTxFile.Name(), "--broadcast-mode", "sync")
		require.NoError(t, err)
		sendResp := UnmarshalTxResponse(t, out.Bytes())
		txHashes = append(txHashes, sendResp.TxHash)
	}

	// Wait for a new block
	err = n.WaitForNextBlock()
	require.NoError(t, err)

	// All Txs are in one block
	height := f.QueryTx(txHashes[0]).Height
	for _, txHash := range txHashes {
		require.Equal(t, height, f.QueryTx(txHash).Height)
	}
}

func TestLFBWasmContract(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	defer f.Cleanup()

	// start lfb server with minimum fees
	n := f.LFBStart("")
	defer n.Cleanup()

	fooAddr := f.KeyAddress(keyFoo)

	flagFromFoo := fmt.Sprintf("--from=%s", fooAddr)
	flagGas := "--gas=auto"
	flagGasAdjustment := "--gas-adjustment=1.2"
	workDir, _ := os.Getwd()
	tmpDir := path.Join(workDir, "tmp-dir-for-test-queue")
	dirContract := path.Join(workDir, "contracts", "queue")
	hashFile := path.Join(dirContract, "hash.txt")
	wasmQueue := path.Join(dirContract, "contract.wasm")
	codeID := uint64(1)
	amountSend := uint64(10)
	denomSend := fooDenom

	var contractAddress string
	count := 0
	initValue := 0
	enqueueValue := 2

	// make tmpDir
	os.Mkdir(tmpDir, os.ModePerm)

	// validate that there are no code in the chain
	{
		listCode := f.QueryListCodeWasm()
		require.Len(t, listCode.CodeInfos, 0)
	}

	// store the contract queue
	{
		_, err := f.TxStoreWasm(wasmQueue, flagFromFoo, flagGasAdjustment, flagGas, "-y")
		require.NoError(t, err)
		// Wait for a new block
		err = n.WaitForNextBlock()
		require.NoError(t, err)
	}

	// validate the code is stored
	{
		queryCodesResponse := f.QueryListCodeWasm()
		require.Len(t, queryCodesResponse.CodeInfos, 1)

		//validate the hash is the same
		expectedRow, _ := ioutil.ReadFile(hashFile)
		expected, err := hex.DecodeString(string(expectedRow[:64]))
		require.NoError(t, err)
		actual := queryCodesResponse.CodeInfos[0].DataHash.Bytes()
		require.Equal(t, expected, actual)
	}

	// validate getCode get the exact same wasm
	{
		outputPath := path.Join(tmpDir, "queue-tmp.wasm")
		f.QueryCodeWasm(codeID, outputPath)
		fLocal, _ := os.Open(wasmQueue)
		fChain, _ := os.Open(outputPath)

		// 2000000 is enough length
		dataLocal := make([]byte, 2000000)
		dataChain := make([]byte, 2000000)
		fLocal.Read(dataLocal)
		fChain.Read(dataChain)
		require.Equal(t, dataLocal, dataChain)
	}

	// validate that there are no contract using the code (id=1)
	{
		listContract := f.QueryListContractByCodeWasm(codeID)
		require.Len(t, listContract.Contracts, 0)
	}

	// instantiate a contract with the code queue
	{
		msgJSON := fmt.Sprintf("{}")
		flagLabel := "--label=queue-test"
		flagAmount := fmt.Sprintf("--amount=%d%s", amountSend, denomSend)
		_, err := f.TxInstantiateWasm(codeID, msgJSON, flagFromFoo, flagGasAdjustment, flagGas, flagLabel, flagAmount, flagFromFoo, "-y")
		require.NoError(t, err)
		// Wait for a new block
		err = n.WaitForNextBlock()
		require.NoError(t, err)
	}

	// validate there is only one contract using codeID=1 and get contractAddress
	{
		listContract := f.QueryListContractByCodeWasm(codeID)
		require.Len(t, listContract.Contracts, 1)
		contractAddress = listContract.Contracts[0]
	}

	// check queue count and sum
	{
		res := f.QueryContractStateSmartWasm(contractAddress, "{\"count\":{}}")
		require.Equal(t, fmt.Sprintf("{\"data\":{\"count\":%d}}", count), strings.TrimRight(res, "\n"))

		res = f.QueryContractStateSmartWasm(contractAddress, "{\"sum\":{}}")
		require.Equal(t, fmt.Sprintf("{\"data\":{\"sum\":%d}}", initValue), strings.TrimRight(res, "\n"))
	}

	// execute contract(enqueue function)
	{
		msgJSON := fmt.Sprintf("{\"enqueue\":{\"value\":%d}}", enqueueValue)
		_, err := f.TxExecuteWasm(contractAddress, msgJSON, flagFromFoo, flagGasAdjustment, flagGas, "-y")
		require.NoError(t, err)
		// Wait for a new block
		err = n.WaitForNextBlock()
		require.NoError(t, err)
		count++
	}

	// check queue count and sum
	{
		res := f.QueryContractStateSmartWasm(contractAddress, "{\"count\":{}}")
		require.Equal(t, fmt.Sprintf("{\"data\":{\"count\":%d}}", count), strings.TrimRight(res, "\n"))

		res = f.QueryContractStateSmartWasm(contractAddress, "{\"sum\":{}}")
		require.Equal(t, fmt.Sprintf("{\"data\":{\"sum\":%d}}", initValue+enqueueValue), strings.TrimRight(res, "\n"))
	}

	// remove tmp dir
	os.RemoveAll(tmpDir)
}
