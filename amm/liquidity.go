package amm

import (
	"errors"
	"math/big"
)

// DepositToken0Params sizes a deposit from one side of the pool.
// Token0 is the side the caller names; token1 is the other side.
type DepositToken0Params struct {
	Token0        uint64
	Slippage      float64
	Token0Reserve uint64
	Token1Reserve uint64
	TotalLpTokens uint64
}

// DepositLpTokenParams sizes the tokens a deposit of lpToken must bring.
type DepositLpTokenParams struct {
	LpToken          uint64
	Slippage         float64
	BaseReserve      uint64
	QuoteReserve     uint64
	TotalLpTokens    uint64
	QuoteTransferFee *QuoteTransferFee
}

// WithdrawParams sizes the tokens a withdrawal of lp returns.
type WithdrawParamsCalc struct {
	LpAmount         uint64
	Slippage         float64
	BaseReserve      uint64
	QuoteReserve     uint64
	TotalLpTokens    uint64
	QuoteTransferFee *QuoteTransferFee
}

// DepositBaseParams is a base-denominated deposit. Quote is wallet-side.
type DepositBaseParams struct {
	Base             uint64
	Slippage         float64
	BaseReserve      uint64
	QuoteReserve     uint64
	TotalLpTokens    uint64
	QuoteTransferFee *QuoteTransferFee
}

// DepositQuoteParams is a quote-denominated deposit. Quote is wallet-side.
type DepositQuoteParams struct {
	Quote            uint64
	Slippage         float64
	BaseReserve      uint64
	QuoteReserve     uint64
	TotalLpTokens    uint64
	QuoteTransferFee *QuoteTransferFee
}

// DepositBaseResult is the wallet-side quote a base deposit debits.
type DepositBaseResult struct {
	Quote    uint64
	LpToken  uint64
	MaxBase  uint64
	MaxQuote uint64
}

// DepositQuoteResult is the base a quote deposit must match.
type DepositQuoteCalcResult struct {
	Base     uint64
	LpToken  uint64
	MaxBase  uint64
	MaxQuote uint64
}

func checkSlippage(slippage float64) error {
	if slippage < 0 || slippage > 100 {
		return errors.New("Slippage must be between 0 and 100 (0% to 100%)")
	}
	return nil
}

// DepositToken0 returns the other side, the LP minted, and the slippage caps.
func DepositToken0(p DepositToken0Params) (DepositResult, error) {
	if err := checkSlippage(p.Slippage); err != nil {
		return DepositResult{}, err
	}
	if p.Token0Reserve == 0 {
		return DepositResult{}, errors.New("Division by zero: token0 reserve cannot be zero")
	}
	token1 := new(big.Int).Mul(u64(p.Token0), u64(p.Token1Reserve))
	token1.Quo(token1, u64(p.Token0Reserve))
	lp := new(big.Int).Mul(u64(p.Token0), u64(p.TotalLpTokens))
	lp.Quo(lp, u64(p.Token0Reserve))

	max0, err := fitUint64(applySlippage(u64(p.Token0), p.Slippage, true))
	if err != nil {
		return DepositResult{}, err
	}
	max1, err := fitUint64(applySlippage(token1, p.Slippage, true))
	if err != nil {
		return DepositResult{}, err
	}
	token1U, err := fitUint64(token1)
	if err != nil {
		return DepositResult{}, err
	}
	lpU, err := fitUint64(lp)
	if err != nil {
		return DepositResult{}, err
	}
	return DepositResult{
		Token1:    token1U,
		LpToken:   lpU,
		MaxToken0: max0,
		MaxToken1: max1,
	}, nil
}

// DepositLpToken returns the maximum base and quote a deposit of lp may debit.
// The quote cap is grossed up for a quote transfer fee.
func DepositLpToken(p DepositLpTokenParams) (DepositLpTokenResult, error) {
	if p.TotalLpTokens == 0 {
		return DepositLpTokenResult{}, errors.New("Division by zero: totalLpTokens cannot be zero")
	}
	baseIn, err := ceilDiv(new(big.Int).Mul(u64(p.BaseReserve), u64(p.LpToken)), u64(p.TotalLpTokens))
	if err != nil {
		return DepositLpTokenResult{}, err
	}
	quotePool, err := ceilDiv(new(big.Int).Mul(u64(p.QuoteReserve), u64(p.LpToken)), u64(p.TotalLpTokens))
	if err != nil {
		return DepositLpTokenResult{}, err
	}
	quoteIn, err := GrossForTransferFee(p.QuoteTransferFee, quotePool)
	if err != nil {
		return DepositLpTokenResult{}, err
	}
	maxBase, err := fitUint64(applySlippage(baseIn, p.Slippage, true))
	if err != nil {
		return DepositLpTokenResult{}, err
	}
	maxQuote, err := fitUint64(applySlippage(quoteIn, p.Slippage, true))
	if err != nil {
		return DepositLpTokenResult{}, err
	}
	return DepositLpTokenResult{MaxBase: maxBase, MaxQuote: maxQuote}, nil
}

