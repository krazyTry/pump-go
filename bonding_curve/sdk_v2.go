package bonding_curve

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	pump "github.com/krazyTry/pump-go/gen/pump"
)

// Token 2022
func (s *Client) CreateV2Instruction(creator, user, baseMint, quoteMint, quoteTokenProgram solana.PublicKey, name, symbol, uri string, mayhemMode bool, cashback bool) (solana.Instruction, error) {

	mintAuthority := DeriveMintAuthority()

	solVault := DeriveSolVaultPDA()
	bondingCurve := DeriveBondingCurve(baseMint)

	bondingCurveAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, solana.Token2022ProgramID)

	globalParams := DeriveMayhemGlobalParamsPDA()
	mayhemState := DeriveMayhemStatePDA(baseMint)

	mayhemTokenVault := helpers.FindAssociatedTokenAddress(solVault, baseMint, solana.Token2022ProgramID)

	ix, err := pump.NewCreateV2Instruction(
		name,
		symbol,
		uri,
		creator,
		mayhemMode,
		pump.OptionBool{V0: cashback},
		baseMint,
		mintAuthority,
		bondingCurve,
		bondingCurveAta,
		s.Global,
		user,
		solana.SystemProgramID,
		solana.Token2022ProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		MayhemProgramID,
		globalParams,
		solVault,
		mayhemState,
		mayhemTokenVault,
		s.EventAuthority,
		ProgramID,
	)

	if err != nil {
		return nil, err
	}

	if !(quoteMint.Equals(solana.WrappedSol) || quoteMint.Equals(solana.PublicKey{})) {
		if gi, ok := ix.(*solana.GenericInstruction); ok {
			gi.AccountValues = append(
				gi.AccountValues,
				solana.NewAccountMeta(quoteMint, false, false),
				solana.NewAccountMeta(helpers.FindAssociatedTokenAddress(bondingCurve, quoteMint, quoteTokenProgram), false, false),
				solana.NewAccountMeta(quoteTokenProgram, false, false),
			)
		}
	}
	return ix, nil
}

func (c *Client) GetBuyV2Instruction(creator, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount solana.PublicKey, amount, maxSolCost uint64, feeRecipient, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
	bondingCurve := DeriveBondingCurve(baseMint)
	bondingCurveBaseAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, baseTokenProgram)
	bondingCurveQuoteAta := helpers.FindAssociatedTokenAddress(bondingCurve, quoteMint, quoteTokenProgram)
	creatorVault := DeriveCreatorVault(creator)
	creatorVaultAta := helpers.FindAssociatedTokenAddress(creatorVault, quoteMint, quoteTokenProgram)
	userVolumeAccumulator := DeriveUserVolumeAccumulator(user)
	userVolumeAccumulatorAta := helpers.FindAssociatedTokenAddress(userVolumeAccumulator, quoteMint, quoteTokenProgram)
	sharingConfig := DeriveFeesFeeSharingConfig(baseMint)

	feeRecipientAta := helpers.FindAssociatedTokenAddress(feeRecipient, quoteMint, quoteTokenProgram)
	buybackFeeRecipientAta := helpers.FindAssociatedTokenAddress(buybackFeeRecipient, quoteMint, quoteTokenProgram)

	return pump.NewBuyV2Instruction(
		amount,
		maxSolCost,
		c.Global,
		baseMint,
		quoteMint,
		baseTokenProgram,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
		feeRecipient,
		feeRecipientAta,
		buybackFeeRecipient,
		buybackFeeRecipientAta,
		bondingCurve,
		bondingCurveBaseAta,
		bondingCurveQuoteAta,
		user,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		creatorVault,
		creatorVaultAta,
		sharingConfig,
		c.GlobalVolumeAccumulator,
		userVolumeAccumulator,
		userVolumeAccumulatorAta,
		c.FeeConfig,
		FeeProgramID,
		solana.SystemProgramID,
		c.EventAuthority,
		ProgramID,
	)
}

func (c *Client) GetBuyExactQuoteInV2Instruction(creator, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount solana.PublicKey, amount, tokenCost uint64, feeRecipient, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
	bondingCurve := DeriveBondingCurve(baseMint)
	bondingCurveBaseAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, baseTokenProgram)
	bondingCurveQuoteAta := helpers.FindAssociatedTokenAddress(bondingCurve, quoteMint, quoteTokenProgram)
	creatorVault := DeriveCreatorVault(creator)
	creatorVaultAta := helpers.FindAssociatedTokenAddress(creatorVault, quoteMint, quoteTokenProgram)
	userVolumeAccumulator := DeriveUserVolumeAccumulator(user)
	userVolumeAccumulatorAta := helpers.FindAssociatedTokenAddress(userVolumeAccumulator, quoteMint, quoteTokenProgram)
	sharingConfig := DeriveFeesFeeSharingConfig(baseMint)

	feeRecipientAta := helpers.FindAssociatedTokenAddress(feeRecipient, quoteMint, quoteTokenProgram)
	buybackFeeRecipientAta := helpers.FindAssociatedTokenAddress(buybackFeeRecipient, quoteMint, quoteTokenProgram)

	return pump.NewBuyExactQuoteInV2Instruction(
		amount,
		tokenCost,
		c.Global,
		baseMint,
		quoteMint,
		baseTokenProgram,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
		feeRecipient,
		feeRecipientAta,
		buybackFeeRecipient,
		buybackFeeRecipientAta,
		bondingCurve,
		bondingCurveBaseAta,
		bondingCurveQuoteAta,
		user,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		creatorVault,
		creatorVaultAta,
		sharingConfig,
		c.GlobalVolumeAccumulator,
		userVolumeAccumulator,
		userVolumeAccumulatorAta,
		c.FeeConfig,
		FeeProgramID,
		solana.SystemProgramID,
		c.EventAuthority,
		ProgramID,
	)
}

