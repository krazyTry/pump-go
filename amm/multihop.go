package amm

import (
	"errors"
	"fmt"
	"math/big"

	solana "github.com/solana-foundation/solana-go/v2"
)

// MultiHopLegFees is the fee components one hop of a route charges.
// LP is never set on a bonding-curve hop.
type MultiHopLegFees struct {
	Protocol bool
	Creator  bool
	LP       bool
}

// MultiHopRoute is the direction, mint chain and per-hop fees of a route.
type MultiHopRoute struct {
	IsBuy bool
	Mints []solana.PublicKey
	Legs  []MultiHopLegFees
}

// CurveQuote prices one bonding-curve hop the way pump's multi_hop_curve_swap does.
// Only the components legs names are charged, and a curve never charges an LP fee.
type CurveQuote func(amountIn *big.Int, isBuy bool, legs MultiHopLegFees) (*big.Int, error)

// MultiHopHop is one pool or curve hop, with the state a quote is priced from.
type MultiHopHop struct {
	// Kind is "pool" or "curve".
	Kind              string
	Pool              *Pool
	PoolKey           solana.PublicKey
	BaseMint          solana.PublicKey
	QuoteMint         solana.PublicKey
	PoolBaseAmount    uint64
	PoolQuoteAmount   uint64
	BaseMintSupply    uint64
	BaseTokenProgram  solana.PublicKey
	QuoteTokenProgram solana.PublicKey
	QuoteTransferFee  *QuoteTransferFee
	Quote             CurveQuote
}

// VenueQuoteTokenProgram is the token program of a hop's quote. A SOL curve always uses legacy SPL Token.
func VenueQuoteTokenProgram(hop MultiHopHop) solana.PublicKey {
	if IsSolCurve(hop) || hop.QuoteTokenProgram.Equals(solana.PublicKey{}) {
		return solana.TokenProgramID
	}
	return hop.QuoteTokenProgram
}

func venueBaseTokenProgram(hop MultiHopHop) solana.PublicKey {
	if hop.BaseTokenProgram.Equals(solana.PublicKey{}) {
		return solana.TokenProgramID
	}
	return hop.BaseTokenProgram
}

// MultiHopQuoteResult is the output of a route and the floor to build it with.
type MultiHopQuoteResult struct {
	AmountOut     *big.Int
	MinAmountOut  *big.Int
	HopAmountsOut []*big.Int
}

// IsSolCurve reports a bonding-curve hop quoted in SOL. The curve stores the
// zero key; the route names that quote as WSOL.
func IsSolCurve(hop MultiHopHop) bool {
	return hop.Kind == "curve" && CanonicalPoolQuoteMint(hop.QuoteMint).Equals(solana.WrappedSol)
}

func venueMints(hop MultiHopHop) (baseMint, quoteMint solana.PublicKey) {
	if hop.Kind == "pool" && hop.Pool != nil {
		return hop.Pool.BaseMint, hop.Pool.QuoteMint
	}
	return hop.BaseMint, CanonicalPoolQuoteMint(hop.QuoteMint)
}

// MultiHopLegFeesFor reports which fee components hop index of hops charges.
func MultiHopLegFeesFor(isBuy bool, index, hops int, isPool bool) MultiHopLegFees {
	first := index == 0
	last := index+1 == hops
	protocol, target := first, last
	if !isBuy {
		protocol, target = last, first
	}
	return MultiHopLegFees{Protocol: protocol, Creator: target, LP: target && isPool}
}

