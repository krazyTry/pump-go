package amm

import "math/big"

// floorMulDiv is currentSolVolume * tokenSupply / solVolume with floor division.
// The product is computed with big.Int so a quotient that still fits in uint64 is not lost when the product exceeds 2^64-1.
func volumeMulDiv(volume, supply, solVolume uint64) uint64 {
	n := new(big.Int).Mul(new(big.Int).SetUint64(volume), new(big.Int).SetUint64(supply))
	n.Quo(n, new(big.Int).SetUint64(solVolume))
	if !n.IsUint64() {
		return ^uint64(0)
	}
	return n.Uint64()
}

// TotalUnclaimedTokens is the user's unclaimed incentive tokens at currentTimestamp.
func TotalUnclaimedTokens(g *GlobalVolumeAccumulator, u *UserVolumeAccumulator, currentTimestamp int64) uint64 {
	if g.StartTime == 0 || g.EndTime == 0 || g.SecondsInADay == 0 {
		return u.TotalUnclaimedTokens
	}
	if currentTimestamp < g.StartTime || u.LastUpdateTimestamp < g.StartTime || g.EndTime < g.StartTime {
		return u.TotalUnclaimedTokens
	}
	currentDay := (currentTimestamp - g.StartTime) / g.SecondsInADay
	lastDay := (u.LastUpdateTimestamp - g.StartTime) / g.SecondsInADay
	endDay := (g.EndTime - g.StartTime) / g.SecondsInADay
	if currentDay > lastDay && lastDay <= endDay {
		i := int(lastDay)
		if i < 0 || i >= len(g.SolVolumes) || g.SolVolumes[i] == 0 {
			return u.TotalUnclaimedTokens
		}
		return u.TotalUnclaimedTokens + volumeMulDiv(u.CurrentSolVolume, g.TotalTokenSupply[i], g.SolVolumes[i])
	}
	return u.TotalUnclaimedTokens
}

// CurrentDayTokens is the user's incentive share for the day containing currentTimestamp.
func CurrentDayTokens(g *GlobalVolumeAccumulator, u *UserVolumeAccumulator, currentTimestamp int64) uint64 {
	if g.StartTime == 0 || g.EndTime == 0 || g.SecondsInADay == 0 {
		return 0
	}
	if currentTimestamp < g.StartTime || currentTimestamp > g.EndTime || u.LastUpdateTimestamp < g.StartTime || g.EndTime < g.StartTime {
		return 0
	}
	currentDay := (currentTimestamp - g.StartTime) / g.SecondsInADay
	lastDay := (u.LastUpdateTimestamp - g.StartTime) / g.SecondsInADay
	if currentDay != lastDay {
		return 0
	}
	i := int(currentDay)
	if i < 0 || i >= len(g.SolVolumes) || g.SolVolumes[i] == 0 {
		return 0
	}
	return volumeMulDiv(u.CurrentSolVolume, g.TotalTokenSupply[i], g.SolVolumes[i])
}
