package amm

import (
	"encoding/binary"

	"github.com/krazyTry/pump-go/amm/helpers"
	solana "github.com/solana-foundation/solana-go/v2"
)

// PumpAmmPda is the pump-amm program address for seeds.
func PumpAmmPda(seeds [][]byte) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress(seeds, ProgramID)
	return key
}

func DerivePoolPDA(index uint16, owner, baseMint, quoteMint solana.PublicKey) solana.PublicKey {
	idx := make([]byte, 2)
	binary.LittleEndian.PutUint16(idx, index)
	return PumpAmmPda([][]byte{
		[]byte("pool"),
		idx,
		owner.Bytes(),
		baseMint.Bytes(),
		quoteMint.Bytes(),
	})
}

func DeriveCanonicalPumpPoolPDA(index uint16, baseMint, quoteMint solana.PublicKey) solana.PublicKey {
	return DerivePoolPDA(index, DerivePumpPoolAuthority(baseMint), baseMint, quoteMint)
}

// CanonicalPoolQuoteMint maps a bonding curve's stored quote mint onto the
// mint its graduated pool is keyed by. SOL curves store the zero key; the
// pool is quoted in WSOL.
func CanonicalPoolQuoteMint(quoteMint solana.PublicKey) solana.PublicKey {
	if quoteMint.Equals(solana.PublicKey{}) {
		return solana.WrappedSol
	}
	return quoteMint
}

// CanonicalPumpPoolPDA is the pool a coin graduates into. quoteMint may be
// the pool quote or the bonding curve's quote (the zero key for SOL).
func CanonicalPumpPoolPDA(baseMint, quoteMint solana.PublicKey) solana.PublicKey {
	return DeriveCanonicalPumpPoolPDA(CanonicalPoolIndex, baseMint, CanonicalPoolQuoteMint(quoteMint))
}

func DerivePoolV2PDA(baseMint solana.PublicKey) solana.PublicKey {
	return PumpAmmPda([][]byte{[]byte("pool-v2"), baseMint.Bytes()})
}

func DeriveUserVolumeAccumulator(user solana.PublicKey) solana.PublicKey {
	return PumpAmmPda([][]byte{[]byte("user_volume_accumulator"), user.Bytes()})
}

func DeriveCoinCreatorVault(coinCreator solana.PublicKey) solana.PublicKey {
	return PumpAmmPda([][]byte{[]byte("creator_vault"), coinCreator.Bytes()})
}

func DeriveGlobalConfig() solana.PublicKey {
	return PumpAmmPda([][]byte{[]byte("global_config")})
}

func DeriveEventAuthority() solana.PublicKey {
	return PumpAmmPda([][]byte{[]byte("__event_authority")})
}

func DeriveGlobalVolumeAccumulator() solana.PublicKey {
	return PumpAmmPda([][]byte{[]byte("global_volume_accumulator")})
}

func DeriveLpMint(pool solana.PublicKey) solana.PublicKey {
	return PumpAmmPda([][]byte{[]byte("pool_lp_mint"), pool.Bytes()})
}

func DeriveBoostVaultAuthority(pool solana.PublicKey) solana.PublicKey {
	return PumpAmmPda([][]byte{[]byte("boost_vault"), pool.Bytes()})
}

// LpMintAta is the owner's Token-2022 associated account for a pool LP mint.
func LpMintAta(lpMint, owner solana.PublicKey) solana.PublicKey {
	return helpers.FindAssociatedTokenAddress(owner, lpMint, solana.Token2022ProgramID)
}

// CoinCreatorVaultAta is the coin-creator vault's associated account for quoteMint.
func CoinCreatorVaultAta(coinCreatorVaultAuthority, quoteMint, quoteTokenProgram solana.PublicKey) solana.PublicKey {
	return helpers.FindAssociatedTokenAddress(coinCreatorVaultAuthority, quoteMint, quoteTokenProgram)
}

// BoostVaultAta is the boost vault authority's associated account for quoteMint.
func BoostVaultAta(boostVaultAuthority, quoteMint, quoteTokenProgram solana.PublicKey) solana.PublicKey {
	return helpers.FindAssociatedTokenAddress(boostVaultAuthority, quoteMint, quoteTokenProgram)
}
