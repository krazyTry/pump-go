package bonding_curve

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	"github.com/krazyTry/pump-go/bonding_curve/math"
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

// IsLegacyQuoteMint reports whether create_v2 treats the mint as SOL.
func IsLegacyQuoteMint(quoteMint solana.PublicKey) bool {
	return helpers.IsLegacyQuoteMint(quoteMint)
}

// IsSolLikeQuoteMint reports whether pump-fees charges the SOL fee tiers.
func IsSolLikeQuoteMint(quoteMint solana.PublicKey) bool {
	return helpers.IsSolLikeQuoteMint(quoteMint)
}

// IsStableQuoteMint reports whether pump-fees charges the stable fee tiers.
func IsStableQuoteMint(quoteMint solana.PublicKey) bool {
	return helpers.IsStableQuoteMint(quoteMint)
}

// IsExoticQuoteMint reports a quote that pays exotic (or flat) fees.
func IsExoticQuoteMint(quoteMint solana.PublicKey) bool {
	return helpers.IsExoticQuoteMint(quoteMint)
}

// NormalizeQuoteMint maps the zero key and WSOL to WSOL.
func NormalizeQuoteMint(quoteMint solana.PublicKey) solana.PublicKey {
	return helpers.NormalizeQuoteMint(quoteMint)
}

// IsBondingCurveMigrated reports a curve whose reserves were zeroed by migration.
func IsBondingCurveMigrated(curve *BondingCurve) bool {
	return curve.RealQuoteReserves == 0 &&
		curve.VirtualQuoteReserves == 0 &&
		curve.RealTokenReserves == 0 &&
		curve.VirtualTokenReserves == 0
}

// AssertCreateV2Flags checks the flags create_v2 rejects before it builds a coin.
func AssertCreateV2Flags(global *Global, mint solana.PublicKey, cashback, holderReward bool, creatorFeeBps uint64) error {
	if cashback {
		return &CashbackDeprecatedError{Mint: mint}
	}
	if holderReward && (global == nil || !global.IsHolderRewardEnabled) {
		return &HolderRewardDisabledError{}
	}
	if creatorFeeBps == 0 {
		return nil
	}
	if global == nil || !global.CreatorFeeConfigurable {
		return &CreatorFeeNotConfigurableError{}
	}
	if creatorFeeBps > global.MaxConfigurableCreatorFeeBps {
		return &CreatorFeeBpsOutOfRangeError{CreatorFeeBps: creatorFeeBps, Max: global.MaxConfigurableCreatorFeeBps}
	}
	return nil
}

// CurveLiquidityOf is the constant-product liquidity the curve prices against.
func CurveLiquidityOf(curve *BondingCurve, mintSupply uint64) (CurveLiquidity, error) {
	if curve == nil {
		return CurveLiquidity{}, errors.New("bonding curve is nil")
	}
	marketCap, err := helpers.BondingCurveMarketCap(mintSupply, curve.VirtualQuoteReserves, curve.VirtualTokenReserves)
	if err != nil {
		return CurveLiquidity{}, err
	}
	return CurveLiquidity{
		VirtualTokenReserves: curve.VirtualTokenReserves,
		VirtualQuoteReserves: curve.VirtualQuoteReserves,
		RealTokenReserves:    curve.RealTokenReserves,
		RealQuoteReserves:    curve.RealQuoteReserves,
		K:                    math.ConstantProduct(curve.VirtualTokenReserves, curve.VirtualQuoteReserves),
		MarketCap:            marketCap,
		TradableSupply:       curve.RealTokenReserves,
	}, nil
}

// InitialVirtualQuoteReservesFor is the virtualQuoteReserves a new curve quoted
// in quoteMint starts from, the way create_v2 seeds it.
func InitialVirtualQuoteReservesFor(global *Global, quoteMint solana.PublicKey, quoteControlKnown bool, quoteControl *pump.QuoteControl, pumpQuote *PumpQuoteCurve) (uint64, error) {
	if global == nil {
		return 0, errors.New("global is nil")
	}
	if IsLegacyQuoteMint(quoteMint) {
		return global.InitialVirtualSolReserves, nil
	}
	if !quoteControlKnown && pumpQuote != nil && pumpQuote.QuoteControlKnown {
		quoteControlKnown = true
		quoteControl = pumpQuote.QuoteControl
	}
	if quoteMint.Equals(helpers.NativeMint2022) {
		return 0, &UnsupportedQuoteMintError{QuoteMint: quoteMint}
	}
	if quoteWhitelisted(global, quoteMint) {
		return global.InitialVirtualQuoteReserves, nil
	}
	if reserves, ok := quoteControlReserves(quoteControl, quoteMint); ok {
		return reserves, nil
	}
	if pumpQuote != nil {
		if !pumpQuote.Mint.Equals(quoteMint) {
			return 0, fmt.Errorf("pumpQuote is for %s, not the quote mint %s", pumpQuote.Mint, quoteMint)
		}
		seed, err := PumpQuoteReserves(global, pumpQuote, quoteControlKnown, quoteControl)
		if err != nil {
			return 0, err
		}
		return seed.VirtualQuoteReserves, nil
	}
	if quoteControlKnown {
		return 0, &UnsupportedQuoteMintError{QuoteMint: quoteMint}
	}
	return global.InitialVirtualQuoteReserves, nil
}

