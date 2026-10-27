package bonding_curve

import (
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
	feesgen "github.com/krazyTry/pump-go/gen/fees"
	pumpgen "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

var (
	ProgramID              = pumpgen.ProgramID
	AmmProgramID           = ammgen.ProgramID
	FeeProgramID           = feesgen.ProgramID
	MayhemProgramID        = solana.MustPublicKeyFromBase58("MAyhSmzXzV1pTf7LsNkrNwkWKTo4ougAJ1PPg47MD4e")
	PumpTokenMint          = solana.MustPublicKeyFromBase58("pumpCmXqMfrsAkQ5r49WcJnRayYRqmXz6ae8H7H9Dfn")
	DonationRelayProgramID = solana.MustPublicKeyFromBase58("RLAYHr9TRFcKB2ubYQhspcnXiaGpaVzNQvHytt47RZu")
)

const (
	MaxShareholders = 10

	// MaxCurveDepth is pump_quote::MAX_CURVE_DEPTH. create_v2 never admits a
	// deeper coin, whatever Global.MaxCurveDepth is set to.
	MaxCurveDepth = 5

	VirtualTokenReserves uint64 = 1_073_000_000_000_000
	VirtualSolReserves   uint64 = 30_000_000_000
	VirtualQuoteReserves uint64 = 4_292_000_000

	RealTokenReserves uint64 = 793_100_000_000_000
)

// NativeMint2022 is the Token-2022 native mint. create_v2 rejects it as a quote.
var NativeMint2022 = helpers.NativeMint2022

// USDCMint is the mainnet USDC mint charged on the stable fee schedule.
var USDCMint = helpers.USDCMint
