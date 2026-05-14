package amm

import (
	"encoding/binary"

	solana "github.com/gagliardetto/solana-go"
)

const CanonicalPoolIndex uint16 = 0

func DeriveCanonicalPumpPoolPDA(index uint16, creator, baseMint, quoteMint solana.PublicKey) solana.PublicKey {
	idx := make([]byte, 2)
	binary.LittleEndian.PutUint16(idx, index)

	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("pool"), idx, creator.Bytes(), baseMint.Bytes(), quoteMint.Bytes()}, ProgramID)
	return key
}

func DerivePoolV2PDA(baseMint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("pool-v2"), baseMint.Bytes()}, ProgramID)
	return key
}

func DeriveUserVolumeAccumulator(user solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("user_volume_accumulator"), user.Bytes()}, ProgramID)
	return key
}

func DeriveCoinCreatorVault(coinCreator solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("creator_vault"), coinCreator.Bytes()}, ProgramID)
	return key
}

func DeriveGlobalConfig() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("global_config")}, ProgramID)
	return key
}

func DeriveEventAuthority() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("__event_authority")}, ProgramID)
	return key
}

func DeriveGlobalVolumeAccumulator() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("global_volume_accumulator")}, ProgramID)
	return key
}

func DeriveLpMint(pool solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("pool_lp_mint"), pool.Bytes()}, ProgramID)
	return key
}
