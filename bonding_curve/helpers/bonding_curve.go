package helpers

import (
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

// NewBondingCurve seeds a curve the way create_v2 does when the caller has not
// fetched QuoteControl and the quote is not a nested pump coin.
// SOL (the zero key or WSOL) starts at InitialVirtualSolReserves. Every other
// mint starts at InitialVirtualQuoteReserves, which is correct for a mint
// whitelisted on Global and for callers that predate quote control.
func NewBondingCurve(global *pump.Global, quoteMint solana.PublicKey) *pump.BondingCurve {
	virtualQuote := global.InitialVirtualSolReserves
	stored := solana.PublicKey{}
	if !IsLegacyQuoteMint(quoteMint) {
		virtualQuote = global.InitialVirtualQuoteReserves
		stored = quoteMint
	}
	return &pump.BondingCurve{
		VirtualTokenReserves:        global.InitialVirtualTokenReserves,
		VirtualQuoteReserves:        virtualQuote,
		RealTokenReserves:           global.InitialRealTokenReserves,
		RealQuoteReserves:           0,
		TokenTotalSupply:            global.TokenTotalSupply,
		Complete:                    false,
		Creator:                     solana.PublicKey{},
		IsMayhemMode:                global.MayhemModeEnabled,
		IsCashbackCoin:              false,
		QuoteMint:                   stored,
		CreatorFeeBps:               0,
		CanEditCreatorFee:           false,
		IsHolderReward:              false,
		Depth:                       0,
		InitialVirtualQuoteReserves: virtualQuote,
	}
}
