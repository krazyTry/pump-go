package bonding_curve

import (
	solana "github.com/solana-foundation/solana-go/v2"
)

// Mayhem
func DeriveMayhemGlobalParamsPDA() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("global-params")}, MayhemProgramID)
	return key
}
func DeriveMayhemStatePDA(mint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("mayhem-state"), mint.Bytes()}, MayhemProgramID)
	return key
}
func DeriveSolVaultPDA() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("sol-vault")}, MayhemProgramID)
	return key
}
