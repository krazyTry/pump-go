package bonding_curve

import (
	"math/big"

	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

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

// QuoteTransferFee is the Token-2022 transfer fee a quote mint charges.
type QuoteTransferFee = helpers.QuoteTransferFee

// PumpQuotePool is a migrated pump coin's canonical pool, as create_v2 prices it.
// VirtualQuoteReserves may be negative.
type PumpQuotePool struct {
	BaseReserves         *big.Int
	QuoteReserves        *big.Int
	VirtualQuoteReserves *big.Int
}

// NewPumpQuotePool builds a pool from unsigned vault balances and a signed virtual quote.
func NewPumpQuotePool(base, quote uint64, virtual int64) *PumpQuotePool {
	return &PumpQuotePool{
		BaseReserves:         new(big.Int).SetUint64(base),
		QuoteReserves:        new(big.Int).SetUint64(quote),
		VirtualQuoteReserves: big.NewInt(virtual),
	}
}

// PumpQuoteChainCoin is one pump coin on a quote chain.
type PumpQuoteChainCoin struct {
	Mint         solana.PublicKey
	BondingCurve *BondingCurve
	// Pool is set once the coin's curve has migrated.
	Pool *PumpQuotePool
}

// PumpQuoteCurve is a pump coin offered as a create_v2 quote mint.
type PumpQuoteCurve struct {
	PumpQuoteChainCoin
	// QuoteControlKnown distinguishes "not fetched" from a missing account.
	// A missing account is QuoteControlKnown with a nil QuoteControl.
	QuoteControlKnown bool
	QuoteControl      *pump.QuoteControl
	// Chain is the rest of the quote chain, nearest first. Empty for a depth-0 coin.
	Chain []PumpQuoteChainCoin
}

// CurveSeedParams is the input to NewBondingCurve.
type CurveSeedParams struct {
	QuoteMint solana.PublicKey
	// QuoteControlKnown is set when the caller fetched QuoteControl, including
	// a missing account (nil QuoteControl). Leave it false to fall back to
	// Global.InitialVirtualQuoteReserves for an unlisted mint.
	QuoteControlKnown bool
	QuoteControl      *pump.QuoteControl
	CreatorFeeBpsSet  bool
	CreatorFeeBps     uint64
	IsHolderReward    bool
	PumpQuote         *PumpQuoteCurve
}

// TradeQuoteParams prices a v1/v2 buy or sell.
// MintSupplySet false means the supply is unknown (a curve about to be created).
type TradeQuoteParams struct {
	Global            *Global
	FeeConfig         *FeeConfig
	MintSupplySet     bool
	MintSupply        uint64
	BondingCurve      *BondingCurve
	Amount            uint64
	QuoteMint         solana.PublicKey
	QuoteControlKnown bool
	QuoteControl      *pump.QuoteControl
	CreatorFeeBpsSet  bool
	CreatorFeeBps     uint64
	PumpQuote         *PumpQuoteCurve
	QuoteTransferFee  *QuoteTransferFee
}

// V3QuoteParams prices a buy_v3 or buy_exact_quote_in_v3.
type V3QuoteParams struct {
	Global                *Global
	FeeConfig             *FeeConfig
	MintSupply            uint64
	BondingCurve          *BondingCurve
	Amount                uint64
	CurveBaseTokenBalance uint64
	QuoteTransferFee      *QuoteTransferFee
}

// CurveLiquidity is the constant-product liquidity a curve quotes against.
type CurveLiquidity struct {
	VirtualTokenReserves uint64
	VirtualQuoteReserves uint64
	RealTokenReserves    uint64
	RealQuoteReserves    uint64
	// K is virtualTokenReserves * virtualQuoteReserves.
	K *big.Int
	// MarketCap is floor(virtualQuoteReserves * mintSupply / virtualTokenReserves).
	MarketCap *big.Int
	// TradableSupply is the base the curve can still sell.
	TradableSupply uint64
}

// MigrationLiquidity is the pool migrate_v2 would deposit once the curve leg completes.
type MigrationLiquidity struct {
	Base  uint64
	Quote uint64
}

// PumpQuoteAccounts are the extra create_v2 accounts for a pump-coin quote.
type PumpQuoteAccounts struct {
	Pool  *PumpQuotePoolAccounts
	Chain []PumpQuoteChainAccounts
}

// PumpQuotePoolAccounts identifies a migrated coin's canonical pool and vaults.
type PumpQuotePoolAccounts struct {
	Pool       solana.PublicKey
	BaseVault  solana.PublicKey
	QuoteVault solana.PublicKey
}

// PumpQuoteChainAccounts is one coin below the quote mint.
type PumpQuoteChainAccounts struct {
	Mint solana.PublicKey
	Pool *PumpQuotePoolAccounts
}
