package math

import (
	amm "github.com/krazyTry/pump-go/amm"
)

func TotalUnclaimedTokens(g *amm.GlobalVolumeAccumulator, u *amm.UserVolumeAccumulator, currentTimestamp int64) uint64 {
	if g.StartTime == 0 || g.EndTime == 0 || g.SecondsInADay == 0 {
		return u.TotalUnclaimedTokens
	}

	if currentTimestamp < g.StartTime || u.LastUpdateTimestamp < g.StartTime || g.EndTime < g.StartTime {
		return u.TotalUnclaimedTokens
	}

	currentDay := currentTimestamp - g.StartTime

	currentDay = currentDay / g.SecondsInADay

	lastDay := u.LastUpdateTimestamp - g.StartTime

	lastDay = lastDay / g.SecondsInADay

	endDay := g.EndTime - g.StartTime

	endDay = endDay / g.SecondsInADay

	if currentDay > lastDay && lastDay <= endDay {
		i := int(lastDay)
		if g.SolVolumes[i] == 0 {
			return u.TotalUnclaimedTokens
		}
		return u.TotalUnclaimedTokens + u.CurrentSolVolume*g.TotalTokenSupply[i]/g.SolVolumes[i]
	}
	return u.TotalUnclaimedTokens
}

func CurrentDayTokens(g *amm.GlobalVolumeAccumulator, u *amm.UserVolumeAccumulator, currentTimestamp int64) uint64 {

	if g.StartTime == 0 || g.EndTime == 0 || g.SecondsInADay == 0 {
		return 0
	}

	if currentTimestamp < g.StartTime || currentTimestamp > g.EndTime || u.LastUpdateTimestamp < g.StartTime || g.EndTime < g.StartTime {
		return 0
	}

	currentDay := currentTimestamp - g.StartTime
	currentDay = currentDay / g.SecondsInADay

	lastDay := u.LastUpdateTimestamp - g.StartTime
	lastDay = lastDay / g.SecondsInADay
	if currentDay != lastDay {
		return 0
	}
	i := int(currentDay)
	if g.SolVolumes[i] == 0 {
		return 0
	}
	return u.CurrentSolVolume * g.TotalTokenSupply[i] / g.SolVolumes[i]
}
