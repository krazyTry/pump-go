package amm

import (
	solana "github.com/solana-foundation/solana-go/v2"
)

// PumpFeePda is the pump-fees program address for seeds.
func PumpFeePda(seeds [][]byte) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress(seeds, FeeProgramID)
	return key
}

func DeriveFeesFeeConfig() solana.PublicKey {
	return PumpFeePda([][]byte{[]byte("fee_config"), ProgramID.Bytes()})
}

func DerivePumpFeeConfig() solana.PublicKey {
	return PumpFeePda([][]byte{[]byte("fee_config"), PumpProgramID.Bytes()})
}

func DeriveFeeSharingConfig(mint solana.PublicKey) solana.PublicKey {
	return PumpFeePda([][]byte{[]byte("sharing-config"), mint.Bytes()})
}
