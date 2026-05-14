package helpers

import (
	solana "github.com/gagliardetto/solana-go"
	pump "github.com/krazyTry/pump-go/gen/pump"
	"github.com/shopspring/decimal"
)

const (
	OneBillionSupplyU64 = uint64(1_000_000_000_000_000)
)

type CalculatedFeesBps struct {
	ProtocolFeeBps uint64
	CreatorFeeBps  uint64
}

func fee(amount, bps uint64) uint64 {
	amountDecimal := decimal.NewFromUint64(amount)
	bpsDecimal := decimal.NewFromUint64(bps)

	aDecimal := amountDecimal.Mul(bpsDecimal)
	bDecimal := decimal.NewFromInt(10_000)

	// (a + b - 1) / b
	return aDecimal.Add(bDecimal).Sub(decimal.NewFromInt(1)).Div(bDecimal).Ceil().BigInt().Uint64()
}

func CalculateFeeTier(feeTiers []pump.FeeTier, marketCap uint64) pump.Fees {
	first := feeTiers[0]
	if marketCap < first.MarketCapLamportsThreshold.BigInt().Uint64() {
		return first.Fees
	}
	for i := len(feeTiers) - 1; i >= 0; i-- {
		if marketCap >= feeTiers[i].MarketCapLamportsThreshold.BigInt().Uint64() {
			return feeTiers[i].Fees
		}
	}
	return first.Fees
}

func ComputeFeesBps(global *pump.Global, feeConfig *pump.FeeConfig, mintSupply, vQuote, vToken uint64) (CalculatedFeesBps, error) {
	if feeConfig != nil {
		mc, err := BondingCurveMarketCap(mintSupply, vQuote, vToken)
		if err != nil {
			return CalculatedFeesBps{}, err
		}
		t := CalculateFeeTier(feeConfig.FeeTiers, mc)
		return CalculatedFeesBps{ProtocolFeeBps: t.ProtocolFeeBps, CreatorFeeBps: t.CreatorFeeBps}, nil
	}
	return CalculatedFeesBps{ProtocolFeeBps: global.FeeBasisPoints, CreatorFeeBps: global.CreatorFeeBasisPoints}, nil
}

func GetFee(global *pump.Global, feeConfig *pump.FeeConfig, mintSupply uint64, curve *pump.BondingCurve, amount uint64, isNew bool) (uint64, error) {
	m := mintSupply
	if !curve.IsMayhemMode {
		m = OneBillionSupplyU64
	}
	bps, err := ComputeFeesBps(global, feeConfig, m, curve.VirtualQuoteReserves, curve.VirtualTokenReserves)
	if err != nil {
		return 0, err
	}
	out := fee(amount, bps.ProtocolFeeBps)
	if isNew || !curve.Creator.Equals(solana.PublicKey{}) {
		out += fee(amount, bps.CreatorFeeBps)
	}
	return out, nil
}

func GetFeeRecipient(global *pump.Global, mayhem bool) solana.PublicKey {
	if mayhem {
		i := rnd.Intn(1 + len(global.ReservedFeeRecipients))
		if i == 0 {
			return global.ReservedFeeRecipient
		}
		return global.ReservedFeeRecipients[i-1]
	}

	i := rnd.Intn(1 + len(global.FeeRecipients))
	if i == 0 {
		return global.FeeRecipient
	}
	return global.FeeRecipients[i-1]
}
