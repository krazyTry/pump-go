package amm

import (
	solana "github.com/gagliardetto/solana-go"
)

func DerivePumpPoolAuthority(mint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("pool-authority"), mint.Bytes()}, PumpProgramID)
	return key
}

func DerivePumpBondingCurve(mint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("bonding-curve"), mint.Bytes()}, PumpProgramID)
	return key
}

func DerivePumpCreatorVault(creator solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("creator-vault"), creator.Bytes()}, PumpProgramID)
	return key
}
