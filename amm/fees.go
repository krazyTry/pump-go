package amm

import (
	"errors"
	"math/rand"

	solana "github.com/gagliardetto/solana-go"
	"github.com/shopspring/decimal"
)

func PoolMarketCap(baseMintSupply, baseReserve, quoteReserve decimal.Decimal) (decimal.Decimal, error) {
	if baseReserve.Sign() == 0 {
		return decimal.Decimal{}, errors.New("cannot divide by zero")
	}
	return quoteReserve.Mul(baseMintSupply).Div(baseReserve), nil
}

func IsPumpPool(baseMint, poolAuthority solana.PublicKey) bool {
	return DerivePumpPoolAuthority(baseMint).Equals(poolAuthority)
}

func CalculateFeeTier(feeTiers []FeeTier, marketCap decimal.Decimal) *Fees {
	first := feeTiers[0]

	if marketCap.Cmp(decimal.RequireFromString(first.MarketCapLamportsThreshold.String())) < 0 {
		return &first.Fees
	}
	for i := len(feeTiers) - 1; i >= 0; i-- {
		if marketCap.Cmp(decimal.RequireFromString(feeTiers[i].MarketCapLamportsThreshold.String())) >= 0 {
			return &feeTiers[i].Fees
		}
	}
	return &first.Fees
}

func ComputeFeesBps(globalConfig *GlobalConfig, feeConfig *FeeConfig, creator solana.PublicKey, baseMintSupply decimal.Decimal, baseMint solana.PublicKey, baseReserve, quoteReserve, tradeSize decimal.Decimal) (*Fees, error) {
	_ = tradeSize
	if feeConfig != nil {
		mc, err := PoolMarketCap(baseMintSupply, baseReserve, quoteReserve)
		if err != nil {
			return &Fees{}, err
		}
		if IsPumpPool(baseMint, creator) {
			return CalculateFeeTier(feeConfig.FeeTiers, mc), nil
		}
		return &feeConfig.FlatFees, nil
	}
	return &Fees{LpFeeBps: globalConfig.LpFeeBasisPoints, ProtocolFeeBps: globalConfig.ProtocolFeeBasisPoints, CreatorFeeBps: globalConfig.CoinCreatorFeeBasisPoints}, nil
}

func GetFeeRecipient(globalConfig *GlobalConfig, isMayhemMode bool) solana.PublicKey {
	if isMayhemMode {
		arr := append([]solana.PublicKey{globalConfig.ReservedFeeRecipient}, globalConfig.ReservedFeeRecipients[:]...)
		return arr[rand.Intn(len(arr))]
	}
	return globalConfig.ProtocolFeeRecipients[rand.Intn(len(globalConfig.ProtocolFeeRecipients))]
}

func GetBuybackFeeRecipient(globalConfig *GlobalConfig) solana.PublicKey {
	return globalConfig.BuybackFeeRecipients[rand.Intn(len(globalConfig.BuybackFeeRecipients))]
}
