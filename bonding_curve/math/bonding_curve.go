package math

import "github.com/shopspring/decimal"

func GetBuySolAmountFromTokenAmountQuote(minAmount, vToken, vQuote uint64) uint64 {
	minAmountDecimal := decimal.NewFromUint64(minAmount)
	vTokenDecimal := decimal.NewFromUint64(vToken)
	vQuoteDecimal := decimal.NewFromUint64(vQuote)

	return minAmountDecimal.Mul(vQuoteDecimal).Div(vTokenDecimal.Sub(minAmountDecimal)).Add(decimal.NewFromUint64(1)).Ceil().BigInt().Uint64()
}
func GetBuyTokenAmountFromSolAmountQuote(inputAmount, vToken, vQuote uint64) uint64 {
	inputAmountDecimal := decimal.NewFromUint64(inputAmount)
	vTokenDecimal := decimal.NewFromUint64(vToken)
	vQuoteDecimal := decimal.NewFromUint64(vQuote)

	return inputAmountDecimal.Mul(vTokenDecimal).Div(vQuoteDecimal.Add(inputAmountDecimal)).Ceil().BigInt().Uint64()
}
func GetSellSolAmountFromTokenAmountQuote(inputAmount, vToken, vQuote uint64) uint64 {
	inputAmountDecimal := decimal.NewFromUint64(inputAmount)
	vTokenDecimal := decimal.NewFromUint64(vToken)
	vQuoteDecimal := decimal.NewFromUint64(vQuote)

	return inputAmountDecimal.Mul(vQuoteDecimal).Div(vTokenDecimal.Add(inputAmountDecimal)).Ceil().BigInt().Uint64()
}
