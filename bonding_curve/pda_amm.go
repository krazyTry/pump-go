package bonding_curve

import (
	"encoding/binary"

	solana "github.com/gagliardetto/solana-go"
)

// Amm

const CanonicalPoolIndex uint16 = 0

func DeriveAmmEventAuthority() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("__event_authority")}, AmmProgramID)
	return key
}

func DeriveCanonicalPoolPDAWithQuote(poolAuthority solana.PublicKey, baseMint, quoteMint solana.PublicKey) solana.PublicKey {
	idx := make([]byte, 2)
	binary.LittleEndian.PutUint16(idx, CanonicalPoolIndex)
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("pool"), idx, poolAuthority.Bytes(), baseMint.Bytes(), quoteMint.Bytes()}, AmmProgramID)
	return key
}

func DeriveCoinCreatorVaultAuthority(sharing solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("coin_creator_vault_authority"), sharing.Bytes()}, AmmProgramID)
	return key
}

func DeriveAmmLpMint(pool solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("pool_lp_mint"), pool.Bytes()}, AmmProgramID)
	return key
}

func DeriveAmmGlobalConfig() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("global_config")}, AmmProgramID)
	return key
}
