package helpers

import solana "github.com/solana-foundation/solana-go/v2"

// NativeMint2022 is the Token-2022 native mint. create_v2 rejects it as a quote.
var NativeMint2022 = solana.MustPublicKeyFromBase58("9pan9bMn5HatX4EJdBwg9VgCa7Uz5HL8N1m5D3NdXejP")

// USDCMint is the mainnet USDC mint pump-fees treats as a stable quote.
// Devnet USDC is not in the on-chain list.
var USDCMint = solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")

// IsLegacyQuoteMint reports whether create_v2 treats the mint as SOL:
// WSOL or the zero key legacy curves store.
func IsLegacyQuoteMint(quoteMint solana.PublicKey) bool {
	return quoteMint.Equals(solana.PublicKey{}) || quoteMint.Equals(solana.WrappedSol)
}

// NormalizeQuoteMint maps the zero key and WSOL to WSOL. Every other mint is unchanged.
func NormalizeQuoteMint(quoteMint solana.PublicKey) solana.PublicKey {
	if IsLegacyQuoteMint(quoteMint) {
		return solana.WrappedSol
	}
	return quoteMint
}

// IsSolLikeQuoteMint mirrors pump-fees: the zero key, WSOL and the Token-2022 native mint.
func IsSolLikeQuoteMint(quoteMint solana.PublicKey) bool {
	return quoteMint.Equals(solana.PublicKey{}) ||
		quoteMint.Equals(solana.WrappedSol) ||
		quoteMint.Equals(NativeMint2022)
}

// IsStableQuoteMint reports whether pump-fees charges stableFeeTiers.
func IsStableQuoteMint(quoteMint solana.PublicKey) bool {
	return quoteMint.Equals(USDCMint)
}

// IsExoticQuoteMint reports a quote that is neither SOL-like nor a listed stable.
func IsExoticQuoteMint(quoteMint solana.PublicKey) bool {
	return !IsSolLikeQuoteMint(quoteMint) && !IsStableQuoteMint(quoteMint)
}
