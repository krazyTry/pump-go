package helpers

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
)

// TokenInfo mirrors needed fields for Token2022 fee calculations.
type TokenInfo struct {
	Owner solana.PublicKey
	Mint  *token.Mint
	Ext   *Extensions
	Fee   TransferFee
}