// PumpQuoteSeed is the virtual quote and depth create_v2 gives a coin quoted in a pump coin.
type PumpQuoteSeed struct {
	VirtualQuoteReserves uint64
	Depth                int
}

// PumpQuoteReserves derives the virtual quote reserves of a coin quoted in pump coin Q.
func PumpQuoteReserves(global *Global, quote *PumpQuoteCurve, quoteControlKnown bool, quoteControl *pump.QuoteControl) (PumpQuoteSeed, error) {
	if quote == nil {
		return PumpQuoteSeed{}, errors.New("pump quote is nil")
	}
	admission, err := pumpQuoteAdmission(global, quote, quoteControlKnown, quoteControl)
	if err != nil {
		return PumpQuoteSeed{}, err
	}
	hops := make([]func(*big.Int) (*big.Int, error), len(admission.coins))
	for i, coin := range admission.coins {
		hop, err := chainCoinHop(global, coin)
		if err != nil {
			return PumpQuoteSeed{}, err
		}
		hops[i] = hop
	}
	tradable := new(big.Int).Sub(u64(global.InitialVirtualTokenReserves), u64(global.InitialRealTokenReserves))
	if tradable.Sign() <= 0 {
		return PumpQuoteSeed{}, errors.New("tradable token supply is zero")
	}
	raise := new(big.Int).Mul(u64(admission.target), u64(global.InitialRealTokenReserves))
	raise.Quo(raise, tradable)
	for i := len(hops) - 1; i >= 0; i-- {
		raise, err = hops[i](raise)
		if err != nil {
			return PumpQuoteSeed{}, err
		}
	}
	derived := new(big.Int).Mul(raise, tradable)
	derived.Quo(derived, u64(global.InitialRealTokenReserves))
	if derived.Cmp(big.NewInt(1)) < 0 {
		return PumpQuoteSeed{}, &QuoteReservesOutOfRangeError{QuoteMint: quote.Mint, Derived: derived}
	}
	reserves, err := fitUint64(derived)
	if err != nil {
		return PumpQuoteSeed{}, fmt.Errorf("derived virtual quote reserves overflow uint64: %s", derived)
	}
	return PumpQuoteSeed{VirtualQuoteReserves: reserves, Depth: admission.depth}, nil
}

// NewBondingCurve is the curve create_v2 would initialize for the quote mint.
func NewBondingCurve(global *Global, params CurveSeedParams) (*BondingCurve, error) {
	if global == nil {
		return nil, errors.New("global is nil")
	}
	virtualQuote, err := InitialVirtualQuoteReservesFor(global, params.QuoteMint, params.QuoteControlKnown, params.QuoteControl, params.PumpQuote)
	if err != nil {
		return nil, err
	}
	listed := params.QuoteControl
	if !params.QuoteControlKnown && params.PumpQuote != nil {
		listed = params.PumpQuote.QuoteControl
	}
	_, inControl := quoteControlReserves(listed, params.QuoteMint)
	isPumpQuoted := params.PumpQuote != nil &&
		!IsLegacyQuoteMint(params.QuoteMint) &&
		!quoteWhitelisted(global, params.QuoteMint) &&
		!inControl
	var depth uint8
	if isPumpQuoted {
		next := int(params.PumpQuote.BondingCurve.Depth) + 1
		if next > 255 {
			return nil, errors.New("curve depth overflows uint8")
		}
		depth = uint8(next)
	}
	var creatorFee uint64
	if params.CreatorFeeBpsSet {
		creatorFee = params.CreatorFeeBps
	}
	return curveFromSeed(global, params.QuoteMint, virtualQuote, depth, creatorFee, params.IsHolderReward), nil
}

