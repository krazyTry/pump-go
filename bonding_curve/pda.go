package bonding_curve

import (
	solana "github.com/gagliardetto/solana-go"
)

func DeriveMintMetadata(mint solana.PublicKey) solana.PublicKey {
	pub, _, _ := solana.FindProgramAddress([][]byte{[]byte("metadata"), MetaplexProgramID.Bytes(), mint.Bytes()}, MetaplexProgramID)
	return pub
}

func DeriveMintAuthority() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("mint-authority")}, ProgramID)
	return key
}

func DerivePoolAuthority(mint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("pool-authority"), mint.Bytes()}, ProgramID)
	return key
}

func DeriveEventAuthority() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("__event_authority")}, ProgramID)
	return key
}

func DeriveGlobal() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("global")}, ProgramID)
	return key
}

func DeriveGlobalVolumeAccumulator() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("global_volume_accumulator")}, ProgramID)
	return key
}

func DeriveBondingCurve(mint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("bonding-curve"), mint.Bytes()}, ProgramID)
	return key
}

func DeriveBondingCurveV2(mint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("bonding-curve-v2"), mint.Bytes()}, ProgramID)
	return key
}

func DeriveCreatorVault(creator solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("creator-vault"), creator.Bytes()}, ProgramID)
	return key
}

func DeriveUserVolumeAccumulator(user solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("user_volume_accumulator"), user.Bytes()}, ProgramID)
	return key
}
