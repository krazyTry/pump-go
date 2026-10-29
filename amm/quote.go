package amm

import (
	"errors"
	"math/big"

	solana "github.com/solana-foundation/solana-go/v2"
)

// BuyBaseInputParams prices a buy that takes an exact amount of base out of the pool.
type BuyBaseInputParams struct {
	Base                 uint64
	Slippage             float64
	BaseReserve          uint64
	QuoteReserve         uint64
	VirtualQuoteReserves *big.Int
	GlobalConfig         *GlobalConfig
	BaseMintSupply       uint64
	BaseMint             solana.PublicKey
	CoinCreator          solana.PublicKey
	Creator              solana.PublicKey
	FeeConfig            *FeeConfig
	QuoteMint            solana.PublicKey
	IsMayhemMode         bool
	CreatorFeeBps        uint64
	QuoteTransferFee     *QuoteTransferFee
	// TradeV2 prices the debit the way buy_v2 charges it. The buyback slice
	// of the protocol fee is not grossed up for a quote transfer fee.
	TradeV2 bool
}

// BuyQuoteInputParams prices a buy that spends a quote budget.
type BuyQuoteInputParams struct {
	Quote                uint64
	Slippage             float64
	BaseReserve          uint64
	QuoteReserve         uint64
	VirtualQuoteReserves *big.Int
	GlobalConfig         *GlobalConfig
	BaseMintSupply       uint64
	BaseMint             solana.PublicKey
	CoinCreator          solana.PublicKey
	Creator              solana.PublicKey
	FeeConfig            *FeeConfig
	QuoteMint            solana.PublicKey
	IsMayhemMode         bool
	CreatorFeeBps        uint64
	QuoteTransferFee     *QuoteTransferFee
	TradeV2              bool
}

// SellBaseInputParams prices a sell of an exact amount of base.
type SellBaseInputParams struct {
	Base                 uint64
	Slippage             float64
	BaseReserve          uint64
	QuoteReserve         uint64
	VirtualQuoteReserves *big.Int
	// FeeBucketsTotal is Pool.protocolFees + Pool.creatorFees. Sells are paid
	// only from quoteReserve - FeeBucketsTotal.
	FeeBucketsTotal  uint64
	GlobalConfig     *GlobalConfig
	BaseMintSupply   uint64
	BaseMint         solana.PublicKey
	CoinCreator      solana.PublicKey
	Creator          solana.PublicKey
	FeeConfig        *FeeConfig
	QuoteMint        solana.PublicKey
	IsMayhemMode     bool
	CreatorFeeBps    uint64
	QuoteTransferFee *QuoteTransferFee
}

// SellQuoteInputParams prices the base required to receive a quote amount.
type SellQuoteInputParams struct {
	Quote                uint64
	Slippage             float64
	BaseReserve          uint64
	QuoteReserve         uint64
	VirtualQuoteReserves *big.Int
	FeeBucketsTotal      uint64
	GlobalConfig         *GlobalConfig
	BaseMintSupply       uint64
	BaseMint             solana.PublicKey
	CoinCreator          solana.PublicKey
	Creator              solana.PublicKey
	FeeConfig            *FeeConfig
	QuoteMint            solana.PublicKey
	IsMayhemMode         bool
	CreatorFeeBps        uint64
	QuoteTransferFee     *QuoteTransferFee
}

// ExactQuoteInFees is the fee split of a v2 exact-quote-in buy.
// Fees are ceil-computed on the net before any shave; the shave comes off the net.
type ExactQuoteInFeesResult struct {
	NetQuoteForSwap uint64
	LpFee           uint64
	ProtocolFee     uint64
	CoinCreatorFee  uint64
}

func effectiveQuoteReserve(quoteReserve uint64, virtual *big.Int) *big.Int {
	return new(big.Int).Add(u64(quoteReserve), nonNil(virtual))
}

func feeRates(p FeeParams) (Fees, error) {
	p.QuoteReserve = nonNil(p.QuoteReserve)
	return ComputeFeesBps(p)
}

