package helpers

import (
	"errors"
	"math/big"

	"github.com/shopspring/decimal"
)

var BpsDenominator = decimal.NewFromUint64(10_000)
var SlippageScale = decimal.NewFromUint64(1_000_000_000)

func CeilDiv(a, b decimal.Decimal) (decimal.Decimal, error) {
	if b.Sign() == 0 {
		return decimal.Decimal{}, errors.New("cannot divide by zero")
	}
	return a.Add(b).Sub(decimal.NewFromInt(1)).Div(b), nil
}

func Fee(amount, basisPoints decimal.Decimal) (decimal.Decimal, error) {
	return CeilDiv(amount.Mul(basisPoints), BpsDenominator)
}

func SlippageFactor(slippage float64, isBuy bool) decimal.Decimal {
	v := 1.0 + slippage/100
	if !isBuy {
		v = 1.0 - slippage/100
	}

	f := new(big.Float).Mul(big.NewFloat(v), big.NewFloat(1_000_000_000))

	out, _ := f.Int(nil)
	return decimal.RequireFromString(out.String())
}