// BuyTokenAmountFromQuote is the base bought by a quote budget, fees included.
func BuyTokenAmountFromQuote(params TradeQuoteParams) (uint64, error) {
	amount, err := helpers.TransferFeeNetOf(params.QuoteTransferFee, params.Amount)
	if err != nil {
		return 0, err
	}
	if amount == 0 {
		return 0, nil
	}
	curve, mintSupply, isNew, err := resolveTradeCurve(params)
	if err != nil {
		return 0, err
	}
	if curve.VirtualTokenReserves == 0 {
		return 0, nil
	}
	fees, err := helpers.ComputeFeesBps(params.Global, params.FeeConfig, mintSupply, curve.VirtualQuoteReserves, curve.VirtualTokenReserves, curve.QuoteMint, curve.CreatorFeeBps)
	if err != nil {
		return 0, err
	}
	total := u64(fees.ProtocolFeeBps)
	if isNew || !curve.Creator.Equals(solana.PublicKey{}) {
		total = new(big.Int).Add(total, u64(fees.CreatorFeeBps))
	}
	input := new(big.Int).Mul(u64(amount-1), big.NewInt(10_000))
	input.Quo(input, new(big.Int).Add(total, big.NewInt(10_000)))
	input64, err := fitUint64(input)
	if err != nil {
		return 0, err
	}
	tokens, err := math.GetBuyTokenAmountFromSolAmountQuote(input64, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)
	if err != nil {
		return 0, err
	}
	if tokens > curve.RealTokenReserves {
		return curve.RealTokenReserves, nil
	}
	return tokens, nil
}

// BuyQuoteAmountFromTokenAmount is the quote a buy of amount base costs, fees included.
func BuyQuoteAmountFromTokenAmount(params TradeQuoteParams) (uint64, error) {
	if params.Amount == 0 {
		return 0, nil
	}
	curve, mintSupply, isNew, err := resolveTradeCurve(params)
	if err != nil {
		return 0, err
	}
	if curve.VirtualTokenReserves == 0 {
		return 0, nil
	}
	minAmount := params.Amount
	if minAmount > curve.RealTokenReserves {
		minAmount = curve.RealTokenReserves
	}
	solCost, err := math.GetBuySolAmountFromTokenAmountQuote(minAmount, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)
	if err != nil {
		return 0, err
	}
	fee, err := helpers.GetFee(params.Global, params.FeeConfig, mintSupply, curve, solCost, isNew)
	if err != nil {
		return 0, err
	}
	gross, err := helpers.TransferFeeGrossFor(params.QuoteTransferFee, solCost)
	if err != nil {
		return 0, err
	}
	return addUint(gross, fee)
}

// SellQuoteAmountFromTokenAmount is the quote a sell of amount base lands in the wallet.
func SellQuoteAmountFromTokenAmount(params TradeQuoteParams) (uint64, error) {
	if params.Amount == 0 {
		return 0, nil
	}
	curve := params.BondingCurve
	if curve == nil {
		return 0, errors.New("bonding curve is nil")
	}
	if curve.VirtualTokenReserves == 0 {
		return 0, nil
	}
	solCost, err := math.GetSellSolAmountFromTokenAmountQuote(params.Amount, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)
	if err != nil {
		return 0, err
	}
	fee, err := helpers.GetFee(params.Global, params.FeeConfig, params.MintSupply, curve, solCost, false)
	if err != nil {
		return 0, err
	}
	net := uint64(0)
	if solCost > fee {
		net = solCost - fee
	}
	return helpers.TransferFeeNetOf(params.QuoteTransferFee, net)
}

// MigrationPoolLiquidity is the pool migrate_v2 deposits once curveTokens are bought for curveNetQuote.
func MigrationPoolLiquidity(global *Global, curve *BondingCurve, curveBaseTokenBalance, curveTokens, curveNetQuote uint64) (MigrationLiquidity, error) {
	if global == nil || curve == nil {
		return MigrationLiquidity{}, errors.New("global and bonding curve are required")
	}
	raised := new(big.Int).Add(u64(curve.RealQuoteReserves), u64(curveNetQuote))
	base := new(big.Int).Sub(u64(curveBaseTokenBalance), u64(curveTokens))
	quote := new(big.Int).Set(raised)
	if IsLegacyQuoteMint(curve.QuoteMint) {
		quote = new(big.Int).Sub(quote, u64(global.PoolMigrationFee))
	}
	if base.Sign() <= 0 || quote.Sign() <= 0 {
		return MigrationLiquidity{}, fmt.Errorf("empty pool-to-be (base %s, quote %s): check curveBaseTokenBalance", base, quote)
	}
	base64, err := fitUint64(base)
	if err != nil {
		return MigrationLiquidity{}, err
	}
	quote64, err := fitUint64(quote)
	if err != nil {
		return MigrationLiquidity{}, err
	}
	return MigrationLiquidity{Base: base64, Quote: quote64}, nil
}

