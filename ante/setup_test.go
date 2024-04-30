package ante_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	gomock "go.uber.org/mock/gomock"

	"github.com/evmos/vesting/ante"
	"github.com/evmos/vesting/testutil"
	"github.com/evmos/vesting/x/vesting"
	"github.com/evmos/vesting/x/vesting/types"
)

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	dec           sdk.AnteDecorator
	ctx           sdk.Context
	clientCtx     client.Context
	txBuilder     client.TxBuilder
	accountKeeper *testutil.MockAccountKeeper
	bankKeeper    *testutil.MockBankKeeper
	stakingKeeper *testutil.MockStakingKeeper
	encCfg        moduletestutil.TestEncodingConfig
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func setupTestSuite(t *testing.T) *AnteTestSuite {
	suite := &AnteTestSuite{}
	ctrl := gomock.NewController(t)
	suite.accountKeeper = testutil.NewMockAccountKeeper(ctrl)
	suite.bankKeeper = testutil.NewMockBankKeeper(ctrl)
	suite.stakingKeeper = testutil.NewMockStakingKeeper(ctrl)

	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeight(1)
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(vesting.AppModuleBasic{})

	// We're using TestMsg encoding in some tests, so register it here.
	testdata.RegisterInterfaces(suite.encCfg.InterfaceRegistry)

	suite.clientCtx = client.Context{}.
		WithTxConfig(suite.encCfg.TxConfig)

	anteHandler := ante.NewVestingDelegationDecorator(
		suite.accountKeeper,
		suite.stakingKeeper,
		suite.bankKeeper,
		suite.encCfg.Codec,
	)

	suite.dec = anteHandler

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// Setup response for bond denom
	suite.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(testutil.StakeDenom).AnyTimes()

	return suite
}

// createTestTx is a helper function to create a tx with given inputs.
func (suite *AnteTestSuite) createTestTx(priv cryptotypes.PrivKey, accNum, accSeq uint64, chainID string) (xauthsigning.Tx, error) {
	signerData := xauthsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	sigV2, err := tx.SignWithPrivKey(
		suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
		suite.txBuilder, priv, suite.clientCtx.TxConfig, accSeq)
	if err != nil {
		return nil, err
	}

	err = suite.txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	return suite.txBuilder.GetTx(), nil
}

// nextFn is a no-op function that returns the context and no error in order to mock
// the next function in the AnteHandler chain.
//
// It can be used in unit tests when calling a decorator's AnteHandle method, e.g.
// `dec.AnteHandle(ctx, tx, false, nextFn)`
func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}
