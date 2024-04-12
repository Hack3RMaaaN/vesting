package ante_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/vesting/testutil"
	"github.com/evmos/vesting/x/vesting/types"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestVestingDelegationDecorator(t *testing.T) {
	var (
		priv, _, addr    = testdata.KeyTestPubAddr()
		_, _, funderAddr = testdata.KeyTestPubAddr()
		coins            = sdk.NewCoins(sdk.NewCoin(testutil.StakeDenom, math.NewInt(70)))
		// instanciate a base and clawback vesting account to use in the tests
		now     = time.Now()
		baseAcc = authtypes.NewBaseAccountWithAddress(addr)
		vestAcc = types.NewClawbackVestingAccount(baseAcc, funderAddr, testutil.OrigCoins, now, testutil.LockupPeriods, testutil.VestingPeriods)
		sendMsg = &banktypes.MsgSend{
			FromAddress: addr.String(),
			ToAddress:   funderAddr.String(),
			Amount:      coins,
		}
		delMsg = &stakingtypes.MsgDelegate{
			DelegatorAddress: addr.String(),
			ValidatorAddress: sdk.ValAddress(funderAddr).String(),
			Amount:           coins[0],
		}
		// validator pub key
		pubKey      = ed25519.GenPrivKey().PubKey()
		commissions = stakingtypes.NewCommissionRates(
			sdk.NewDecWithPrec(5, 2),
			sdk.NewDecWithPrec(2, 1),
			sdk.NewDecWithPrec(5, 2),
		)
		createValMsg, err = stakingtypes.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			pubKey,
			coins[0],
			stakingtypes.NewDescription("T", "E", "S", "T", "Z"),
			commissions,
			sdk.OneInt(),
		)
	)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		msg       sdk.Msg
		malleate  func(suite *AnteTestSuite)
		expPass   bool
		expErrMsg string
	}{
		{
			name:    "MsgSend - no-op",
			msg:     sendMsg,
			expPass: true,
		},
		{
			name: "MsgDelegate with non-existent account as Delegator - should fail",
			msg:  delMsg,
			malleate: func(suite *AnteTestSuite) {
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return nil.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(nil)
			},
			expPass:   false,
			expErrMsg: fmt.Sprintf("account %s does not exist", addr),
		},
		{
			name: "MsgCreateValidator with non-existent account as Validator - should fail",
			msg:  createValMsg,
			malleate: func(suite *AnteTestSuite) {
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return nil.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(nil)
			},
			expPass:   false,
			expErrMsg: fmt.Sprintf("account %s does not exist", addr),
		},
		{
			name: "MsgDelegate with normal account as Delegator - no-op",
			msg:  delMsg,
			malleate: func(suite *AnteTestSuite) {
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return the base account defined in 'baseAcc' var.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(baseAcc)
			},
			expPass: true,
		},
		{
			name: "MsgCreateValidator with normal account as Delegator - no-op",
			msg:  createValMsg,
			malleate: func(suite *AnteTestSuite) {
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return the base account defined in 'baseAcc' var.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(baseAcc)
			},
			expPass: true,
		},
		{
			name: "MsgDelegate with clawback account without vested coins as Delegator - should fail",
			msg:  delMsg,
			malleate: func(suite *AnteTestSuite) {
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return the clawback vesting account defined in 'vestAcc' var.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(vestAcc)
			},
			expPass:   false,
			expErrMsg: "account has no vested coins",
		},
		{
			name: "MsgCreateValidator with clawback account without vested coins as Delegator - should fail",
			msg:  createValMsg,
			malleate: func(suite *AnteTestSuite) {
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return the clawback vesting account defined in 'vestAcc' var.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(vestAcc)
			},
			expPass:   false,
			expErrMsg: "account has no vested coins",
		},
		{
			name: "MsgDelegate with clawback account with vested coins < delegation amount - should fail",
			msg:  delMsg,
			malleate: func(suite *AnteTestSuite) {
				// 50 percent of coins are vested after 1st vesting period,
				// but before unlocking (all locked coins)
				suite.ctx = suite.ctx.WithBlockTime(now.Add(12 * time.Hour))
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return the clawback vesting account defined in 'vestAcc' var.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(vestAcc)
			},
			expPass:   false,
			expErrMsg: "cannot delegate unvested coins",
		},
		{
			name: "MsgCreateValidator with clawback account with vested coins < delegation amount - should fail",
			msg:  createValMsg,
			malleate: func(suite *AnteTestSuite) {
				// 50 percent of coins are vested after 1st vesting period,
				// but before unlocking (all locked coins)
				suite.ctx = suite.ctx.WithBlockTime(now.Add(12 * time.Hour))
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return the clawback vesting account defined in 'vestAcc' var.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(vestAcc)
			},
			expPass:   false,
			expErrMsg: "cannot delegate unvested coins",
		},
		{
			name: "MsgDelegate with clawback account with free coins and vested tokens",
			msg: &stakingtypes.MsgDelegate{
				DelegatorAddress: addr.String(),
				ValidatorAddress: sdk.ValAddress(funderAddr).String(),
				Amount:           sdk.NewCoin(testutil.StakeDenom, math.NewInt(60)), // 10 free coins + 50 locked vested coins
			},
			malleate: func(suite *AnteTestSuite) {
				// 50 percent of coins are vested after 1st vesting period,
				// but before unlocking (all locked coins)
				suite.ctx = suite.ctx.WithBlockTime(now.Add(12 * time.Hour))
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return the clawback vesting account defined in 'vestAcc' var.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(vestAcc)
			},
			expPass: true,
		},
		{
			name: "MsgCreateValidator with clawback account with free coins and vested tokens",
			msg: &stakingtypes.MsgCreateValidator{
				Description:       stakingtypes.NewDescription("T", "E", "S", "T", "Z"),
				DelegatorAddress:  addr.String(),
				ValidatorAddress:  sdk.ValAddress(addr).String(),
				Pubkey:            baseAcc.PubKey,
				Value:             sdk.NewCoin(testutil.StakeDenom, math.NewInt(60)), // 10 free coins + 50 locked vested coins,
				Commission:        commissions,
				MinSelfDelegation: sdk.OneInt(),
			},
			malleate: func(suite *AnteTestSuite) {
				// 50 percent of coins are vested after 1st vesting period,
				// but before unlocking (all locked coins)
				suite.ctx = suite.ctx.WithBlockTime(now.Add(12 * time.Hour))
				// Asserts that the first and only call to GetAccount() with the 'addr' address
				// will return the clawback vesting account defined in 'vestAcc' var.
				suite.accountKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(vestAcc)
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			suite := setupTestSuite(t)

			// set mock to return account balance to be the vesting coins + a few 'stake' coins
			freeAmt := math.NewInt(10)
			totalAmt := testutil.OrigCoins.AmountOf(testutil.StakeDenom).Add(freeAmt)
			suite.bankKeeper.EXPECT().GetBalance(gomock.Any(), addr, testutil.StakeDenom).Return(sdk.NewCoin(testutil.StakeDenom, totalAmt)).AnyTimes()

			require.NoError(t, suite.txBuilder.SetMsgs(tc.msg))
			suite.txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
			suite.txBuilder.SetGasLimit(testdata.NewTestGasLimit())

			tx, err := suite.createTestTx(priv, 0, 0, suite.ctx.ChainID())
			require.NoError(t, err)

			if tc.malleate != nil {
				tc.malleate(suite)
			}
			newCtx, err := suite.dec.AnteHandle(suite.ctx, tx, false, nextFn)

			if tc.expPass {
				require.NoError(t, err)
				require.NotNil(t, newCtx)

				suite.ctx = newCtx
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expErrMsg)
		})
	}
}
