package amm

import (
	"encoding/binary"
	"fmt"

	ammgen "github.com/krazyTry/pump-go/gen/amm"
)

const (
	feeConfigFeeTiersOffset = 8 + 1 + 32 + 24
	feesSize                = 24
	feeTierSize             = 16 + feesSize
)

func padTrailing(data []byte, size int) []byte {
	if len(data) >= size {
		return data
	}
	out := make([]byte, size)
	copy(out, data)
	return out
}

func feeTierVecEnd(data []byte, offset int) (int, error) {
	if offset+4 > len(data) {
		return 0, fmt.Errorf("FeeConfig fee tier vector length at offset %d runs past the account data (%d bytes)", offset, len(data))
	}
	count := binary.LittleEndian.Uint32(data[offset : offset+4])
	end := offset + 4 + int(count)*feeTierSize
	if end > len(data) {
		return 0, fmt.Errorf("FeeConfig fee tier vector at offset %d runs past the account data (%d bytes)", offset, len(data))
	}
	return end, nil
}

// VersionedFeeConfigData keeps only the fields the account length selects and zero-pads the rest,
// so a shorter FeeConfig decodes with empty stable tiers and zero exotic fees.
func VersionedFeeConfigData(data []byte) ([]byte, error) {
	if len(data) < FeeConfigSizePreStable {
		return nil, fmt.Errorf("FeeConfig account is %d bytes; expected at least %d", len(data), FeeConfigSizePreStable)
	}
	end, err := feeTierVecEnd(data, feeConfigFeeTiersOffset)
	if err != nil {
		return nil, err
	}
	if len(data) >= FeeConfigSizePostStable {
		end, err = feeTierVecEnd(data, end)
		if err != nil {
			return nil, err
		}
	}
	if len(data) >= FeeConfigSizePostExotic {
		end += feesSize
		if end > len(data) {
			return nil, fmt.Errorf("FeeConfig exotic flat fees run past the account data (%d bytes)", len(data))
		}
	}
	return padTrailing(data[:end], FeeConfigSizePostExotic), nil
}

// DecodeGlobalConfig zero-pads a pre-upgrade GlobalConfig so the trailing fields read as zero.
func DecodeGlobalConfig(data []byte) (*GlobalConfig, error) {
	return ammgen.ParseAccount_GlobalConfig(padTrailing(data, GlobalConfigSize))
}

// DecodeFeeConfig applies the program's length gate before decoding the current FeeConfig layout.
func DecodeFeeConfig(data []byte) (*FeeConfig, error) {
	versioned, err := VersionedFeeConfigData(data)
	if err != nil {
		return nil, err
	}
	return ammgen.ParseAccount_FeeConfig(versioned)
}

// DecodePool zero-pads a pre-upgrade pool. An account already at PoolSize or longer decodes as stored.
func DecodePool(data []byte) (*Pool, error) {
	return ammgen.ParseAccount_Pool(padTrailing(data, PoolSize))
}

// DecodePoolNullable returns nil when data is not a pool account.
func DecodePoolNullable(data []byte) *Pool {
	pool, err := DecodePool(data)
	if err != nil {
		return nil
	}
	return pool
}

// DecodeGlobalVolumeAccumulator decodes a global volume accumulator account.
func DecodeGlobalVolumeAccumulator(data []byte) (*GlobalVolumeAccumulator, error) {
	return ammgen.ParseAccount_GlobalVolumeAccumulator(data)
}

// DecodeUserVolumeAccumulator decodes a user volume accumulator account.
func DecodeUserVolumeAccumulator(data []byte) (*UserVolumeAccumulator, error) {
	return ammgen.ParseAccount_UserVolumeAccumulator(data)
}

// DecodeUserVolumeAccumulatorNullable returns nil when data is not a user volume accumulator.
func DecodeUserVolumeAccumulatorNullable(data []byte) *UserVolumeAccumulator {
	acc, err := DecodeUserVolumeAccumulator(data)
	if err != nil {
		return nil
	}
	return acc
}