func creatorBps(coinCreator solana.PublicKey, bps uint64) uint64 {
	if coinCreator.Equals(solana.PublicKey{}) {
		return 0
	}
	return bps
}

// BuyDebit is everything that leaves the buyer's account.
// The pool books quoteAmountIn + lpFee + protocolFee + coinCreatorFee.
// A transfer fee is added on the vault's share so the vault still receives it.
// v1 sends the vault quoteAmountIn + lpFee. v2 sends everything except the
// buyback slice of the protocol fee, and only on a pool that is not in mayhem mode.
func BuyDebit(quoteAmountIn, lpFee, protocolFee, coinCreatorFee uint64, quoteTransferFee *QuoteTransferFee, tradeV2 bool, globalConfig *GlobalConfig, isMayhemMode bool) (uint64, error) {
	booked := new(big.Int).Add(u64(quoteAmountIn), u64(lpFee))
	booked.Add(booked, u64(protocolFee))
	booked.Add(booked, u64(coinCreatorFee))

	buyback := big.NewInt(0)
	if tradeV2 && !isMayhemMode && globalConfig != nil {
		buyback = floorMulDiv(u64(protocolFee), globalConfig.BuybackBasisPoints)
	}
	vault := new(big.Int).Add(u64(quoteAmountIn), u64(lpFee))
	if tradeV2 {
		vault = new(big.Int).Sub(booked, buyback)
	}
	gross, err := GrossForTransferFee(quoteTransferFee, vault)
	if err != nil {
		return 0, err
	}
	debit := new(big.Int).Add(booked, new(big.Int).Sub(gross, vault))
	return fitUint64(debit)
}

// ExactQuoteInFees splits a spendable quote budget into the net that prices
// the swap and the fees, then shaves the net until net + fees fits the budget.
func ExactQuoteInFees(spendable, lpFeeBps, protocolFeeBps, creatorFeeBps uint64) (ExactQuoteInFeesResult, error) {
	totalFeeBps := lpFeeBps + protocolFeeBps + creatorFeeBps
	den := new(big.Int).Add(u64(totalFeeBps), big.NewInt(10_000))
	net := new(big.Int).Mul(u64(spendable), big.NewInt(10_000))
	net.Quo(net, den)
	if net.Sign() <= 0 {
		return ExactQuoteInFeesResult{}, errors.New("The quote budget does not cover the fees.")
	}
	lpFee, err := feeAmount(net, lpFeeBps)
	if err != nil {
		return ExactQuoteInFeesResult{}, err
	}
	protocolFee, err := feeAmount(net, protocolFeeBps)
	if err != nil {
		return ExactQuoteInFeesResult{}, err
	}
	coinCreatorFee, err := feeAmount(net, creatorFeeBps)
	if err != nil {
		return ExactQuoteInFeesResult{}, err
	}
	total := new(big.Int).Add(net, lpFee)
	total.Add(total, protocolFee)
	total.Add(total, coinCreatorFee)
	excess := new(big.Int).Sub(total, u64(spendable))
	if excess.Sign() > 0 {
		net = new(big.Int).Sub(net, excess)
		if net.Sign() <= 0 {
			return ExactQuoteInFeesResult{}, errors.New("The quote budget does not cover the fees.")
		}
	}
	netU, err := fitUint64(net)
	if err != nil {
		return ExactQuoteInFeesResult{}, err
	}
	lpU, err := fitUint64(lpFee)
	if err != nil {
		return ExactQuoteInFeesResult{}, err
	}
	protocolU, err := fitUint64(protocolFee)
	if err != nil {
		return ExactQuoteInFeesResult{}, err
	}
	creatorU, err := fitUint64(coinCreatorFee)
	if err != nil {
		return ExactQuoteInFeesResult{}, err
	}
	return ExactQuoteInFeesResult{
		NetQuoteForSwap: netU,
		LpFee:           lpU,
		ProtocolFee:     protocolU,
		CoinCreatorFee:  creatorU,
	}, nil
}

