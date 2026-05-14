package bonding_curve

import (
	solana "github.com/gagliardetto/solana-go"
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
	MetaplexProgramID      = solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
)

const (
	MaxShareholders = 10
)