// BuyV3QuoteAmountFromTokenAmount is what buy_v3 charges for amount base, fees included.
func BuyV3QuoteAmountFromTokenAmount(params V3QuoteParams) (uint64, error) {
	curve := params.BondingCurve
	if params.Global == nil || curve == nil {
		return 0, errors.New("global and bonding curve are required")
	}
	if params.FeeConfig == nil {
		return 0, errors.New("fee config is required")
	}
	if params.Amount == 0 || curve.Complete || curve.VirtualTokenReserves == 0 {
		return 0, nil
	}
	rates, err := v3RatesOf(params.Global, params.FeeConfig, params.MintSupply, curve)
	if err != nil {
		return 0, err
	}
	remaining := curve.RealTokenReserves
	curveTokens := params.Amount
	if curveTokens > remaining {
		curveTokens = remaining
	}
	curveNet, err := math.GetBuySolAmountFromTokenAmountQuote(curveTokens, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)
	if err != nil {
		return 0, err
	}
	if params.Amount <= remaining || curve.IsMayhemMode {
		return v3BuyDebit(params.Global, curve, rates, params.QuoteTransferFee, u64(curveNet), nil)
	}
	past := params.Amount - remaining
	pool, err := MigrationPoolLiquidity(params.Global, curve, params.CurveBaseTokenBalance, remaining, curveNet)
	if err != nil {
		return 0, err
	}
	if pool.Base <= past {
		return 0, fmt.Errorf("buy_v3 of %d tokens takes the whole pool-to-be (NotEnoughTokensToBuy)", params.Amount)
	}
	legNet := ceilDivBig(new(big.Int).Mul(u64(pool.Quote), u64(past)), u64(pool.Base-past))
	return v3BuyDebit(params.Global, curve, rates, params.QuoteTransferFee, u64(curveNet), legNet)
}