// ResolveMultiHopRoute walks hops from inMint and refuses a route the program would refuse.
func ResolveMultiHopRoute(inMint solana.PublicKey, hops []MultiHopHop) (MultiHopRoute, error) {
	n := len(hops)
	if n == 0 {
		return MultiHopRoute{}, errors.New("A multi-hop route needs at least one hop.")
	}
	pools := 0
	for _, hop := range hops {
		if hop.Kind == "pool" {
			pools++
		}
	}
	if n > MultiHopMaxHops || (n == MultiHopMaxHops && pools > MultiHopMaxPoolsAtMaxHops) {
		return MultiHopRoute{}, fmt.Errorf("A %d-hop route with %d pool hops exceeds the programs' heap (at most %d hops, %d of them pools at that length).", n, pools, MultiHopMaxHops, MultiHopMaxPoolsAtMaxHops)
	}

	inMint = CanonicalPoolQuoteMint(inMint)
	_, firstQuote := venueMints(hops[0])
	isBuy := inMint.Equals(firstQuote)
	mints := []solana.PublicKey{inMint}
	legs := make([]MultiHopLegFees, n)
	for i, hop := range hops {
		baseMint, quoteMint := venueMints(hop)
		running := mints[i]
		hopIsBuy := running.Equals(quoteMint)
		if !hopIsBuy && !running.Equals(baseMint) {
			return MultiHopRoute{}, fmt.Errorf("Hop %d trades %s/%s, not the running mint %s.", i, baseMint, quoteMint, running)
		}
		if hopIsBuy != isBuy {
			return MultiHopRoute{}, errors.New("Every hop of a multi-hop route must trade in the same direction.")
		}
		next := quoteMint
		if hopIsBuy {
			next = baseMint
		}
		mints = append(mints, next)

		leg := MultiHopLegFeesFor(isBuy, i, n, hop.Kind == "pool")
		if hop.Kind == "pool" {
			if hop.Pool == nil {
				return MultiHopRoute{}, fmt.Errorf("Hop %d: pool is nil.", i)
			}
			if !IsPumpPool(hop.Pool.BaseMint, hop.Pool.Creator) {
				return MultiHopRoute{}, fmt.Errorf("Hop %d: pool %s is not a canonical pump pool.", i, hop.PoolKey)
			}
			if hop.Pool.IsMayhemMode {
				return MultiHopRoute{}, fmt.Errorf("Hop %d: pool %s is a mayhem pool.", i, hop.PoolKey)
			}
			if hop.Pool.IsCashbackCoin && leg.Creator {
				return MultiHopRoute{}, fmt.Errorf("Hop %d: cashback pool %s cannot charge the creator fee.", i, hop.PoolKey)
			}
		}
		legs[i] = leg
	}
	return MultiHopRoute{IsBuy: isBuy, Mints: mints, Legs: legs}, nil
}

// MultiHopSwapQuote prices a route that spends amountIn of inMint.
// Each hop is priced from its state before the route.
func MultiHopSwapQuote(inMint solana.PublicKey, hops []MultiHopHop, amountIn *big.Int, slippage float64, globalConfig *GlobalConfig, feeConfig *FeeConfig) (MultiHopQuoteResult, error) {
	if amountIn == nil || amountIn.Sign() <= 0 {
		return MultiHopQuoteResult{}, errors.New("amountIn must be positive.")
	}
	route, err := ResolveMultiHopRoute(inMint, hops)
	if err != nil {
		return MultiHopQuoteResult{}, err
	}
	amount := new(big.Int).Set(amountIn)
	hopAmounts := make([]*big.Int, len(hops))
	for i, hop := range hops {
		if hop.Kind == "pool" {
			amount, err = poolHopAmountOut(hop, route.IsBuy, amount, route.Legs[i], globalConfig, feeConfig)
		} else {
			if hop.Quote == nil {
				return MultiHopQuoteResult{}, fmt.Errorf("Hop %d: curve quote is nil.", i)
			}
			amount, err = hop.Quote(amount, route.IsBuy, route.Legs[i])
		}
		if err != nil {
			return MultiHopQuoteResult{}, err
		}
		hopAmounts[i] = amount
	}
	minOut := applySlippage(amount, slippage, false)
	one := big.NewInt(1)
	if minOut.Cmp(one) < 0 {
		minOut = one
	}
	return MultiHopQuoteResult{
		AmountOut:     amount,
		MinAmountOut:  minOut,
		HopAmountsOut: hopAmounts,
	}, nil
}

