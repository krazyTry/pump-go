package bonding_curve

import (
	"context"
	"errors"
	"time"

	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	"github.com/krazyTry/pump-go/bonding_curve/math"

	solana "github.com/gagliardetto/solana-go"
	pump "github.com/krazyTry/pump-go/gen/pump"
)

func (o *Client) FetchGlobal(ctx context.Context) (*pump.Global, error) {
	acc, err := o.RPC.GetAccountInfo(ctx, o.Global)
	if err != nil || acc == nil || acc.Value == nil {
		return nil, err
	}
	return pump.ParseAccount_Global(acc.Value.Data.GetBinary())
}

func (o *Client) FetchFeeConfig(ctx context.Context) (*pump.FeeConfig, error) {
	acc, err := o.RPC.GetAccountInfo(ctx, o.FeeConfig)
	if err != nil || acc == nil || acc.Value == nil {
		return nil, err
	}
	return pump.ParseAccount_FeeConfig(acc.Value.Data.GetBinary())
}

func (o *Client) FetchBondingCurve(ctx context.Context, baseMint solana.PublicKey) (*pump.BondingCurve, error) {
	acc, err := o.RPC.GetAccountInfo(ctx, DeriveBondingCurve(baseMint))
	if err != nil || acc == nil || acc.Value == nil {
		return nil, err
	}
	return pump.ParseAccount_BondingCurve(acc.Value.Data.GetBinary())
}

func (o *Client) FetchSharingConfig(ctx context.Context, baseMint solana.PublicKey) (*pump.SharingConfig, error) {
	acc, err := o.RPC.GetAccountInfo(ctx, DeriveFeesFeeSharingConfig(baseMint))
	if err != nil || acc == nil || acc.Value == nil {
		return nil, err
	}
	return pump.ParseAccount_SharingConfig(acc.Value.Data.GetBinary())
}

func (o *Client) FetchGlobalVolumeAccumulator(ctx context.Context) (*pump.GlobalVolumeAccumulator, error) {
	acc, err := o.RPC.GetAccountInfo(ctx, DeriveGlobalVolumeAccumulator())
	if err != nil || acc == nil || acc.Value == nil {
		return nil, err
	}
	return pump.ParseAccount_GlobalVolumeAccumulator(acc.Value.Data.GetBinary())
}

func (o *Client) FetchUserVolumeAccumulator(ctx context.Context, user solana.PublicKey) (*pump.UserVolumeAccumulator, error) {
	acc, err := o.RPC.GetAccountInfo(ctx, DeriveUserVolumeAccumulator(user))
	if err != nil || acc == nil || acc.Value == nil {
		return nil, err
	}
	return pump.ParseAccount_UserVolumeAccumulator(acc.Value.Data.GetBinary())
}

func (o *Client) FetchUserVolumeAccumulatorTotalStats(ctx context.Context, user solana.PublicKey) (*UserVolumeAccumulatorTotalStats, error) {
	u, err := o.FetchUserVolumeAccumulator(ctx, user)
	if err != nil {
		return &UserVolumeAccumulatorTotalStats{}, nil
	}
	return &UserVolumeAccumulatorTotalStats{
		TotalUnclaimedTokens: u.TotalUnclaimedTokens,
		TotalClaimedTokens:   u.TotalClaimedTokens,
		CurrentSolVolume:     u.CurrentSolVolume,
	}, nil
}

func (o *Client) GetTotalUnclaimedTokens(ctx context.Context, user solana.PublicKey) (uint64, error) {
	g, err := o.FetchGlobalVolumeAccumulator(ctx)
	if err != nil {
		return 0, err
	}
	u, err := o.FetchUserVolumeAccumulator(ctx, user)
	if err != nil {
		return 0, err
	}

	return math.TotalUnclaimedTokens(g, u, time.Now().UTC()), nil
}

func (o *Client) GetCurrentDayTokens(ctx context.Context, user solana.PublicKey) (uint64, error) {
	g, err := o.FetchGlobalVolumeAccumulator(ctx)
	if err != nil {
		return 0, err
	}
	u, err := o.FetchUserVolumeAccumulator(ctx, user)
	if err != nil {
		return 0, err
	}
	return math.CurrentDayTokens(g, u, time.Now().UTC()), nil
}