// BuyV3TokenAmountFromQuoteAmount is the base buy_exact_quote_in_v3 buys for a quote budget.
func BuyV3TokenAmountFromQuoteAmount(params V3QuoteParams) (uint64, error) {
	curve := params.BondingCurve
	if params.Global == nil || curve == nil {
		return 0, errors.New("global and bonding curve are required")
	}
	if params.FeeConfig == nil {
		return 0, errors.New("fee config is required")
	}
	if params.Amount == 0 || curve.Complete || curve.VirtualTokenReserves == 0 {
		return 0, nil
	}
	rates, err := v3RatesOf(params.Global, params.FeeConfig, params.MintSupply, curve)
	if err != nil {
		return 0, err
	}
	spendable, err := helpers.TransferFeeNetOf(params.QuoteTransferFee, params.Amount)
	if err != nil {
		return 0, err
	}
	net := v3ExactQuoteInNet(u64(spendable), rates)
	if net.Cmp(big.NewInt(1)) <= 0 {
		return 0, nil
	}
	tokensFor := func(quoteNet *big.Int) (uint64, error) {
		in := new(big.Int).Sub(new(big.Int).Set(quoteNet), big.NewInt(1))
		if in.Sign() < 0 {
			return 0, nil
		}
		in64, err := fitUint64(in)
		if err != nil {
			return 0, err
		}
		return math.GetBuyTokenAmountFromSolAmountQuote(in64, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)
	}
	tokens, err := tokensFor(net)
	if err != nil {
		return 0, err
	}
	remaining := curve.RealTokenReserves
	if tokens <= remaining || curve.IsMayhemMode {
		debit, err := v3BuyDebit(params.Global, curve, rates, params.QuoteTransferFee, net, nil)
		if err != nil {
			return 0, err
		}
		if debit > params.Amount {
			net = new(big.Int).Sub(net, u64(debit-params.Amount))
			if net.Cmp(big.NewInt(1)) <= 0 {
				return 0, nil
			}
			tokens, err = tokensFor(net)
			if err != nil {
				return 0, err
			}
		}
		if tokens > remaining {
			return remaining, nil
		}
		return tokens, nil
	}

	curveNet, err := math.GetBuySolAmountFromTokenAmountQuote(remaining, curve.VirtualTokenReserves, curve.VirtualQuoteReserves)
	if err != nil {
		return 0, err
	}
	remainderDebit, err := v3BuyDebit(params.Global, curve, rates, params.QuoteTransferFee, u64(curveNet), nil)
	if err != nil {
		return 0, err
	}
	if remainderDebit > params.Amount {
		return 0, fmt.Errorf("a budget of %d crosses the remaining supply but does not cover the remainder's debit with the quote's transfer fee (TooMuchSolRequired)", params.Amount)
	}
	spent := new(big.Int).Add(u64(curveNet), v3Fees(u64(curveNet), rates))
	if u64(spendable).Cmp(spent) <= 0 {
		return remaining, nil
	}
	budgetLeft := new(big.Int).Sub(u64(spendable), spent)
	totalBps := new(big.Int).Add(u64(rates.protocol), u64(rates.creator))
	afterFees := new(big.Int).Set(budgetLeft)
	if totalBps.Sign() != 0 {
		afterFees = new(big.Int).Mul(budgetLeft, big.NewInt(10_000))
		afterFees.Quo(afterFees, new(big.Int).Add(totalBps, big.NewInt(10_000)))
	}
	if afterFees.Cmp(big.NewInt(2)) < 0 {
		return remaining, nil
	}
	legNet := v3ExactQuoteInNet(budgetLeft, rates)
	pool, err := MigrationPoolLiquidity(params.Global, curve, params.CurveBaseTokenBalance, remaining, curveNet)
	if err != nil {
		return 0, err
	}
	legTokensFor := func(quoteIn *big.Int) *big.Int {
		in := new(big.Int).Sub(new(big.Int).Set(quoteIn), big.NewInt(1))
		if in.Sign() <= 0 {
			return big.NewInt(0)
		}
		den := new(big.Int).Add(u64(pool.Quote), in)
		num := new(big.Int).Mul(in, u64(pool.Base))
		return num.Quo(num, den)
	}
	legTokens := legTokensFor(legNet)
	for {
		debit, err := v3BuyDebit(params.Global, curve, rates, params.QuoteTransferFee, u64(curveNet), legNet)
		if err != nil {
			return 0, err
		}
		if debit <= params.Amount {
			break
		}
		legNet = new(big.Int).Sub(legNet, u64(debit-params.Amount))
		if legNet.Cmp(big.NewInt(1)) <= 0 {
			return remaining, nil
		}
		legTokens = legTokensFor(legNet)
		if legTokens.Sign() == 0 {
			return remaining, nil
		}
	}
	total := new(big.Int).Add(u64(remaining), legTokens)
	return fitUint64(total)
}

type v3Rate struct {
	protocol uint64
	creator  uint64
}

func v3RatesOf(global *Global, feeConfig *FeeConfig, mintSupply uint64, curve *BondingCurve) (v3Rate, error) {
	supply := helpers.OneBillionSupplyU64
	if curve.IsMayhemMode {
		supply = mintSupply
	}
	fees, err := helpers.ComputeFeesBps(global, feeConfig, supply, curve.VirtualQuoteReserves, curve.VirtualTokenReserves, curve.QuoteMint, curve.CreatorFeeBps)
	if err != nil {
		return v3Rate{}, err
	}
	creator := fees.CreatorFeeBps
	if curve.Creator.Equals(solana.PublicKey{}) {
		creator = 0
	}
	return v3Rate{protocol: fees.ProtocolFeeBps, creator: creator}, nil
}

func v3Fees(amount *big.Int, rates v3Rate) *big.Int {
	return new(big.Int).Add(ceilFee(amount, rates.protocol), ceilFee(amount, rates.creator))
}

func v3BuybackSlice(global *Global, curve *BondingCurve, net *big.Int, rates v3Rate) *big.Int {
	if curve.IsMayhemMode {
		return big.NewInt(0)
	}
	slice := new(big.Int).Mul(ceilFee(net, rates.protocol), u64(global.BuybackBasisPoints))
	return slice.Quo(slice, big.NewInt(10_000))
}

func v3BuyDebit(global *Global, curve *BondingCurve, rates v3Rate, fee *QuoteTransferFee, net, legNet *big.Int) (uint64, error) {
	buyback := v3BuybackSlice(global, curve, net, rates)
	if legNet != nil {
		buyback = new(big.Int).Add(buyback, v3BuybackSlice(global, curve, legNet, rates))
	}
	curveBound := new(big.Int).Add(new(big.Int).Set(net), v3Fees(net, rates))
	if legNet != nil {
		curveBound = new(big.Int).Add(curveBound, new(big.Int).Add(new(big.Int).Set(legNet), v3Fees(legNet, rates)))
	}
	curveBound = new(big.Int).Sub(curveBound, buyback)
	if curveBound.Sign() < 0 {
		return 0, errors.New("v3 buy debit underflow")
	}
	curve64, err := fitUint64(curveBound)
	if err != nil {
		return 0, err
	}
	gross, err := helpers.TransferFeeGrossFor(fee, curve64)
	if err != nil {
		return 0, err
	}
	buyback64, err := fitUint64(buyback)
	if err != nil {
		return 0, err
	}
	return addUint(gross, buyback64)
}

