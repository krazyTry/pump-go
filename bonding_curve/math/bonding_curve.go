package math

import (
	"errors"
	"math/big"
)

// GetBuySolAmountFromTokenAmountQuote is the constant-product quote-in for
// `minAmount` base tokens, plus one unit, matching
// getBuySolAmountFromTokenAmountQuote in the TypeScript SDK.
func GetBuySolAmountFromTokenAmountQuote(minAmount, vToken, vQuote uint64) (uint64, error) {
	if vToken <= minAmount {
		return 0, errors.New("division by zero: virtual token reserves cannot be less than or equal to the token amount")
	}
	num := new(big.Int).Mul(new(big.Int).SetUint64(minAmount), new(big.Int).SetUint64(vQuote))
	den := new(big.Int).SetUint64(vToken - minAmount)
	num.Quo(num, den)
	num.Add(num, big.NewInt(1))
	if !num.IsUint64() {
		return 0, errors.New("quote cost overflows uint64")
	}
	return num.Uint64(), nil
}

// GetBuyTokenAmountFromSolAmountQuote is floor(input * vToken / (vQuote + input)).
func GetBuyTokenAmountFromSolAmountQuote(inputAmount, vToken, vQuote uint64) (uint64, error) {
	den := new(big.Int).Add(new(big.Int).SetUint64(vQuote), new(big.Int).SetUint64(inputAmount))
	if den.Sign() == 0 {
		return 0, errors.New("division by zero: quote reserves")
	}
	num := new(big.Int).Mul(new(big.Int).SetUint64(inputAmount), new(big.Int).SetUint64(vToken))
	num.Quo(num, den)
	if !num.IsUint64() {
		return 0, errors.New("token amount overflows uint64")
	}
	return num.Uint64(), nil
}

// GetSellSolAmountFromTokenAmountQuote is floor(input * vQuote / (vToken + input)).
func GetSellSolAmountFromTokenAmountQuote(inputAmount, vToken, vQuote uint64) (uint64, error) {
	den := new(big.Int).Add(new(big.Int).SetUint64(vToken), new(big.Int).SetUint64(inputAmount))
	if den.Sign() == 0 {
		return 0, errors.New("division by zero: token reserves")
	}
	num := new(big.Int).Mul(new(big.Int).SetUint64(inputAmount), new(big.Int).SetUint64(vQuote))
	num.Quo(num, den)
	if !num.IsUint64() {
		return 0, errors.New("quote proceeds overflow uint64")
	}
	return num.Uint64(), nil
}
