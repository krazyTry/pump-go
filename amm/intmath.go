package amm

import (
	"errors"
	"math"
	"math/big"
)

func u64(v uint64) *big.Int {
	return new(big.Int).SetUint64(v)
}

func fitUint64(n *big.Int) (uint64, error) {
	if n == nil || n.Sign() < 0 || !n.IsUint64() {
		return 0, errors.New("amount overflows uint64")
	}
	return n.Uint64(), nil
}

// ceilDiv is ceil(a / b) for a non-negative a and a positive b.
func ceilDiv(a, b *big.Int) (*big.Int, error) {
	if b == nil || b.Sign() <= 0 {
		return nil, errors.New("Cannot divide by zero.")
	}
	if a.Sign() < 0 {
		return nil, errors.New("ceilDiv of a negative amount is not supported")
	}
	n := new(big.Int).Add(a, b)
	n.Sub(n, big.NewInt(1))
	return n.Quo(n, new(big.Int).Set(b)), nil
}

// feeAmount is ceil(amount * bps / 10_000), the fee the program keeps.
func feeAmount(amount *big.Int, bps uint64) (*big.Int, error) {
	if amount == nil || amount.Sign() == 0 || bps == 0 {
		return big.NewInt(0), nil
	}
	num := new(big.Int).Mul(new(big.Int).Set(amount), new(big.Int).SetUint64(bps))
	return ceilDiv(num, big.NewInt(10_000))
}

// floorMulDiv is floor(amount * bps / 10_000). Buyback uses this, not the ceil fee.
func floorMulDiv(amount *big.Int, bps uint64) *big.Int {
	if amount == nil || amount.Sign() == 0 || bps == 0 {
		return big.NewInt(0)
	}
	num := new(big.Int).Mul(new(big.Int).Set(amount), new(big.Int).SetUint64(bps))
	return num.Quo(num, big.NewInt(10_000))
}

// slippageFactor matches the TypeScript SDK: floor((1 ± slippage/100) * 1e9).
func slippageFactor(slippage float64, increase bool) *big.Int {
	v := 1 + slippage/100
	if !increase {
		v = 1 - slippage/100
	}
	return big.NewInt(int64(math.Floor(v * 1_000_000_000)))
}

func applySlippage(amount *big.Int, slippage float64, increase bool) *big.Int {
	factor := slippageFactor(slippage, increase)
	return new(big.Int).Quo(new(big.Int).Mul(new(big.Int).Set(amount), factor), big.NewInt(1_000_000_000))
}

func nonNil(n *big.Int) *big.Int {
	if n == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(n)
}
