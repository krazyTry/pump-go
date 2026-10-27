package bonding_curve

import (
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

func (s *Client) sweepBondingCurveFee(payer, mint, recipient, quoteMint, quoteTokenProgram solana.PublicKey, creator bool) (solana.Instruction, error) {
	quote := NormalizeQuoteMint(quoteMint)
	curve := DeriveBondingCurve(mint)
	curveATA := helpers.FindAssociatedTokenAddress(curve, quote, quoteTokenProgram)
	recipientATA := helpers.FindAssociatedTokenAddress(recipient, quote, quoteTokenProgram)
	if creator {
		return pump.NewSweepCreatorFeeInstruction(
			payer, s.Global, mint, quote, quoteTokenProgram,
			solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID,
			curve, curveATA, recipient, recipientATA, s.EventAuthority, ProgramID,
		)
	}
	return pump.NewSweepProtocolFeeInstruction(
		payer, s.Global, mint, quote, quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID,
		curve, curveATA, recipient, recipientATA, s.EventAuthority, ProgramID,
	)
}

// SweepProtocolFeeInstruction pays BondingCurve.protocolFees to recipient.
func (s *Client) SweepProtocolFeeInstruction(payer, mint, recipient, quoteMint, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	return s.sweepBondingCurveFee(payer, mint, recipient, quoteMint, quoteTokenProgram, false)
}

// SweepCreatorFeeInstruction pays BondingCurve.creatorFee to creatorVaultPda(creator).
func (s *Client) SweepCreatorFeeInstruction(payer, mint, creator, quoteMint, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	return s.sweepBondingCurveFee(payer, mint, DeriveCreatorVault(creator), quoteMint, quoteTokenProgram, true)
}

// SweepPoolCreatorFeeInstruction pays Pool.creatorFees into the coin-creator vault of coinCreator.
func (s *Client) SweepPoolCreatorFeeInstruction(payer, mint, coinCreator, quoteMint, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	quote := NormalizeQuoteMint(quoteMint)
	pool := DeriveCanonicalPoolPDAWithQuote(DerivePoolAuthority(mint), mint, quote)
	recipient := DeriveCoinCreatorVaultAuthority(coinCreator)
	return ammgen.NewSweepCreatorFeeInstruction(
		payer,
		s.AmmGlobalConfig,
		pool,
		quote,
		quoteTokenProgram,
		helpers.FindAssociatedTokenAddress(pool, quote, quoteTokenProgram),
		recipient,
		helpers.FindAssociatedTokenAddress(recipient, quote, quoteTokenProgram),
		solana.SystemProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		s.AmmEventAuthority,
		AmmProgramID,
	)
}