func (c *Client) GetSellV2Instruction(creator, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount solana.PublicKey, amount, minSolOutput uint64, feeRecipient, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
	bondingCurve := DeriveBondingCurve(baseMint)
	bondingCurveBaseAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, baseTokenProgram)
	bondingCurveQuoteAta := helpers.FindAssociatedTokenAddress(bondingCurve, quoteMint, quoteTokenProgram)
	creatorVault := DeriveCreatorVault(creator)
	creatorVaultAta := helpers.FindAssociatedTokenAddress(creatorVault, quoteMint, quoteTokenProgram)
	userVolumeAccumulator := DeriveUserVolumeAccumulator(user)
	userVolumeAccumulatorAta := helpers.FindAssociatedTokenAddress(userVolumeAccumulator, quoteMint, quoteTokenProgram)
	sharingConfig := DeriveFeesFeeSharingConfig(baseMint)

	feeRecipientAta := helpers.FindAssociatedTokenAddress(feeRecipient, quoteMint, quoteTokenProgram)
	buybackFeeRecipientAta := helpers.FindAssociatedTokenAddress(buybackFeeRecipient, quoteMint, quoteTokenProgram)

	return pump.NewSellV2Instruction(
		amount,
		minSolOutput,
		c.Global,
		baseMint,
		quoteMint,
		baseTokenProgram,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
		feeRecipient,
		feeRecipientAta,
		buybackFeeRecipient,
		buybackFeeRecipientAta,
		bondingCurve,
		bondingCurveBaseAta,
		bondingCurveQuoteAta,
		user,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		creatorVault,
		creatorVaultAta,
		sharingConfig,
		userVolumeAccumulator,
		userVolumeAccumulatorAta,
		c.FeeConfig,
		FeeProgramID,
		solana.SystemProgramID,
		c.EventAuthority,
		ProgramID,
	)
}

func (c *Client) MigrateV2Instruction(withdrawAuthority, user, baseMint, baseTokenProgram, quoteMint, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	if quoteMint.Equals(solana.PublicKey{}) {
		quoteMint = solana.WrappedSol
	}

	if baseTokenProgram.Equals(solana.PublicKey{}) {
		baseTokenProgram = solana.Token2022ProgramID
	}

	if quoteTokenProgram.Equals(solana.PublicKey{}) {
		quoteTokenProgram = solana.TokenProgramID
	}

	poolAuthority := DerivePoolAuthority(baseMint)

	poolAuthorityBaseAta := helpers.FindAssociatedTokenAddress(poolAuthority, baseMint, baseTokenProgram)
	poolAuthorityQuoteAta := helpers.FindAssociatedTokenAddress(poolAuthority, quoteMint, quoteTokenProgram)

	pool := DeriveCanonicalPoolPDAWithQuote(poolAuthority, baseMint, quoteMint)

	bondingCurve := DeriveBondingCurve(baseMint)
	bondingCurveBaseAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, baseTokenProgram)
	bondingCurveQuoteAta := helpers.FindAssociatedTokenAddress(bondingCurve, quoteMint, quoteTokenProgram)

	lpMint := DeriveAmmLpMint(pool)

	userPoolTokenAta := helpers.FindAssociatedTokenAddress(user, lpMint, solana.Token2022ProgramID)
	poolBaseTokenAta := helpers.FindAssociatedTokenAddress(pool, baseMint, baseTokenProgram)
	poolQuoteTokenAta := helpers.FindAssociatedTokenAddress(pool, quoteMint, quoteTokenProgram)

	return pump.NewMigrateV2Instruction(
		c.Global,
		withdrawAuthority,
		baseMint,
		quoteMint,
		bondingCurve,
		bondingCurveBaseAta,
		bondingCurveQuoteAta,
		user,
		solana.SystemProgramID,
		AmmProgramID,
		pool,
		poolAuthority,
		poolAuthorityBaseAta,
		poolAuthorityQuoteAta,
		c.AmmGlobalConfig,
		lpMint,
		userPoolTokenAta,
		poolBaseTokenAta,
		poolQuoteTokenAta,
		baseTokenProgram,
		quoteTokenProgram,
		solana.Token2022ProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		c.AmmEventAuthority,
		solana.SysVarRentPubkey,
		c.EventAuthority,
		ProgramID,
	)
}

func (s *Client) ClaimCashbackV2Instruction(user, quoteMint, quoteTokenProgram, userQuoteTokenAccount solana.PublicKey) (solana.Instruction, error) {
	userVolumeAccumulator := DeriveUserVolumeAccumulator(user)
	userVolumeAccumulatorAta := helpers.FindAssociatedTokenAddress(userVolumeAccumulator, quoteMint, quoteTokenProgram)

	return pump.NewClaimCashbackV2Instruction(
		user,
		userQuoteTokenAccount,
		userVolumeAccumulator,
		userVolumeAccumulatorAta,
		quoteMint,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
}