func v3ExactQuoteInNet(budget *big.Int, rates v3Rate) *big.Int {
	total := new(big.Int).Add(u64(rates.protocol), u64(rates.creator))
	var net *big.Int
	if total.Sign() == 0 {
		net = new(big.Int).Set(budget)
	} else {
		net = new(big.Int).Mul(new(big.Int).Set(budget), big.NewInt(10_000))
		net.Quo(net, new(big.Int).Add(total, big.NewInt(10_000)))
	}
	cost := new(big.Int).Add(new(big.Int).Set(net), v3Fees(net, rates))
	if cost.Cmp(budget) > 0 {
		net = new(big.Int).Sub(net, new(big.Int).Sub(cost, budget))
	}
	return net
}

func resolveTradeCurve(params TradeQuoteParams) (*BondingCurve, uint64, bool, error) {
	curve := params.BondingCurve
	mintSupply := params.MintSupply
	if curve != nil && params.MintSupplySet {
		return curve, mintSupply, false, nil
	}
	seeded, err := NewBondingCurve(params.Global, CurveSeedParams{
		QuoteMint:         params.QuoteMint,
		QuoteControlKnown: params.QuoteControlKnown,
		QuoteControl:      params.QuoteControl,
		CreatorFeeBpsSet:  params.CreatorFeeBpsSet,
		CreatorFeeBps:     params.CreatorFeeBps,
		PumpQuote:         params.PumpQuote,
	})
	if err != nil {
		return nil, 0, false, err
	}
	if params.Global == nil {
		return nil, 0, false, errors.New("global is nil")
	}
	return seeded, params.Global.TokenTotalSupply, true, nil
}

func curveFromSeed(global *Global, quoteMint solana.PublicKey, virtualQuote uint64, depth uint8, creatorFeeBps uint64, isHolderReward bool) *BondingCurve {
	stored := quoteMint
	if IsLegacyQuoteMint(quoteMint) {
		stored = solana.PublicKey{}
	}
	return &BondingCurve{
		VirtualTokenReserves:        global.InitialVirtualTokenReserves,
		VirtualQuoteReserves:        virtualQuote,
		RealTokenReserves:           global.InitialRealTokenReserves,
		RealQuoteReserves:           0,
		TokenTotalSupply:            global.TokenTotalSupply,
		Complete:                    false,
		Creator:                     solana.PublicKey{},
		IsMayhemMode:                global.MayhemModeEnabled,
		IsCashbackCoin:              false,
		QuoteMint:                   stored,
		CreatorFeeBps:               creatorFeeBps,
		CanEditCreatorFee:           false,
		IsHolderReward:              isHolderReward,
		Depth:                       depth,
		InitialVirtualQuoteReserves: virtualQuote,
	}
}

type pumpAdmission struct {
	depth  int
	target uint64
	coins  []PumpQuoteChainCoin
}

func pumpQuoteAdmission(global *Global, quote *PumpQuoteCurve, quoteControlKnown bool, quoteControl *pump.QuoteControl) (pumpAdmission, error) {
	depth, err := pumpQuoteGate(global, quote.Mint, quote.BondingCurve)
	if err != nil {
		return pumpAdmission{}, err
	}
	coins, err := pumpQuoteChain(quote)
	if err != nil {
		return pumpAdmission{}, err
	}
	control := quoteControl
	if !quoteControlKnown {
		if quote.QuoteControlKnown {
			control = quote.QuoteControl
		} else {
			control = nil
		}
	}
	var target uint64
	for i, coin := range coins {
		if i == len(coins)-1 {
			target, err = rootTarget(global, coin, control)
			if err != nil {
				return pumpAdmission{}, err
			}
		}
		if !IsBondingCurveMigrated(coin.BondingCurve) && coin.BondingCurve.Complete {
			return pumpAdmission{}, &QuoteCurveAwaitingMigrationError{QuoteMint: coin.Mint}
		}
	}
	return pumpAdmission{depth: depth, target: target, coins: coins}, nil
}

