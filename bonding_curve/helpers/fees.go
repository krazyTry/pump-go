package helpers

import (
	"errors"
	"math/big"

	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

const (
	OneBillionSupplyU64 = uint64(1_000_000_000_000_000)
	feeBpsDenominator   = uint64(10_000)
)

type CalculatedFeesBps struct {
	ProtocolFeeBps uint64
	CreatorFeeBps  uint64
}

func fee(amount, bps uint64) (uint64, error) {
	if amount == 0 || bps == 0 {
		return 0, nil
	}
	num := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(bps))
	num.Add(num, big.NewInt(int64(feeBpsDenominator-1)))
	num.Quo(num, new(big.Int).SetUint64(feeBpsDenominator))
	if !num.IsUint64() {
		return 0, errors.New("fee overflows uint64")
	}
	return num.Uint64(), nil
}

func isZeroFees(f pump.Fees) bool {
	return f.LpFeeBps == 0 && f.ProtocolFeeBps == 0 && f.CreatorFeeBps == 0
}

// CalculateFeeTier mirrors pump-fees-math::calculate_fee_tier.
// marketCap is in the quote mint's base units.
func CalculateFeeTier(feeTiers []pump.FeeTier, marketCap *big.Int) (pump.Fees, error) {
	if len(feeTiers) == 0 {
		return pump.Fees{}, errors.New("fee tiers cannot be empty")
	}
	first := feeTiers[0]
	if marketCap.Cmp(first.MarketCapLamportsThreshold.BigInt()) < 0 {
		return first.Fees, nil
	}
	for i := len(feeTiers) - 1; i >= 0; i-- {
		if marketCap.Cmp(feeTiers[i].MarketCapLamportsThreshold.BigInt()) >= 0 {
			return feeTiers[i].Fees, nil
		}
	}
	return first.Fees, nil
}

// BondingCurveMarketCap is floor(virtualQuoteReserves * mintSupply / virtualTokenReserves).
func BondingCurveMarketCap(mintSupply, virtualQuoteReserves, virtualTokenReserves uint64) (*big.Int, error) {
	if virtualTokenReserves == 0 {
		return nil, errors.New("division by zero: virtual token reserves cannot be zero")
	}
	num := new(big.Int).Mul(new(big.Int).SetUint64(virtualQuoteReserves), new(big.Int).SetUint64(mintSupply))
	return num.Quo(num, new(big.Int).SetUint64(virtualTokenReserves)), nil
}

// SelectCurveFeeSchedule picks the canonical-curve schedule the way
// FeeConfig::fees_for_quote_mint does with is_pump_pool = true.
func SelectCurveFeeSchedule(feeConfig *pump.FeeConfig, quoteMint solana.PublicKey, marketCap *big.Int) (pump.Fees, error) {
	if feeConfig == nil {
		return pump.Fees{}, errors.New("fee config is nil")
	}
	if IsSolLikeQuoteMint(quoteMint) {
		return CalculateFeeTier(feeConfig.FeeTiers, marketCap)
	}
	if IsStableQuoteMint(quoteMint) {
		tiers := feeConfig.StableFeeTiers
		if len(tiers) == 0 {
			tiers = feeConfig.FeeTiers
		}
		return CalculateFeeTier(tiers, marketCap)
	}
	if isZeroFees(feeConfig.ExoticFlatFees) {
		return feeConfig.FlatFees, nil
	}
	return feeConfig.ExoticFlatFees, nil
}

// ComputeFeesBps is the fee rates a bonding-curve trade pays.
// quoteMint is the curve's stored quote (the zero key means SOL).
// creatorFeeBps replaces the schedule's creator rate while
// global.CreatorFeeConfigurable is on and the value is nonzero.
func ComputeFeesBps(global *pump.Global, feeConfig *pump.FeeConfig, mintSupply, vQuote, vToken uint64, quoteMint solana.PublicKey, creatorFeeBps uint64) (CalculatedFeesBps, error) {
	schedule, err := scheduleFeesBps(global, feeConfig, mintSupply, vQuote, vToken, quoteMint)
	if err != nil {
		return CalculatedFeesBps{}, err
	}
	if !global.CreatorFeeConfigurable || creatorFeeBps == 0 {
		return schedule, nil
	}
	return CalculatedFeesBps{ProtocolFeeBps: schedule.ProtocolFeeBps, CreatorFeeBps: creatorFeeBps}, nil
}