// Withdraw quotes the base and quote a burn of LP returns.
// Quote and minQuote are what lands after a quote transfer fee.
func Withdraw(p WithdrawParamsCalc) (WithdrawResult, error) {
	if p.LpAmount == 0 || p.TotalLpTokens == 0 {
		return WithdrawResult{}, errors.New("LP amount or total LP tokens cannot be zero.")
	}
	base := new(big.Int).Mul(u64(p.BaseReserve), u64(p.LpAmount))
	base.Quo(base, u64(p.TotalLpTokens))
	quotePool := new(big.Int).Mul(u64(p.QuoteReserve), u64(p.LpAmount))
	quotePool.Quo(quotePool, u64(p.TotalLpTokens))
	quote, err := NetOfTransferFee(p.QuoteTransferFee, quotePool)
	if err != nil {
		return WithdrawResult{}, err
	}
	if quote.Sign() < 0 {
		return WithdrawResult{}, errors.New("Fees exceed total output; final quote is negative.")
	}
	baseU, err := fitUint64(base)
	if err != nil {
		return WithdrawResult{}, err
	}
	quoteU, err := fitUint64(quote)
	if err != nil {
		return WithdrawResult{}, err
	}
	minBase, err := fitUint64(applySlippage(base, p.Slippage, false))
	if err != nil {
		return WithdrawResult{}, err
	}
	minQuote, err := fitUint64(applySlippage(quote, p.Slippage, false))
	if err != nil {
		return WithdrawResult{}, err
	}
	return WithdrawResult{
		Base:     baseU,
		Quote:    quoteU,
		MinBase:  minBase,
		MinQuote: minQuote,
	}, nil
}

// DepositBase quotes a deposit denominated in base.
// The returned quote is the wallet debit, grossed up for a transfer fee.
func DepositBase(p DepositBaseParams) (DepositBaseResult, error) {
	sized, err := DepositToken0(DepositToken0Params{
		Token0:        p.Base,
		Slippage:      p.Slippage,
		Token0Reserve: p.BaseReserve,
		Token1Reserve: p.QuoteReserve,
		TotalLpTokens: p.TotalLpTokens,
	})
	if err != nil {
		return DepositBaseResult{}, err
	}
	quote, err := GrossForTransferFee(p.QuoteTransferFee, u64(sized.Token1))
	if err != nil {
		return DepositBaseResult{}, err
	}
	maxQuote, err := GrossForTransferFee(p.QuoteTransferFee, u64(sized.MaxToken1))
	if err != nil {
		return DepositBaseResult{}, err
	}
	quoteU, err := fitUint64(quote)
	if err != nil {
		return DepositBaseResult{}, err
	}
	maxQuoteU, err := fitUint64(maxQuote)
	if err != nil {
		return DepositBaseResult{}, err
	}
	return DepositBaseResult{
		Quote:    quoteU,
		LpToken:  sized.LpToken,
		MaxBase:  sized.MaxToken0,
		MaxQuote: maxQuoteU,
	}, nil
}

// DepositQuote quotes a deposit denominated in quote.
// quote is what the wallet may debit; the pool is sized on what lands.
func DepositQuote(p DepositQuoteParams) (DepositQuoteCalcResult, error) {
	landed, err := NetOfTransferFee(p.QuoteTransferFee, u64(p.Quote))
	if err != nil {
		return DepositQuoteCalcResult{}, err
	}
	if landed.Sign() < 0 {
		return DepositQuoteCalcResult{}, errors.New("The quote budget does not cover the transfer fee.")
	}
	landedU, err := fitUint64(landed)
	if err != nil {
		return DepositQuoteCalcResult{}, err
	}
	sized, err := DepositToken0(DepositToken0Params{
		Token0:        landedU,
		Slippage:      p.Slippage,
		Token0Reserve: p.QuoteReserve,
		Token1Reserve: p.BaseReserve,
		TotalLpTokens: p.TotalLpTokens,
	})
	if err != nil {
		return DepositQuoteCalcResult{}, err
	}
	maxQuote, err := GrossForTransferFee(p.QuoteTransferFee, u64(sized.MaxToken0))
	if err != nil {
		return DepositQuoteCalcResult{}, err
	}
	maxQuoteU, err := fitUint64(maxQuote)
	if err != nil {
		return DepositQuoteCalcResult{}, err
	}
	return DepositQuoteCalcResult{
		Base:     sized.Token1,
		LpToken:  sized.LpToken,
		MaxBase:  sized.MaxToken1,
		MaxQuote: maxQuoteU,
	}, nil
}

// RealQuoteReserves is the vault balance deposits, withdrawals and sells can draw on.
// v2 trades leave protocol and creator fees in the vault; those buckets are not liquidity.
func RealQuoteReserves(vault, protocolFees, creatorFees uint64) (uint64, error) {
	sum := new(big.Int).Add(u64(protocolFees), u64(creatorFees))
	if u64(vault).Cmp(sum) < 0 {
		return 0, errors.New("Insufficient real quote reserves to cover the sell output.")
	}
	return new(big.Int).Sub(u64(vault), sum).Uint64(), nil
}