// PumpQuoteGate is create_v2's depth and mayhem check for a pump-coin quote.
// The returned depth is the new coin's depth (the quote coin's depth plus one).
func PumpQuoteGate(global *Global, mint solana.PublicKey, curve *BondingCurve) (int, error) {
	return pumpQuoteGate(global, mint, curve)
}

// PumpQuoteAdmission is the curve-level admission of a pump-coin quote: depth,
// the root target price, and the chain from the quote coin down to the root.
func PumpQuoteAdmission(global *Global, quote *PumpQuoteCurve, quoteControlKnown bool, quoteControl *pump.QuoteControl) (PumpQuoteAdmissionResult, error) {
	admission, err := pumpQuoteAdmission(global, quote, quoteControlKnown, quoteControl)
	if err != nil {
		return PumpQuoteAdmissionResult{}, err
	}
	return PumpQuoteAdmissionResult{Depth: admission.depth, Target: admission.target, Coins: admission.coins}, nil
}

// PumpQuoteAdmissionResult is what PumpQuoteAdmission returns.
type PumpQuoteAdmissionResult struct {
	Depth  int
	Target uint64
	Coins  []PumpQuoteChainCoin
}

func pumpQuoteGate(global *Global, mint solana.PublicKey, curve *BondingCurve) (int, error) {
	if global == nil || curve == nil {
		return 0, errors.New("global and bonding curve are required")
	}
	depth := int(curve.Depth) + 1
	max := int(global.MaxCurveDepth)
	if max > MaxCurveDepth {
		max = MaxCurveDepth
	}
	if depth > max {
		return 0, &CurveDepthExceededError{Depth: depth, MaxCurveDepth: max}
	}
	if curve.IsMayhemMode {
		return 0, &QuoteBondingCurveNotEligibleError{QuoteMint: mint}
	}
	return depth, nil
}

func pumpQuoteChain(quote *PumpQuoteCurve) ([]PumpQuoteChainCoin, error) {
	if quote.BondingCurve == nil {
		return nil, errors.New("pump quote bonding curve is nil")
	}
	coins := []PumpQuoteChainCoin{{
		Mint:         quote.Mint,
		BondingCurve: quote.BondingCurve,
		Pool:         quote.Pool,
	}}
	for i := 0; i < int(quote.BondingCurve.Depth); i++ {
		expected := coins[i].BondingCurve.QuoteMint
		if i >= len(quote.Chain) || quote.Chain[i].BondingCurve == nil {
			return nil, fmt.Errorf("pump coin %s is nested: the curve of %s is required (QuoteChainAccountsRequired)", quote.Mint, expected)
		}
		coin := quote.Chain[i]
		if !coin.Mint.Equals(expected) {
			return nil, fmt.Errorf("quote chain of %s: coin %d is %s, expected %s", quote.Mint, i+1, coin.Mint, expected)
		}
		coins = append(coins, coin)
	}
	return coins, nil
}

func rootTarget(global *Global, root PumpQuoteChainCoin, quoteControl *pump.QuoteControl) (uint64, error) {
	curve := root.BondingCurve
	if curve.QuoteMint.Equals(solana.PublicKey{}) {
		return global.InitialVirtualSolReserves, nil
	}
	if quoteWhitelisted(global, curve.QuoteMint) {
		return global.InitialVirtualQuoteReserves, nil
	}
	if reserves, ok := quoteControlReserves(quoteControl, curve.QuoteMint); ok {
		return reserves, nil
	}
	return 0, &QuoteBondingCurveNotEligibleError{QuoteMint: root.Mint}
}

