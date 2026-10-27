package amm

import (
	"errors"
	"math/big"

	"github.com/krazyTry/pump-go/amm/helpers"
	solana "github.com/solana-foundation/solana-go/v2"
)

// QuoteTransferFee is the Token-2022 transfer fee a quote mint charges this epoch.
// A nil fee is an untaxed quote.
type QuoteTransferFee struct {
	BasisPoints uint64
	MaximumFee  uint64
}

// ParseTransferFeeConfig reads the TransferFeeConfig extension of a mint account.
// A legacy SPL mint, or a Token-2022 mint without the extension, returns nil.
func ParseTransferFeeConfig(mintOwner solana.PublicKey, data []byte) (*helpers.TransferFeeConfig, error) {
	if !mintOwner.Equals(solana.Token2022ProgramID) {
		return nil, nil
	}
	return helpers.ParseMintTransferFeeConfig(data)
}

// TransferFeeAt is the schedule entry in force at epoch.
func TransferFeeAt(cfg *helpers.TransferFeeConfig, epoch uint64) (QuoteTransferFee, error) {
	if cfg == nil {
		return QuoteTransferFee{}, errors.New("transfer fee config is nil")
	}
	fee := cfg.FeeForEpoch(epoch)
	return QuoteTransferFee{BasisPoints: uint64(fee.FeeBps), MaximumFee: fee.MaxFee}, nil
}

// TransferFeeOn is what Token-2022 withholds when amount is sent:
// min(ceil(amount * bps / 10_000), maximumFee).
func TransferFeeOn(fee *QuoteTransferFee, amount *big.Int) (*big.Int, error) {
	if fee == nil || fee.BasisPoints == 0 || amount == nil || amount.Sign() == 0 {
		return big.NewInt(0), nil
	}
	raw, err := ceilDiv(
		new(big.Int).Mul(new(big.Int).Set(amount), new(big.Int).SetUint64(fee.BasisPoints)),
		big.NewInt(10_000),
	)
	if err != nil {
		return nil, err
	}
	capFee := new(big.Int).SetUint64(fee.MaximumFee)
	if raw.Cmp(capFee) > 0 {
		return capFee, nil
	}
	return raw, nil
}

// NetOfTransferFee is what lands when amount is sent.
func NetOfTransferFee(fee *QuoteTransferFee, amount *big.Int) (*big.Int, error) {
	withheld, err := TransferFeeOn(fee, amount)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Sub(new(big.Int).Set(amount), withheld), nil
}

// GrossForTransferFee is the amount to send so at least net lands.
// It is at most one unit over the exact gross-up, which stays in the receiver.
func GrossForTransferFee(fee *QuoteTransferFee, net *big.Int) (*big.Int, error) {
	if fee == nil || fee.BasisPoints == 0 || net == nil || net.Sign() == 0 {
		return nonNil(net), nil
	}
	if fee.BasisPoints == 10_000 {
		return new(big.Int).Add(new(big.Int).Set(net), new(big.Int).SetUint64(fee.MaximumFee)), nil
	}
	raw, err := ceilDiv(
		new(big.Int).Mul(new(big.Int).Set(net), big.NewInt(10_000)),
		new(big.Int).Sub(big.NewInt(10_000), new(big.Int).SetUint64(fee.BasisPoints)),
	)
	if err != nil {
		return nil, err
	}
	if new(big.Int).Sub(raw, net).Cmp(new(big.Int).SetUint64(fee.MaximumFee)) >= 0 {
		return new(big.Int).Add(new(big.Int).Set(net), new(big.Int).SetUint64(fee.MaximumFee)), nil
	}
	return raw, nil
}
