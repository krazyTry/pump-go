package helpers

import (
	"errors"
	"math/rand"
	"time"

	solana "github.com/gagliardetto/solana-go"
	pump "github.com/krazyTry/pump-go/gen/pump"
	"github.com/shopspring/decimal"
)

var (
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
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

var feeRecipients = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("62qc2CNXwrYqQScmEdiZFFAnJR262PxWEuNQtxfafNgV"),
	solana.MustPublicKeyFromBase58("7VtfL8fvgNfhz17qKRMjzQEXgbdpnHHHQRh54R9jP2RJ"),
	solana.MustPublicKeyFromBase58("7hTckgnGnLQR6sdH7YkqFTAA7VwTfYFaZ6EhEsU3saCX"),
	solana.MustPublicKeyFromBase58("9rPYyANsfQZw3DnDmKE3YCQF5E8oD89UXoHn9JFEhJUz"),
	solana.MustPublicKeyFromBase58("AVmoTthdrX6tKt4nDjco2D775W2YK3sDhxPcMmzUAmTY"),
	solana.MustPublicKeyFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM"),
	solana.MustPublicKeyFromBase58("FWsW1xNtWscwNmKv6wVsU1iTzRN6wmmk3MjxRP5tT7hz"),
	solana.MustPublicKeyFromBase58("G5UZAVbAf46s7cKWoyKu8kYTip9DGTpbLZ2qa9Aq69dP"),
}

func GetStaticRandomFeeRecipient() solana.PublicKey {
	return feeRecipients[rnd.Intn(len(feeRecipients))]
}

var buybackRecipients = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("5YxQFdt3Tr9zJLvkFccqXVUwhdTWJQc1fFg2YPbxvxeD"),
	solana.MustPublicKeyFromBase58("9M4giFFMxmFGXtc3feFzRai56WbBqehoSeRE5GK7gf7"),
	solana.MustPublicKeyFromBase58("GXPFM2caqTtQYC2cJ5yJRi9VDkpsYZXzYdwYpGnLmtDL"),
	solana.MustPublicKeyFromBase58("3BpXnfJaUTiwXnJNe7Ej1rcbzqTTQUvLShZaWazebsVR"),
	solana.MustPublicKeyFromBase58("5cjcW9wExnJJiqgLjq7DEG75Pm6JBgE1hNv4B2vHXUW6"),
	solana.MustPublicKeyFromBase58("EHAAiTxcdDwQ3U4bU6YcMsQGaekdzLS3B5SmYo46kJtL"),
	solana.MustPublicKeyFromBase58("5eHhjP8JaYkz83CWwvGU2uMUXefd3AazWGx4gpcuEEYD"),
	solana.MustPublicKeyFromBase58("A7hAgCzFw14fejgCp387JUJRMNyz4j89JKnhtKU8piqW"),
}

func GetStaticRandomFeeRecipientForBuyback() solana.PublicKey {
	return buybackRecipients[rnd.Intn(len(buybackRecipients))]
}
