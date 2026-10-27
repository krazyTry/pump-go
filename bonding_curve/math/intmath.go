package math

import (
	"errors"
	"math/big"
)

// MulDivFloor returns floor(a * b / d), matching bn.js integer division.
func MulDivFloor(a, b, d uint64) (uint64, error) {
	if d == 0 {
		return 0, errors.New("division by zero")
	}
	num := new(big.Int).Mul(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
	num.Quo(num, new(big.Int).SetUint64(d))
	if !num.IsUint64() {
		return 0, errors.New("muldiv overflow")
	}
	return num.Uint64(), nil
}

// CeilDiv returns ceil(a / b) = floor((a + b - 1) / b) for positive integers.
func CeilDiv(a, b uint64) (uint64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	num := new(big.Int).SetUint64(a)
	num.Add(num, new(big.Int).SetUint64(b-1))
	num.Quo(num, new(big.Int).SetUint64(b))
	if !num.IsUint64() {
		return 0, errors.New("ceil div overflow")
	}
	return num.Uint64(), nil
}

// ConstantProduct is virtualTokenReserves * virtualQuoteReserves.
func ConstantProduct(virtualToken, virtualQuote uint64) *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(virtualToken), new(big.Int).SetUint64(virtualQuote))
}
