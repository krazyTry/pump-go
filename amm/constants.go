package amm

import (
	solana "github.com/gagliardetto/solana-go"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
	feesgen "github.com/krazyTry/pump-go/gen/fees"
	pumpgen "github.com/krazyTry/pump-go/gen/pump"
)

var (
	PumpProgramID = pumpgen.ProgramID
	ProgramID     = ammgen.ProgramID
	FeeProgramID  = feesgen.ProgramID
	PumpMint      = solana.MustPublicKeyFromBase58("pumpCmXqMfrsAkQ5r49WcJnRayYRqmXz6ae8H7H9Dfn")
)