func requireReserves(baseReserve, quoteReserve uint64) error {
	if baseReserve == 0 || quoteReserve == 0 {
		return errors.New("Invalid input: 'baseReserve' or 'quoteReserve' cannot be zero.")
	}
	return nil
}

// BuyBaseInput quotes the quote a buy of base spends, fees and transfer fee included.
func BuyBaseInput(p BuyBaseInputParams) (*BuyBaseInputResult, error) {
	if err := requireReserves(p.BaseReserve, p.QuoteReserve); err != nil {
		return nil, err
	}
	if p.Base > p.BaseReserve {
		return nil, errors.New("Cannot buy more base tokens than the pool reserves.")
	}
	effective := effectiveQuoteReserve(p.QuoteReserve, p.VirtualQuoteReserves)
	numerator := new(big.Int).Mul(effective, u64(p.Base))
	denominator := new(big.Int).Sub(u64(p.BaseReserve), u64(p.Base))
	if denominator.Sign() == 0 {
		return nil, errors.New("Pool would be depleted; denominator is zero.")
	}
	quoteAmountIn, err := ceilDiv(numerator, denominator)
	if err != nil {
		return nil, err
	}
	rates, err := feeRates(FeeParams{
		GlobalConfig:   p.GlobalConfig,
		FeeConfig:      p.FeeConfig,
		Creator:        p.Creator,
		BaseMintSupply: p.BaseMintSupply,
		BaseMint:       p.BaseMint,
		BaseReserve:    p.BaseReserve,
		QuoteReserve:   effective,
		QuoteMint:      p.QuoteMint,
		IsMayhemMode:   p.IsMayhemMode,
		CreatorFeeBps:  p.CreatorFeeBps,
	})
	if err != nil {
		return nil, err
	}
	lpFee, err := feeAmount(quoteAmountIn, rates.LpFeeBps)
	if err != nil {
		return nil, err
	}
	protocolFee, err := feeAmount(quoteAmountIn, rates.ProtocolFeeBps)
	if err != nil {
		return nil, err
	}
	coinCreatorFee, err := feeAmount(quoteAmountIn, creatorBps(p.CoinCreator, rates.CreatorFeeBps))
	if err != nil {
		return nil, err
	}
	quoteU, err := fitUint64(quoteAmountIn)
	if err != nil {
		return nil, err
	}
	lpU, err := fitUint64(lpFee)
	if err != nil {
		return nil, err
	}
	protocolU, err := fitUint64(protocolFee)
	if err != nil {
		return nil, err
	}
	creatorU, err := fitUint64(coinCreatorFee)
	if err != nil {
		return nil, err
	}
	total, err := BuyDebit(quoteU, lpU, protocolU, creatorU, p.QuoteTransferFee, p.TradeV2, p.GlobalConfig, p.IsMayhemMode)
	if err != nil {
		return nil, err
	}
	maxQuote, err := fitUint64(applySlippage(u64(total), p.Slippage, true))
	if err != nil {
		return nil, err
	}
	return &BuyBaseInputResult{
		InternalQuoteAmount: quoteU,
		UIQuote:             total,
		MaxQuote:            maxQuote,
	}, nil
}

