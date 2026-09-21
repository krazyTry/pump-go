package amm

import (
	solana "github.com/solana-foundation/solana-go/v2"
)

func DeriveFeesFeeConfig() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("fee_config"), ProgramID.Bytes()}, FeeProgramID)
	return key
}
