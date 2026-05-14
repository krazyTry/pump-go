package bonding_curve

import pump "github.com/krazyTry/pump-go/gen/pump"

type Global = pump.Global
type FeeConfig = pump.FeeConfig
type BondingCurve = pump.BondingCurve
type GlobalVolumeAccumulator = pump.GlobalVolumeAccumulator
type UserVolumeAccumulator = pump.UserVolumeAccumulator
type SharingConfig = pump.SharingConfig
type Shareholder = pump.Shareholder
type Fees = pump.Fees
type FeeTier = pump.FeeTier

type MinimumDistributableFeeEvent = pump.MinimumDistributableFeeEvent
type DistributeCreatorFeesEvent = pump.DistributeCreatorFeesEvent

type UserVolumeAccumulatorTotalStats struct {
	TotalUnclaimedTokens uint64
	TotalClaimedTokens   uint64
	CurrentSolVolume     uint64
}