func (o *Client) BuildDistributeCreatorFeesInstructions(ctx context.Context, mint solana.PublicKey) ([]solana.Instruction, error) {
	sharing := DeriveFeesFeeSharingConfig(mint)

	ix, err := pump.NewDistributeCreatorFeesInstruction(
		mint,
		DeriveBondingCurve(mint),
		sharing,
		DeriveCreatorVault(sharing),
		solana.SystemProgramID,
		DeriveEventAuthority(),
		ProgramID,
	)
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{ix}, nil
}

func (o *Client) ClaimTokenIncentivesInstruction(ctx context.Context, user, payer solana.PublicKey) (solana.Instruction, error) {
	g, err := o.FetchGlobalVolumeAccumulator(ctx)
	if err != nil {
		return nil, err
	}
	if g.Mint.Equals(solana.PublicKey{}) {
		return nil, errors.New("incentive mint not configured")
	}
	return pump.NewClaimTokenIncentivesInstruction(
		user,
		helpers.FindAssociatedTokenAddress(user, g.Mint, solana.Token2022ProgramID),
		DeriveGlobalVolumeAccumulator(),
		helpers.FindAssociatedTokenAddress(DeriveGlobalVolumeAccumulator(), g.Mint, solana.Token2022ProgramID),
		DeriveUserVolumeAccumulator(user),
		g.Mint,
		solana.Token2022ProgramID,
		solana.SystemProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		DeriveEventAuthority(),
		ProgramID,
		payer,
	)
}

func HasCoinCreatorMigratedToSharingConfig(mint, creator solana.PublicKey) bool {
	return DeriveFeesFeeSharingConfig(mint).Equals(creator)
}

func IsSharingConfigEditable(sc *pump.SharingConfig) bool {
	if sc.Version == 1 {
		return false
	}
	if sc.Version == 2 && sc.AdminRevoked {
		return false
	}
	return true
}

func GetBuyTokenAmountFromSolAmount(global *Global, feeConfig *FeeConfig, mintSupply uint64, curve *BondingCurve, amount uint64) (uint64, error) {
	if amount == 0 {
		return 0, nil
	}

	isNew := false

	if curve == nil || mintSupply == 0 {
		curve = helpers.NewBondingCurve(global, solana.PublicKey{})
		mintSupply = global.TokenTotalSupply
		isNew = true
	}

	if curve.VirtualTokenReserves == 0 {
		return 0, nil
	}

	fees, err := helpers.ComputeFeesBps(global, feeConfig, mintSupply, curve.VirtualQuoteReserves, curve.VirtualTokenReserves)

	if err != nil {
		return 0, err
	}

	totalBps := fees.ProtocolFeeBps

	if isNew || !curve.Creator.Equals(solana.PublicKey{}) {
		totalBps += fees.CreatorFeeBps
	}

	inputAmount := ((amount - 1) * 10_000) / (10_000 + totalBps)

	t := math.GetBuyTokenAmountFromSolAmountQuote(inputAmount, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)
	if t > curve.RealTokenReserves {
		return curve.RealTokenReserves, nil
	}
	return t, nil
}

func GetBuySolAmountFromTokenAmount(global *Global, feeConfig *FeeConfig, mintSupply uint64, curve *BondingCurve, amount uint64) (uint64, error) {
	if amount == 0 {
		return 0, nil
	}

	isNew := false
	if curve == nil || mintSupply == 0 {
		curve = helpers.NewBondingCurve(global, solana.PublicKey{})
		mintSupply = global.TokenTotalSupply
		isNew = true
	}

	if curve.VirtualTokenReserves == 0 {
		return 0, nil
	}

	minAmount := min(amount, curve.RealTokenReserves)

	cost := math.GetBuySolAmountFromTokenAmountQuote(minAmount, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)

	f, err := helpers.GetFee(global, feeConfig, mintSupply, curve, cost, isNew)
	if err != nil {
		return 0, err
	}
	return cost + f, nil
}

func GetSellSolAmountFromTokenAmount(global *Global, feeConfig *FeeConfig, mintSupply uint64, curve *BondingCurve, amount uint64) (uint64, error) {
	if amount == 0 || curve.VirtualTokenReserves == 0 {
		return 0, nil
	}
	cost := math.GetSellSolAmountFromTokenAmountQuote(amount, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)

	f, err := helpers.GetFee(global, feeConfig, mintSupply, curve, cost, false)
	if err != nil {
		return 0, err
	}
	return cost - f, nil
}
