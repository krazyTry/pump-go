package bonding_curve

import (
	solana "github.com/solana-foundation/solana-go/v2"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
	feesgen "github.com/krazyTry/pump-go/gen/fees"
	pumpgen "github.com/krazyTry/pump-go/gen/pump"
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

	VirtualTokenReserves uint64 = 1_073_000_000_000_000
	VirtualSolReserves   uint64 = 30_000_000_000
	VirtualQuoteReserves uint64 = 4_292_000_000

	RealTokenReserves uint64 = 793_100_000_000_000
)