// BuyQuoteInput quotes the base a quote budget buys.
// quote is what leaves the wallet, transfer fee included.
func BuyQuoteInput(p BuyQuoteInputParams) (*BuyQuoteInputResult, error) {
	if err := requireReserves(p.BaseReserve, p.QuoteReserve); err != nil {
		return nil, err
	}
	effective := effectiveQuoteReserve(p.QuoteReserve, p.VirtualQuoteReserves)
	rates, err := feeRates(FeeParams{
		GlobalConfig:   p.GlobalConfig,
		FeeConfig:      p.FeeConfig,
		Creator:        p.Creator,
		BaseMintSupply: p.BaseMintSupply,
		BaseMint:       p.BaseMint,
		BaseReserve:    p.BaseReserve,
		QuoteReserve:   effective,
		QuoteMint:      p.QuoteMint,
		IsMayhemMode:   p.IsMayhemMode,
		CreatorFeeBps:  p.CreatorFeeBps,
	})
	if err != nil {
		return nil, err
	}
	creatorRate := creatorBps(p.CoinCreator, rates.CreatorFeeBps)
	totalFeeBps := rates.LpFeeBps + rates.ProtocolFeeBps + creatorRate
	denominator := new(big.Int).Add(big.NewInt(10_000), u64(totalFeeBps))

	spendable, err := NetOfTransferFee(p.QuoteTransferFee, u64(p.Quote))
	if err != nil {
		return nil, err
	}
	if spendable.Sign() <= 0 {
		return nil, errors.New("The quote budget does not cover the transfer fee.")
	}
	effectiveQuote := new(big.Int).Mul(spendable, big.NewInt(10_000))
	effectiveQuote.Quo(effectiveQuote, denominator)

	lpFee, err := feeAmount(effectiveQuote, rates.LpFeeBps)
	if err != nil {
		return nil, err
	}
	protocolFee, err := feeAmount(effectiveQuote, rates.ProtocolFeeBps)
	if err != nil {
		return nil, err
	}
	coinCreatorFee, err := feeAmount(effectiveQuote, creatorRate)
	if err != nil {
		return nil, err
	}
	totalWithFees := new(big.Int).Add(effectiveQuote, lpFee)
	totalWithFees.Add(totalWithFees, protocolFee)
	totalWithFees.Add(totalWithFees, coinCreatorFee)
	if totalWithFees.Cmp(spendable) > 0 {
		effectiveQuote = new(big.Int).Sub(effectiveQuote, new(big.Int).Sub(totalWithFees, spendable))
	}
	if effectiveQuote.Sign() <= 0 {
		return nil, errors.New("The quote budget does not cover the fees.")
	}

	quoteU, err := fitUint64(effectiveQuote)
	if err != nil {
		return nil, err
	}
	lpU, err := fitUint64(lpFee)
	if err != nil {
		return nil, err
	}
	protocolU, err := fitUint64(protocolFee)
	if err != nil {
		return nil, err
	}
	creatorU, err := fitUint64(coinCreatorFee)
	if err != nil {
		return nil, err
	}
	debit, err := BuyDebit(quoteU, lpU, protocolU, creatorU, p.QuoteTransferFee, p.TradeV2, p.GlobalConfig, p.IsMayhemMode)
	if err != nil {
		return nil, err
	}
	if debit > p.Quote {
		effectiveQuote = new(big.Int).Sub(effectiveQuote, u64(debit-p.Quote))
	}
	if effectiveQuote.Sign() <= 0 {
		return nil, errors.New("The quote budget does not cover the fees.")
	}

	inputAmount := new(big.Int).Sub(effectiveQuote, big.NewInt(1))
	numerator := new(big.Int).Mul(u64(p.BaseReserve), inputAmount)
	denominatorEffective := new(big.Int).Add(effective, inputAmount)
	if denominatorEffective.Sign() == 0 {
		return nil, errors.New("Pool would be depleted; denominator is zero.")
	}
	baseOut := new(big.Int).Quo(numerator, denominatorEffective)
	baseU, err := fitUint64(baseOut)
	if err != nil {
		return nil, err
	}
	internalU, err := fitUint64(effectiveQuote)
	if err != nil {
		return nil, err
	}
	maxQuote, err := fitUint64(applySlippage(u64(p.Quote), p.Slippage, true))
	if err != nil {
		return nil, err
	}
	return &BuyQuoteInputResult{
		Base:                     baseU,
		InternalQuoteWithoutFees: internalU,
		MaxQuote:                 maxQuote,
	}, nil
}

