package amm

import (
	solana "github.com/solana-foundation/solana-go/v2"
)

// PumpPda is the pump program address for seeds.
func PumpPda(seeds [][]byte) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress(seeds, PumpProgramID)
	return key
}

func DerivePumpPoolAuthority(mint solana.PublicKey) solana.PublicKey {
	return PumpPda([][]byte{[]byte("pool-authority"), mint.Bytes()})
}

func DerivePumpBondingCurve(mint solana.PublicKey) solana.PublicKey {
	return PumpPda([][]byte{[]byte("bonding-curve"), mint.Bytes()})
}

func DerivePumpCreatorVault(creator solana.PublicKey) solana.PublicKey {
	return PumpPda([][]byte{[]byte("creator-vault"), creator.Bytes()})
}

func DeriveHolderRewards(mint solana.PublicKey) solana.PublicKey {
	return PumpPda([][]byte{[]byte("holder-rewards"), mint.Bytes()})
}

func DerivePumpGlobal() solana.PublicKey {
	return PumpPda([][]byte{[]byte("global")})
}

func DerivePumpEventAuthority() solana.PublicKey {
	return PumpPda([][]byte{[]byte("__event_authority")})
}

func DeriveMetadata(mint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{
		[]byte("metadata"),
		MplTokenMetadataProgramID.Bytes(),
		mint.Bytes(),
	}, MplTokenMetadataProgramID)
	return key
}
