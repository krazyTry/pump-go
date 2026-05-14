package math

import (
	"time"

	pump "github.com/krazyTry/pump-go/gen/pump"
)

func TotalUnclaimedTokens(g *pump.GlobalVolumeAccumulator, u *pump.UserVolumeAccumulator, now time.Time) uint64 {
	res := u.TotalUnclaimedTokens
	if g.StartTime == 0 || g.EndTime == 0 || g.SecondsInADay == 0 {
		return res
	}
	cur := now.Unix()
	if cur < g.StartTime || u.LastUpdateTimestamp < g.StartTime || g.EndTime < g.StartTime {
		return res
	}
	curIdx := uint64((cur - g.StartTime) / g.SecondsInADay)
	lastIdx := uint64((u.LastUpdateTimestamp - g.StartTime) / g.SecondsInADay)
	endIdx := uint64((g.EndTime - g.StartTime) / g.SecondsInADay)
	if curIdx > lastIdx && lastIdx <= endIdx {
		sv := g.SolVolumes[lastIdx]
		if sv == 0 {
			return res
		}
		return res + (u.CurrentSolVolume*g.TotalTokenSupply[lastIdx])/sv
	}
	return res
}

func CurrentDayTokens(g *pump.GlobalVolumeAccumulator, u *pump.UserVolumeAccumulator, now time.Time) uint64 {
	if g.StartTime == 0 || g.EndTime == 0 || g.SecondsInADay == 0 {
		return 0
	}
	cur := now.Unix()
	if cur < g.StartTime || cur > g.EndTime || u.LastUpdateTimestamp < g.StartTime || g.EndTime < g.StartTime {
		return 0
	}
	curIdx := uint64((cur - g.StartTime) / g.SecondsInADay)
	lastIdx := uint64((u.LastUpdateTimestamp - g.StartTime) / g.SecondsInADay)
	if curIdx != lastIdx {
		return 0
	}
	sv := g.SolVolumes[curIdx]
	if sv == 0 {
		return 0
	}
	return (u.CurrentSolVolume * g.TotalTokenSupply[curIdx]) / sv
}
