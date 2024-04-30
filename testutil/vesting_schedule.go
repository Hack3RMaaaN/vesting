package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	StakeDenom    = "stake"
	FeeDenom      = "fee"
	LockupPeriods = sdkvesting.Periods{
		sdkvesting.Period{
			Length: int64(16 * 60 * 60), // 16hs
			Amount: sdk.NewCoins(sdk.NewInt64Coin(FeeDenom, 1000), sdk.NewInt64Coin(StakeDenom, 100)),
		},
	}
	VestingPeriods = sdkvesting.Periods{
		sdkvesting.Period{
			Length: int64(12 * 60 * 60), // 12hs
			Amount: GetPercentOfVestingCoins(50),
		},
		sdkvesting.Period{
			Length: int64(6 * 60 * 60), // 6hs
			Amount: GetPercentOfVestingCoins(25),
		},
		sdkvesting.Period{
			Length: int64(6 * 60 * 60), // 6hs
			Amount: GetPercentOfVestingCoins(25),
		},
	}
	OrigCoins = sdk.Coins{sdk.NewInt64Coin(FeeDenom, 1000), sdk.NewInt64Coin(StakeDenom, 100)}
)

// GetPercentOfVestingCoins is a helper function to calculate
// the specified percentage of the coins in the vesting schedule
func GetPercentOfVestingCoins(percentage int64) sdk.Coins {
	if percentage < 0 || percentage > 100 {
		panic("invalid percentage passed!")
	}
	var retCoins sdk.Coins
	for _, coin := range OrigCoins {
		retCoins = retCoins.Add(sdk.NewCoin(coin.Denom, coin.Amount.MulRaw(percentage).QuoRaw(100)))
	}
	return retCoins
}