func chainCoinHop(global *Global, coin PumpQuoteChainCoin) (func(*big.Int) (*big.Int, error), error) {
	curve := coin.BondingCurve
	if curve == nil {
		return nil, fmt.Errorf("pump coin %s: bonding curve is nil", coin.Mint)
	}
	if IsBondingCurveMigrated(curve) {
		if coin.Pool == nil || coin.Pool.BaseReserves == nil || coin.Pool.QuoteReserves == nil || coin.Pool.VirtualQuoteReserves == nil {
			return nil, fmt.Errorf("pump coin %s has migrated: its pool reserves are required", coin.Mint)
		}
		quoteSide := new(big.Int).Add(new(big.Int).Set(coin.Pool.QuoteReserves), coin.Pool.VirtualQuoteReserves)
		if quoteSide.Sign() < 0 {
			return nil, fmt.Errorf("pump coin %s: negative effective pool quote reserves (Overflow)", coin.Mint)
		}
		base := new(big.Int).Set(coin.Pool.BaseReserves)
		return func(quoteIn *big.Int) (*big.Int, error) {
			return chainBuy(coin.Mint, base, quoteSide, quoteIn)
		}, nil
	}
	vToken := u64(curve.VirtualTokenReserves)
	vQuote := u64(curve.VirtualQuoteReserves)
	remaining := u64(curve.RealTokenReserves)
	return func(quoteIn *big.Int) (*big.Int, error) {
		onCurve, err := chainBuy(coin.Mint, vToken, vQuote, quoteIn)
		if err != nil {
			return nil, err
		}
		if onCurve.Cmp(remaining) <= 0 {
			return onCurve, nil
		}
		remainderPrice, err := buyQuoteForBase(remaining, vToken, vQuote)
		if err != nil {
			return nil, err
		}
		budget := new(big.Int).Sub(new(big.Int).Set(quoteIn), remainderPrice)
		if budget.Cmp(big.NewInt(2)) < 0 {
			return new(big.Int).Set(remaining), nil
		}
		poolQuote := new(big.Int).Add(u64(curve.RealQuoteReserves), remainderPrice)
		if curve.QuoteMint.Equals(solana.PublicKey{}) {
			poolQuote = new(big.Int).Sub(poolQuote, u64(global.PoolMigrationFee))
		}
		if poolQuote.Sign() < 0 {
			return nil, fmt.Errorf("pump coin %s: the pool-to-be's raise is below the migration fee (Overflow)", coin.Mint)
		}
		poolBase := new(big.Int).Sub(u64(curve.TokenTotalSupply), u64(global.InitialRealTokenReserves))
		bought, err := chainBuy(coin.Mint, poolBase, poolQuote, new(big.Int).Sub(budget, big.NewInt(1)))
		if err != nil {
			return nil, err
		}
		return bought.Add(bought, remaining), nil
	}, nil
}

func chainBuy(mint solana.PublicKey, base, quoteSide, quoteIn *big.Int) (*big.Int, error) {
	if quoteSide.Sign() <= 0 {
		return nil, fmt.Errorf("pump coin %s: zero quote reserves (DivisionByZero)", mint)
	}
	den := new(big.Int).Add(new(big.Int).Set(quoteSide), quoteIn)
	num := new(big.Int).Mul(new(big.Int).Set(quoteIn), base)
	return num.Quo(num, den), nil
}

func buyQuoteForBase(minAmount, vToken, vQuote *big.Int) (*big.Int, error) {
	den := new(big.Int).Sub(new(big.Int).Set(vToken), minAmount)
	if den.Sign() <= 0 {
		return nil, errors.New("division by zero: virtual token reserves")
	}
	num := new(big.Int).Mul(new(big.Int).Set(minAmount), vQuote)
	num.Quo(num, den)
	return num.Add(num, big.NewInt(1)), nil
}

func quoteWhitelisted(global *Global, mint solana.PublicKey) bool {
	if mint.Equals(solana.PublicKey{}) {
		return false
	}
	for _, candidate := range global.WhitelistedQuoteMints {
		if candidate.Equals(mint) {
			return true
		}
	}
	return false
}

func quoteControlReserves(control *pump.QuoteControl, mint solana.PublicKey) (uint64, bool) {
	if control == nil {
		return 0, false
	}
	for _, entry := range control.Mints {
		if entry.Mint.Equals(mint) {
			return entry.InitialVirtualQuoteReserves, true
		}
	}
	return 0, false
}

func ceilFee(amount *big.Int, bps uint64) *big.Int {
	if amount == nil || amount.Sign() == 0 || bps == 0 {
		return big.NewInt(0)
	}
	n := new(big.Int).Mul(new(big.Int).Set(amount), new(big.Int).SetUint64(bps))
	n.Add(n, big.NewInt(9_999))
	return n.Quo(n, big.NewInt(10_000))
}

func ceilDivBig(a, b *big.Int) *big.Int {
	n := new(big.Int).Add(a, b)
	n.Sub(n, big.NewInt(1))
	return n.Quo(n, b)
}

func u64(v uint64) *big.Int {
	return new(big.Int).SetUint64(v)
}

func fitUint64(v *big.Int) (uint64, error) {
	if v == nil || v.Sign() < 0 || !v.IsUint64() {
		return 0, fmt.Errorf("amount overflows uint64")
	}
	return v.Uint64(), nil
}

func addUint(a, b uint64) (uint64, error) {
	sum := new(big.Int).Add(u64(a), u64(b))
	return fitUint64(sum)
}
