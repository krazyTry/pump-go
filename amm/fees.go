package amm

import (
	"errors"
	"math/big"
	"math/rand"

	solana "github.com/solana-foundation/solana-go/v2"
)

// FeeParams is everything computeFeesBps reads off a pool and its accounts.
type FeeParams struct {
	GlobalConfig   *GlobalConfig
	FeeConfig      *FeeConfig
	Creator        solana.PublicKey
	BaseMintSupply uint64
	BaseMint       solana.PublicKey
	BaseReserve    uint64
	// QuoteReserve is the reserve the trade prices against, virtual reserves included.
	QuoteReserve  *big.Int
	QuoteMint     solana.PublicKey
	IsMayhemMode  bool
	CreatorFeeBps uint64
}

func IsSolLikeQuoteMint(quoteMint solana.PublicKey) bool {
	for _, mint := range SolLikeQuoteMints {
		if mint.Equals(quoteMint) {
			return true
		}
	}
	return false
}

func IsStableQuoteMint(quoteMint solana.PublicKey) bool {
	for _, mint := range StableQuoteMints {
		if mint.Equals(quoteMint) {
			return true
		}
	}
	return false
}

func IsZeroFees(fees Fees) bool {
	return fees.LpFeeBps == 0 && fees.ProtocolFeeBps == 0 && fees.CreatorFeeBps == 0
}

func IsPumpPool(baseMint, poolCreator solana.PublicKey) bool {
	return DerivePumpPoolAuthority(baseMint).Equals(poolCreator)
}

// SupportsTradeV2 reports whether buy_v2 / sell_v2 accept the pool.
// A cashback coin stays on v1, because v2 never pays the buyer's cashback.
func SupportsTradeV2(pool *Pool) bool {
	return pool != nil && !pool.IsCashbackCoin
}

// PoolMarketCap is floor(quoteReserve * circulatingSupply / baseReserve).
// A mayhem pool uses the fixed total supply as the circulating supply.
func PoolMarketCap(baseMintSupply, baseReserve uint64, quoteReserve *big.Int, isMayhemMode bool) (*big.Int, error) {
	if baseReserve == 0 {
		return nil, errors.New("Division by zero: pool base token reserves cannot be zero")
	}
	supply := baseMintSupply
	if isMayhemMode {
		supply = PumpAmmTotalTokenSupply
	}
	num := new(big.Int).Mul(nonNil(quoteReserve), u64(supply))
	return num.Quo(num, u64(baseReserve)), nil
}

// CalculateFeeTier picks the highest tier whose threshold the market cap has reached.
func CalculateFeeTier(feeTiers []FeeTier, marketCap *big.Int) (Fees, error) {
	if len(feeTiers) == 0 {
		return Fees{}, errors.New("Fee tiers cannot be empty.")
	}
	first := feeTiers[0]
	cap := nonNil(marketCap)
	if cap.Cmp(first.MarketCapLamportsThreshold.BigInt()) < 0 {
		return first.Fees, nil
	}
	for i := len(feeTiers) - 1; i >= 0; i-- {
		if cap.Cmp(feeTiers[i].MarketCapLamportsThreshold.BigInt()) >= 0 {
			return feeTiers[i].Fees, nil
		}
	}
	return first.Fees, nil
}

// FeesForQuoteMint selects the schedule a trade pays.
// A non-pump pool pays flatFees. A canonical pump pool pays feeTiers for a
// SOL-like quote, stableFeeTiers for USDC, and exoticFlatFees otherwise,
// falling back to flatFees while the exotic schedule is still all-zero.
func FeesForQuoteMint(feeConfig *FeeConfig, isPumpPool bool, marketCap *big.Int, quoteMint solana.PublicKey) (Fees, error) {
	if feeConfig == nil {
		return Fees{}, errors.New("fee config is nil")
	}
	if !isPumpPool {
		return feeConfig.FlatFees, nil
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
	if IsZeroFees(feeConfig.ExoticFlatFees) {
		return feeConfig.FlatFees, nil
	}
	return feeConfig.ExoticFlatFees, nil
}

// ComputeFeesBps is the LP, protocol and creator rates a trade pays.
// While creator fees are configurable and the pool's rate is nonzero, that
// rate replaces the schedule's creator rate. LP and protocol rates stay.
func ComputeFeesBps(p FeeParams) (Fees, error) {
	if p.FeeConfig != nil {
		marketCap, err := PoolMarketCap(p.BaseMintSupply, p.BaseReserve, p.QuoteReserve, p.IsMayhemMode)
		if err != nil {
			return Fees{}, err
		}
		fees, err := FeesForQuoteMint(p.FeeConfig, IsPumpPool(p.BaseMint, p.Creator), marketCap, p.QuoteMint)
		if err != nil {
			return Fees{}, err
		}
		if p.GlobalConfig != nil && p.GlobalConfig.CreatorFeeConfigurable && p.CreatorFeeBps > 0 {
			fees.CreatorFeeBps = p.CreatorFeeBps
		}
		return fees, nil
	}
	if p.GlobalConfig == nil {
		return Fees{}, errors.New("global config is nil")
	}
	return Fees{
		LpFeeBps:       p.GlobalConfig.LpFeeBasisPoints,
		ProtocolFeeBps: p.GlobalConfig.ProtocolFeeBasisPoints,
		CreatorFeeBps:  p.GlobalConfig.CoinCreatorFeeBasisPoints,
	}, nil
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