// sellSplit splits a sell's gross quote into what leaves the reserves once the
// LP fee stays, and what the seller receives once the protocol and creator fees stay too.
// user may be negative when the fees exceed the output; the caller decides what that means.
func sellSplit(quoteAmountOut, lpFee, protocolFee, coinCreatorFee *big.Int, realQuoteReserves uint64) (without, user *big.Int, err error) {
	without = new(big.Int).Sub(nonNil(quoteAmountOut), nonNil(lpFee))
	if u64(realQuoteReserves).Cmp(without) < 0 {
		return nil, nil, errors.New("Insufficient real quote reserves to cover the sell output.")
	}
	user = new(big.Int).Sub(without, nonNil(coinCreatorFee))
	user.Sub(user, nonNil(protocolFee))
	return without, user, nil
}

// SellAmounts splits a sell's gross quote into what leaves the reserves
// (LP fee stays) and what the seller receives (protocol and creator fees stay too).
func SellAmounts(quoteAmountOut, lpFee, protocolFee, coinCreatorFee, realQuoteReserves uint64) (quoteOutWithoutLpFee, userQuoteAmountOut uint64, err error) {
	without, user, err := sellSplit(u64(quoteAmountOut), u64(lpFee), u64(protocolFee), u64(coinCreatorFee), realQuoteReserves)
	if err != nil {
		return 0, 0, err
	}
	if user.Sign() < 0 {
		return 0, 0, errors.New("Fees exceed total output; final quote is negative.")
	}
	withoutU, err := fitUint64(without)
	if err != nil {
		return 0, 0, err
	}
	userU, err := fitUint64(user)
	if err != nil {
		return 0, 0, err
	}
	return withoutU, userU, nil
}

func realQuoteAfterBuckets(quoteReserve, feeBucketsTotal uint64) (uint64, error) {
	if feeBucketsTotal > quoteReserve {
		return 0, errors.New("Insufficient real quote reserves to cover the sell output.")
	}
	return quoteReserve - feeBucketsTotal, nil
}

// SellBaseInput quotes the quote a seller receives for base, after fees.
func SellBaseInput(p SellBaseInputParams) (SellBaseInputResult, error) {
	if err := requireReserves(p.BaseReserve, p.QuoteReserve); err != nil {
		return SellBaseInputResult{}, err
	}
	effective := effectiveQuoteReserve(p.QuoteReserve, p.VirtualQuoteReserves)
	quoteAmountOut := new(big.Int).Mul(effective, u64(p.Base))
	quoteAmountOut.Quo(quoteAmountOut, new(big.Int).Add(u64(p.BaseReserve), u64(p.Base)))

	rates, err := feeRates(FeeParams{
		GlobalConfig:   p.GlobalConfig,
		FeeConfig:      p.FeeConfig,
		Creator:        p.Creator,
		BaseMintSupply: p.BaseMintSupply,
		BaseMint:       p.BaseMint,
		BaseReserve:    p.BaseReserve,
		QuoteReserve:   effective,
		QuoteMint:      p.QuoteMint,
		IsMayhemMode:   p.IsMayhemMode,
		CreatorFeeBps:  p.CreatorFeeBps,
	})
	if err != nil {
		return SellBaseInputResult{}, err
	}
	lpFee, err := feeAmount(quoteAmountOut, rates.LpFeeBps)
	if err != nil {
		return SellBaseInputResult{}, err
	}
	protocolFee, err := feeAmount(quoteAmountOut, rates.ProtocolFeeBps)
	if err != nil {
		return SellBaseInputResult{}, err
	}
	coinCreatorFee, err := feeAmount(quoteAmountOut, creatorBps(p.CoinCreator, rates.CreatorFeeBps))
	if err != nil {
		return SellBaseInputResult{}, err
	}
	rawU, err := fitUint64(quoteAmountOut)
	if err != nil {
		return SellBaseInputResult{}, err
	}
	real, err := realQuoteAfterBuckets(p.QuoteReserve, p.FeeBucketsTotal)
	if err != nil {
		return SellBaseInputResult{}, err
	}
	_, userOut, err := sellSplit(quoteAmountOut, lpFee, protocolFee, coinCreatorFee, real)
	if err != nil {
		return SellBaseInputResult{}, err
	}
	if userOut.Sign() < 0 {
		return SellBaseInputResult{}, errors.New("Fees exceed total output; final quote is negative.")
	}
	finalQuote, err := NetOfTransferFee(p.QuoteTransferFee, userOut)
	if err != nil {
		return SellBaseInputResult{}, err
	}
	if finalQuote.Sign() < 0 {
		return SellBaseInputResult{}, errors.New("Fees exceed total output; final quote is negative.")
	}
	finalU, err := fitUint64(finalQuote)
	if err != nil {
		return SellBaseInputResult{}, err
	}
	minQuote, err := fitUint64(applySlippage(finalQuote, p.Slippage, false))
	if err != nil {
		return SellBaseInputResult{}, err
	}
	return SellBaseInputResult{
		UIQuote:                finalU,
		MinQuote:               minQuote,
		InternalQuoteAmountOut: rawU,
	}, nil
}

