package helpers

import (
	"errors"
	"math/big"
)

// QuoteTransferFee is the Token-2022 transfer fee in force for a quote mint.
type QuoteTransferFee struct {
	// BasisPoints is withheld from every transfer, at most 10_000.
	BasisPoints uint64
	// MaximumFee caps one transfer, in the quote's base units.
	MaximumFee uint64
}

const transferFeeBps = uint64(10_000)

// TransferFeeWithheld is what Token-2022 keeps when amount is sent:
// ceil(amount * bps / 10_000), capped at MaximumFee.
func TransferFeeWithheld(fee *QuoteTransferFee, amount uint64) (uint64, error) {
	if fee == nil || fee.BasisPoints == 0 || amount == 0 {
		return 0, nil
	}
	raw, err := ceilDivBig(
		new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(fee.BasisPoints)),
		new(big.Int).SetUint64(transferFeeBps),
	)
	if err != nil {
		return 0, err
	}
	if raw > fee.MaximumFee {
		return fee.MaximumFee, nil
	}
	return raw, nil
}

// TransferFeeNetOf is what lands when amount is sent.
func TransferFeeNetOf(fee *QuoteTransferFee, amount uint64) (uint64, error) {
	withheld, err := TransferFeeWithheld(fee, amount)
	if err != nil {
		return 0, err
	}
	if withheld > amount {
		return 0, nil
	}
	return amount - withheld, nil
}

// TransferFeeGrossFor is the amount to send so at least net lands.
func TransferFeeGrossFor(fee *QuoteTransferFee, net uint64) (uint64, error) {
	if fee == nil || fee.BasisPoints == 0 || net == 0 {
		return net, nil
	}
	if fee.BasisPoints == transferFeeBps {
		return addUint64(net, fee.MaximumFee)
	}
	if fee.BasisPoints > transferFeeBps {
		return 0, errors.New("transfer fee basis points exceed 10000")
	}
	denominator := transferFeeBps - fee.BasisPoints
	rawBig := new(big.Int).Mul(new(big.Int).SetUint64(net), new(big.Int).SetUint64(transferFeeBps))
	rawBig.Add(rawBig, new(big.Int).SetUint64(denominator-1))
	rawBig.Quo(rawBig, new(big.Int).SetUint64(denominator))
	if !rawBig.IsUint64() {
		return 0, errors.New("transfer fee gross-up overflows uint64")
	}
	raw := rawBig.Uint64()
	if raw < net {
		return 0, errors.New("transfer fee gross-up underflow")
	}
	if raw-net >= fee.MaximumFee {
		return addUint64(net, fee.MaximumFee)
	}
	return raw, nil
}

func ceilDivBig(a, b *big.Int) (uint64, error) {
	if b.Sign() == 0 {
		return 0, errors.New("division by zero")
	}
	n := new(big.Int).Add(a, b)
	n.Sub(n, big.NewInt(1))
	n.Quo(n, b)
	if n.Sign() < 0 || !n.IsUint64() {
		return 0, errors.New("ceil div overflow")
	}
	return n.Uint64(), nil
}

func addUint64(a, b uint64) (uint64, error) {
	sum := new(big.Int).Add(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
	if !sum.IsUint64() {
		return 0, errors.New("uint64 addition overflow")
	}
	return sum.Uint64(), nil
}