func poolHopAmountOut(hop MultiHopHop, isBuy bool, amountIn *big.Int, legs MultiHopLegFees, globalConfig *GlobalConfig, feeConfig *FeeConfig) (*big.Int, error) {
	pool := hop.Pool
	baseReserve := u64(hop.PoolBaseAmount)
	quoteReserve := new(big.Int).Add(u64(hop.PoolQuoteAmount), pool.VirtualQuoteReserves.BigInt())
	rates, err := ComputeFeesBps(FeeParams{
		GlobalConfig:   globalConfig,
		FeeConfig:      feeConfig,
		Creator:        pool.Creator,
		BaseMintSupply: hop.BaseMintSupply,
		BaseMint:       pool.BaseMint,
		BaseReserve:    hop.PoolBaseAmount,
		QuoteReserve:   quoteReserve,
		QuoteMint:      pool.QuoteMint,
		IsMayhemMode:   pool.IsMayhemMode,
		CreatorFeeBps:  pool.CreatorFeeBps,
	})
	if err != nil {
		return nil, err
	}
	var lpBps, protocolBps, creatorRate uint64
	if legs.LP {
		lpBps = rates.LpFeeBps
	}
	if legs.Protocol {
		protocolBps = rates.ProtocolFeeBps
	}
	if legs.Creator && !pool.CoinCreator.Equals(solana.PublicKey{}) {
		creatorRate = rates.CreatorFeeBps
	}

	if isBuy {
		received, err := NetOfTransferFee(hop.QuoteTransferFee, amountIn)
		if err != nil {
			return nil, err
		}
		if received.Sign() <= 0 {
			return nil, errors.New("The quote budget does not cover the transfer fee.")
		}
		if !received.IsUint64() {
			return nil, errors.New("amount overflows uint64")
		}
		split, err := ExactQuoteInFees(received.Uint64(), lpBps, protocolBps, creatorRate)
		if err != nil {
			return nil, err
		}
		input := new(big.Int).Sub(received, u64(split.LpFee))
		input.Sub(input, u64(split.ProtocolFee))
		input.Sub(input, u64(split.CoinCreatorFee))
		priced := new(big.Int).Sub(input, big.NewInt(1))
		baseOut := new(big.Int).Mul(baseReserve, priced)
		baseOut.Quo(baseOut, new(big.Int).Add(quoteReserve, priced))
		if baseOut.Sign() <= 0 || baseOut.Cmp(baseReserve) > 0 {
			return nil, errors.New("The hop buys no base tokens.")
		}
		return baseOut, nil
	}

	quoteOut := new(big.Int).Mul(quoteReserve, amountIn)
	quoteOut.Quo(quoteOut, new(big.Int).Add(baseReserve, amountIn))
	if !quoteOut.IsUint64() {
		return nil, errors.New("amount overflows uint64")
	}
	lpFee, err := feeAmount(quoteOut, lpBps)
	if err != nil {
		return nil, err
	}
	protocolFee, err := feeAmount(quoteOut, protocolBps)
	if err != nil {
		return nil, err
	}
	creatorFee, err := feeAmount(quoteOut, creatorRate)
	if err != nil {
		return nil, err
	}
	real, err := RealQuoteReserves(hop.PoolQuoteAmount, pool.ProtocolFees, pool.CreatorFees)
	if err != nil {
		return nil, err
	}
	_, userOut, err := sellSplit(quoteOut, lpFee, protocolFee, creatorFee, real)
	if err != nil {
		return nil, err
	}
	if userOut.Sign() <= 0 {
		return nil, errors.New("The hop sells for no quote tokens.")
	}
	return NetOfTransferFee(hop.QuoteTransferFee, userOut)
}
