package amm

import (
	solana "github.com/gagliardetto/solana-go"
)

func DeriveFeesFeeConfig() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("fee_config"), ProgramID.Bytes()}, FeeProgramID)
	return key
}
