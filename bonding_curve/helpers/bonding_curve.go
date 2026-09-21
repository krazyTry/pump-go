package helpers

import (
	"errors"

	solana "github.com/solana-foundation/solana-go/v2"
	pump "github.com/krazyTry/pump-go/gen/pump"
	"github.com/shopspring/decimal"
)

func NewBondingCurve(global *pump.Global, quoteMint solana.PublicKey) *pump.BondingCurve {
	q := global.InitialVirtualSolReserves
	if !quoteMint.Equals(solana.WrappedSol) {
		q = global.InitialVirtualQuoteReserves
	}
	return &pump.BondingCurve{
		VirtualTokenReserves: global.InitialVirtualTokenReserves,
		VirtualQuoteReserves: q,
		RealTokenReserves:    global.InitialRealTokenReserves,
		RealQuoteReserves:    0,
		TokenTotalSupply:     global.TokenTotalSupply,
		Complete:             false,
		Creator:              solana.PublicKey{},
		IsMayhemMode:         global.MayhemModeEnabled,
		IsCashbackCoin:       false,
		QuoteMint:            solana.PublicKey{},
	}
}

func BondingCurveMarketCap(mintSupply, virtualQuoteReserves, virtualTokenReserves uint64) (uint64, error) {
	if virtualTokenReserves == 0 {
		return 0, errors.New("division by zero: virtual token reserves")
	}
	mintSupplyDecimal := decimal.NewFromUint64(mintSupply)
	virtualQuoteReservesDecimal := decimal.NewFromUint64(virtualQuoteReserves)
	virtualTokenReservesDecimal := decimal.NewFromUint64(virtualTokenReserves)
	return virtualQuoteReservesDecimal.Mul(mintSupplyDecimal).Div(virtualTokenReservesDecimal).Ceil().BigInt().Uint64(), nil
}
