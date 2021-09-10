// +build cli_multi_node_test

package clitest

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptocodec "github.com/line/lbm-sdk/crypto/codec"
	sdk "github.com/line/lbm-sdk/types"

	"github.com/line/ostracon/privval"
)

func TestMultiValidatorAndSendTokens(t *testing.T) {
	t.Parallel()

	fg := InitFixturesGroup(t)

	fg.LFBStartCluster()
	defer fg.Cleanup()

	f := fg.Fixture(0)

	var (
		keyFoo = f.Moniker
	)

	fooAddr := f.KeyAddress(keyFoo)
	f.KeysDelete(keyBaz)
	f.KeysAdd(keyBaz)
	bazAddr := f.KeyAddress(keyBaz)

	fg.AddFullNode()

	require.NoError(t, fg.Network.WaitForNextBlock())
	{
		fooBal := f.QueryBalances(fooAddr)
		startTokens := sdk.TokensFromConsensusPower(50)
		require.Equal(t, startTokens, fooBal.GetBalances().AmountOf(denom))

		// Send some tokens from one account to the other
		sendTokens := sdk.TokensFromConsensusPower(10)
		_, err := f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())

		// Ensure account balances match expected
		barBal := f.QueryBalances(bazAddr)
		require.Equal(t, sendTokens, barBal.GetBalances().AmountOf(denom))
		fooBal = f.QueryBalances(fooAddr)
		require.Equal(t, startTokens.Sub(sendTokens), fooBal.GetBalances().AmountOf(denom))

		// Test --dry-run
		_, err = f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), "--dry-run")
		require.NoError(t, err)

		// Test --generate-only
		out, err := f.TxSend(
			fooAddr.String(), bazAddr, sdk.NewCoin(denom, sendTokens), "--generate-only=true",
		)
		require.NoError(t, err)
		msg := UnmarshalTx(f.T, out.Bytes())
		require.NotZero(t, msg.AuthInfo.GetFee().GetGasLimit())
		require.Len(t, msg.GetMsgs(), 1)
		require.Len(t, msg.GetSignatures(), 0)

		// Check state didn't change
		fooBal = f.QueryBalances(fooAddr)
		require.Equal(t, startTokens.Sub(sendTokens), fooBal.GetBalances().AmountOf(denom))

		// test autosequencing
		_, err = f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())

		// Ensure account balances match expected
		barBal = f.QueryBalances(bazAddr)
		require.Equal(t, sendTokens.MulRaw(2), barBal.GetBalances().AmountOf(denom))
		fooBal = f.QueryBalances(fooAddr)
		require.Equal(t, startTokens.Sub(sendTokens.MulRaw(2)), fooBal.GetBalances().AmountOf(denom))

		// test memo
		_, err = f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), "--memo='testmemo'", "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())

		// Ensure account balances match expected
		barBal = f.QueryBalances(bazAddr)
		require.Equal(t, sendTokens.MulRaw(3), barBal.GetBalances().AmountOf(denom))
		fooBal = f.QueryBalances(fooAddr)
		require.Equal(t, startTokens.Sub(sendTokens.MulRaw(3)), fooBal.GetBalances().AmountOf(denom))
	}
}

func TestMultiValidatorAddNodeAndPromoteValidator(t *testing.T) {
	t.Parallel()

	fg := InitFixturesGroup(t)
	fg.LFBStartCluster()
	defer fg.Cleanup()

	f1 := fg.Fixture(0)

	f2 := fg.AddFullNode()

	{
		f2.KeysDelete(keyBar)
		f2.KeysAdd(keyBar)
	}

	barAddr := f2.KeyAddress(keyBar)
	barVal := barAddr.ToValAddress()

	sendTokens := sdk.TokensFromConsensusPower(10)
	{
		_, err := f1.TxSend(f1.Moniker, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())

		barBal := f2.QueryBalances(barAddr)
		require.Equal(t, sendTokens, barBal.GetBalances().AmountOf(denom))
	}

	newValTokens := sdk.TokensFromConsensusPower(2)
	{
		privVal := privval.LoadFilePVEmptyState(f2.PrivValidatorKeyFile(), "")
		pubkey, err := privVal.GetPubKey()
		require.NoError(t, err)

		tmValPubKey, err := cryptocodec.FromOcPubKeyInterface(pubkey)
		require.NoError(t, err)
		consPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, tmValPubKey)

		_, err = f2.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewCoin(denom, newValTokens), "-y")
		require.NoError(t, err)
		require.NoError(t, fg.Network.WaitForNextBlock())
	}
	{
		// Ensure funds were deducted properly
		barBal := f2.QueryBalances(barAddr)
		require.Equal(t, sendTokens.Sub(newValTokens), barBal.GetBalances().AmountOf(denom))

		// Ensure that validator state is as expected
		validator := f2.QueryStakingValidator(barVal)
		require.Equal(t, validator.OperatorAddress, barVal.String())
		require.True(sdk.IntEq(t, newValTokens, validator.Tokens))

		// Query delegations to the validator
		validatorDelegations := f2.QueryStakingDelegationsTo(barVal)
		require.Len(t, validatorDelegations.DelegationResponses, 1)
		require.NotZero(t, validatorDelegations.DelegationResponses[0].Delegation.GetShares())
	}
}