package helpers

import (
	solana "github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/token"
)

// TokenInfo mirrors needed fields for Token2022 fee calculations.
type TokenInfo struct {
	Owner solana.PublicKey
	Mint  *token.Mint
	Ext   *Extensions
	Fee   TransferFee
}