func scheduleFeesBps(global *pump.Global, feeConfig *pump.FeeConfig, mintSupply, vQuote, vToken uint64, quoteMint solana.PublicKey) (CalculatedFeesBps, error) {
	if feeConfig != nil {
		mc, err := BondingCurveMarketCap(mintSupply, vQuote, vToken)
		if err != nil {
			return CalculatedFeesBps{}, err
		}
		fees, err := SelectCurveFeeSchedule(feeConfig, quoteMint, mc)
		if err != nil {
			return CalculatedFeesBps{}, err
		}
		return CalculatedFeesBps{ProtocolFeeBps: fees.ProtocolFeeBps, CreatorFeeBps: fees.CreatorFeeBps}, nil
	}
	return CalculatedFeesBps{ProtocolFeeBps: global.FeeBasisPoints, CreatorFeeBps: global.CreatorFeeBasisPoints}, nil
}

// GetFee is the protocol fee plus the creator fee when the curve is new or has a creator.
// Non-mayhem curves are tiered at a one-billion supply, matching the TypeScript SDK.
func GetFee(global *pump.Global, feeConfig *pump.FeeConfig, mintSupply uint64, curve *pump.BondingCurve, amount uint64, isNew bool) (uint64, error) {
	m := mintSupply
	if !curve.IsMayhemMode {
		m = OneBillionSupplyU64
	}
	bps, err := ComputeFeesBps(global, feeConfig, m, curve.VirtualQuoteReserves, curve.VirtualTokenReserves, curve.QuoteMint, curve.CreatorFeeBps)
	if err != nil {
		return 0, err
	}
	out, err := fee(amount, bps.ProtocolFeeBps)
	if err != nil {
		return 0, err
	}
	if isNew || !curve.Creator.Equals(solana.PublicKey{}) {
		creator, err := fee(amount, bps.CreatorFeeBps)
		if err != nil {
			return 0, err
		}
		sum := new(big.Int).Add(new(big.Int).SetUint64(out), new(big.Int).SetUint64(creator))
		if !sum.IsUint64() {
			return 0, errors.New("fee overflows uint64")
		}
		out = sum.Uint64()
	}
	return out, nil
}

// FeeRecipientCandidates is the non-zero fee recipients for a curve, primary first.
// An all-zero list falls back to the primary key, which may itself be zero.
func FeeRecipientCandidates(global *pump.Global, mayhem bool) []solana.PublicKey {
	var keys []solana.PublicKey
	if mayhem {
		keys = append(keys, global.ReservedFeeRecipient)
		keys = append(keys, global.ReservedFeeRecipients[:]...)
	} else {
		keys = append(keys, global.FeeRecipient)
		keys = append(keys, global.FeeRecipients[:]...)
	}
	nonzero := make([]solana.PublicKey, 0, len(keys))
	for _, key := range keys {
		if !key.Equals(solana.PublicKey{}) {
			nonzero = append(nonzero, key)
		}
	}
	if len(nonzero) == 0 {
		return []solana.PublicKey{keys[0]}
	}
	return nonzero
}

func GetFeeRecipient(global *pump.Global, mayhem bool) solana.PublicKey {
	candidates := FeeRecipientCandidates(global, mayhem)
	return candidates[rnd.Intn(len(candidates))]
}

// GetBuybackFeeRecipientForTokenQuote is the wallet whose quote ATA v3 token-quote
// trades pay. It is always the first listed buyback recipient.
func GetBuybackFeeRecipientForTokenQuote() solana.PublicKey {
	return buybackFeeRecipients[0]
}
