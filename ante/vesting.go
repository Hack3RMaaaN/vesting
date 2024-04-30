// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ante

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/vesting/x/vesting/types"
)

// VestingDelegationDecorator validates delegation of vested coins
type VestingDelegationDecorator struct {
	ak  types.AccountKeeper
	sk  types.StakingKeeper
	bk  types.BankKeeper
	cdc codec.BinaryCodec
}

// NewVestingDelegationDecorator creates a new VestingDelegationDecorator
func NewVestingDelegationDecorator(ak types.AccountKeeper, sk types.StakingKeeper, bk types.BankKeeper, cdc codec.BinaryCodec) VestingDelegationDecorator {
	return VestingDelegationDecorator{
		ak:  ak,
		sk:  sk,
		bk:  bk,
		cdc: cdc,
	}
}

// AnteHandle checks if the tx contains a staking delegation.
// It errors if the coins are still locked or the bond amount is greater than
// the coins already vested
func (vdd VestingDelegationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		switch msg := msg.(type) {
		case *authz.MsgExec:
			// Check for bypassing authorization
			if err := vdd.validateAuthz(ctx, msg); err != nil {
				return ctx, err
			}
		default:
			if err := vdd.validateMsg(ctx, msg); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

// validateAuthz validates the authorization internal message
func (vdd VestingDelegationDecorator) validateAuthz(ctx sdk.Context, execMsg *authz.MsgExec) error {
	for _, v := range execMsg.Msgs {
		var innerMsg sdk.Msg
		if err := vdd.cdc.UnpackAny(v, &innerMsg); err != nil {
			return errorsmod.Wrap(err, "cannot unmarshal authz exec msgs")
		}

		if err := vdd.validateMsg(ctx, innerMsg); err != nil {
			return err
		}
	}

	return nil
}

// validateMsg checks that the only vested coins can be delegated
func (vdd VestingDelegationDecorator) validateMsg(ctx sdk.Context, msg sdk.Msg) error {
	var delegationAmt math.Int
	// need to validate delegation amount in MsgDelegate
	// and self delegation amount in MsgCreateValidator
	switch stkMsg := msg.(type) {
	case *stakingtypes.MsgDelegate:
		delegationAmt = stkMsg.Amount.Amount
	case *stakingtypes.MsgCreateValidator:
		delegationAmt = stkMsg.Value.Amount
	default:
		return nil
	}

	for _, addr := range msg.GetSigners() {
		acc := vdd.ak.GetAccount(ctx, addr)
		if acc == nil {
			return errorsmod.Wrapf(
				errortypes.ErrUnknownAddress,
				"account %s does not exist", addr,
			)
		}

		clawbackAccount, isClawback := acc.(*types.ClawbackVestingAccount)
		if !isClawback {
			// continue to next decorator as this logic only applies to vesting
			return nil
		}

		// error if bond amount is > vested coins
		bondDenom := vdd.sk.BondDenom(ctx)
		coins := clawbackAccount.GetVestedCoins(ctx.BlockTime())
		if coins == nil || coins.Empty() {
			return errorsmod.Wrap(
				types.ErrInsufficientVestedCoins,
				"account has no vested coins",
			)
		}

		balance := vdd.bk.GetBalance(ctx, addr, bondDenom)
		unvestedCoins := clawbackAccount.GetVestingCoins(ctx.BlockTime())
		// Can only delegate bondable coins
		unvestedBondableAmt := unvestedCoins.AmountOf(bondDenom)
		// A ClawbackVestingAccount can delegate coins from the vesting schedule
		// when having vested locked coins or unlocked vested coins.
		// It CANNOT delegate unvested coins
		availableAmt := balance.Amount.Sub(unvestedBondableAmt)
		if availableAmt.IsNegative() {
			availableAmt = math.ZeroInt()
		}

		if availableAmt.LT(delegationAmt) {
			return errorsmod.Wrapf(
				types.ErrInsufficientVestedCoins,
				"cannot delegate unvested coins. delegatable coins < delegation amount (%s < %s)",
				availableAmt, delegationAmt,
			)
		}
	}

	return nil
}