func calculateQuoteAmountOut(userQuote *big.Int, lpBps, protocolBps, creatorBps uint64) (*big.Int, error) {
	total := lpBps + protocolBps + creatorBps
	if total > 10_000 {
		return nil, errors.New("Fees exceed total output; final quote is negative.")
	}
	denominator := new(big.Int).Sub(big.NewInt(10_000), u64(total))
	return ceilDiv(new(big.Int).Mul(userQuote, big.NewInt(10_000)), denominator)
}

// SellQuoteInput quotes the base required for the wallet to receive quote.
func SellQuoteInput(p SellQuoteInputParams) (SellQuoteInputResult, error) {
	if err := requireReserves(p.BaseReserve, p.QuoteReserve); err != nil {
		return SellQuoteInputResult{}, err
	}
	effective := effectiveQuoteReserve(p.QuoteReserve, p.VirtualQuoteReserves)
	rates, err := feeRates(FeeParams{
		GlobalConfig:   p.GlobalConfig,
		FeeConfig:      p.FeeConfig,
		Creator:        p.Creator,
		BaseMintSupply: p.BaseMintSupply,
		BaseMint:       p.BaseMint,
		BaseReserve:    p.BaseReserve,
		QuoteReserve:   effective,
		QuoteMint:      p.QuoteMint,
		IsMayhemMode:   p.IsMayhemMode,
		CreatorFeeBps:  p.CreatorFeeBps,
	})
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	grossQuote, err := GrossForTransferFee(p.QuoteTransferFee, u64(p.Quote))
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	rawQuote, err := calculateQuoteAmountOut(grossQuote, rates.LpFeeBps, rates.ProtocolFeeBps, creatorBps(p.CoinCreator, rates.CreatorFeeBps))
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	if rawQuote.Cmp(effective) >= 0 {
		return SellQuoteInputResult{}, errors.New("Invalid input: Desired quote amount exceeds available reserve.")
	}
	baseIn, err := ceilDiv(
		new(big.Int).Mul(u64(p.BaseReserve), rawQuote),
		new(big.Int).Sub(effective, rawQuote),
	)
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	quoteAmountOut := new(big.Int).Mul(effective, baseIn)
	quoteAmountOut.Quo(quoteAmountOut, new(big.Int).Add(u64(p.BaseReserve), baseIn))
	lpFee, err := feeAmount(quoteAmountOut, rates.LpFeeBps)
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	real, err := realQuoteAfterBuckets(p.QuoteReserve, p.FeeBucketsTotal)
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	outWithoutLp := new(big.Int).Sub(quoteAmountOut, lpFee)
	if u64(real).Cmp(outWithoutLp) < 0 {
		return SellQuoteInputResult{}, errors.New("Insufficient real quote reserves to cover the sell output.")
	}
	rawU, err := fitUint64(rawQuote)
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	baseU, err := fitUint64(baseIn)
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	minQuote, err := fitUint64(applySlippage(u64(p.Quote), p.Slippage, false))
	if err != nil {
		return SellQuoteInputResult{}, err
	}
	return SellQuoteInputResult{
		InternalRawQuote: rawU,
		Base:             baseU,
		MinQuote:         minQuote,
	}, nil
}
