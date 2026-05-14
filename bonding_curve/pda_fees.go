package bonding_curve

import (
	solana "github.com/gagliardetto/solana-go"
)

// Fees

func DeriveFeesEventAuthority() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("__event_authority")}, FeeProgramID)
	return key
}

func DeriveFeesDonationFeePDA(mint, configID solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("donation-fee-pda"), mint.Bytes(), configID.Bytes()}, FeeProgramID)
	return key
}

func DeriveFeesSocialFeePDA(userID string, platform uint8) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("social-fee-pda"), []byte(userID), {platform}}, FeeProgramID)
	return key
}

func DeriveFeesFeeConfig() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("fee_config"), ProgramID.Bytes()}, FeeProgramID)
	return key
}

func DeriveFeesFeeProgramGlobal() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("fee-program-global")}, FeeProgramID)
	return key
}

func DeriveFeesFeeSharingConfig(mint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("sharing-config"), mint.Bytes()}, FeeProgramID)
	return key
}
