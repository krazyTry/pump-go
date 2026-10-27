package amm

import (
	ammgen "github.com/krazyTry/pump-go/gen/amm"
	feesgen "github.com/krazyTry/pump-go/gen/fees"
	pumpgen "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

const (
	// PumpAmmTotalTokenSupply is the market-cap basis a mayhem pool uses
	// instead of the live base-mint supply.
	PumpAmmTotalTokenSupply uint64 = 1_000_000_000_000_000
	// CanonicalPoolIndex is the index of the pool a bonding curve graduates into.
	CanonicalPoolIndex uint16 = 0
	// MultiHopMaxHops is the longest route the programs' heap can hold.
	MultiHopMaxHops = 6
	// MultiHopMaxPoolsAtMaxHops is how many of those hops may be pools.
	MultiHopMaxPoolsAtMaxHops = 3

	// PoolSize is the serialized size of the current Pool layout, discriminator included.
	// Older pools are shorter; DecodePool zero-pads them to this size.
	PoolSize = 287
	// PoolAccountNewSize is the length extend_account grows a pool to.
	PoolAccountNewSize = 300
	// GlobalConfigSize is the serialized size of the current GlobalConfig layout.
	GlobalConfigSize = 949
	// FeeConfigSizePreStable is a FeeConfig that ends at feeTiers.
	FeeConfigSizePreStable = 2512
	// FeeConfigSizePostStable is a FeeConfig that also carries stableFeeTiers.
	FeeConfigSizePostStable = 4073
	// FeeConfigSizePostExotic is a FeeConfig that also carries exoticFlatFees.
	FeeConfigSizePostExotic = 4097
)

var (
	PumpProgramID = pumpgen.ProgramID
	ProgramID     = ammgen.ProgramID
	FeeProgramID  = feesgen.ProgramID
	PumpMint      = solana.MustPublicKeyFromBase58("pumpCmXqMfrsAkQ5r49WcJnRayYRqmXz6ae8H7H9Dfn")
	// USDCMint is the only quote mint that selects stableFeeTiers.
	USDCMint = solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	// NativeMint2022 is the Token-2022 native mint. It is SOL-like for fees.
	NativeMint2022 = solana.MustPublicKeyFromBase58("9pan9bMn5HatX4EJdBwg9VgCa7Uz5HL8N1m5D3NdXejP")
	// MplTokenMetadataProgramID owns the metadata account set_coin_creator reads.
	MplTokenMetadataProgramID = solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")

	// StableQuoteMints select stableFeeTiers on a canonical pump pool.
	StableQuoteMints = []solana.PublicKey{USDCMint}
	// SolLikeQuoteMints select feeTiers: the zero key SOL curves store, WSOL, and the Token-2022 native mint.
	SolLikeQuoteMints = []solana.PublicKey{{}, solana.WrappedSol, NativeMint2022}
)
