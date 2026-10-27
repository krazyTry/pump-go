package math

import amm "github.com/krazyTry/pump-go/amm"

// TotalUnclaimedTokens is the user's unclaimed incentive tokens at currentTimestamp.
func TotalUnclaimedTokens(g *amm.GlobalVolumeAccumulator, u *amm.UserVolumeAccumulator, currentTimestamp int64) uint64 {
	return amm.TotalUnclaimedTokens(g, u, currentTimestamp)
}

// CurrentDayTokens is the user's incentive share for the day containing currentTimestamp.
func CurrentDayTokens(g *amm.GlobalVolumeAccumulator, u *amm.UserVolumeAccumulator, currentTimestamp int64) uint64 {
	return amm.CurrentDayTokens(g, u, currentTimestamp)
}
