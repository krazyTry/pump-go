package bonding_curve

import (
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

// Token 2022
func (s *Client) CreateV2Instruction(creator, user, baseMint, quoteMint, quoteTokenProgram solana.PublicKey, name, symbol, uri string, mayhemMode bool, cashback bool, creatorFeeBps uint64, isHolderReward bool) (solana.Instruction, error) {

	mintAuthority := DeriveMintAuthority()

	solVault := DeriveSolVaultPDA()
	bondingCurve := DeriveBondingCurve(baseMint)

	bondingCurveAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, solana.Token2022ProgramID)

	globalParams := DeriveMayhemGlobalParamsPDA()
	mayhemState := DeriveMayhemStatePDA(baseMint)

	mayhemTokenVault := helpers.FindAssociatedTokenAddress(solVault, baseMint, solana.Token2022ProgramID)

	// creatorFeeBpsParam OptionU64,
	// isHolderRewardParam OptionBool,

	ix, err := pump.NewCreateV2Instruction(
		name,
		symbol,
		uri,
		creator,
		mayhemMode,
		pump.OptionBool{V0: cashback},
		pump.OptionU64{V0: creatorFeeBps},
		pump.OptionBool{V0: isHolderReward},

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

	if !helpers.IsLegacyQuoteMint(quoteMint) {
		appendCreateV2QuoteAccounts(ix, baseMint, quoteMint, quoteTokenProgram, nil)
	}
	return ix, nil
}

// CreateV2InstructionWithPumpQuote is create_v2 for a pump-coin quote.
// pumpQuote supplies the canonical pool accounts of Q and of every coin below it.
func (s *Client) CreateV2InstructionWithPumpQuote(creator, user, baseMint, quoteMint, quoteTokenProgram solana.PublicKey, name, symbol, uri string, mayhemMode bool, cashback bool, creatorFeeBps uint64, isHolderReward bool, pumpQuote *PumpQuoteAccounts) (solana.Instruction, error) {
	ix, err := s.CreateV2Instruction(creator, user, baseMint, solana.PublicKey{}, quoteTokenProgram, name, symbol, uri, mayhemMode, cashback, creatorFeeBps, isHolderReward)
	if err != nil {
		return nil, err
	}
	if helpers.IsLegacyQuoteMint(quoteMint) {
		return ix, nil
	}
	appendCreateV2QuoteAccounts(ix, baseMint, quoteMint, quoteTokenProgram, pumpQuote)
	return ix, nil
}

func appendCreateV2QuoteAccounts(ix solana.Instruction, baseMint, quoteMint, quoteTokenProgram solana.PublicKey, pumpQuote *PumpQuoteAccounts) {
	gi, ok := ix.(*solana.GenericInstruction)
	if !ok {
		return
	}
	bondingCurve := DeriveBondingCurve(baseMint)
	gi.AccountValues = append(gi.AccountValues,
		solana.NewAccountMeta(quoteMint, false, false),
		solana.NewAccountMeta(helpers.FindAssociatedTokenAddress(bondingCurve, quoteMint, quoteTokenProgram), true, false),
		solana.NewAccountMeta(quoteTokenProgram, false, false),
		solana.NewAccountMeta(DeriveQuoteControl(), false, false),
	)
	if pumpQuote == nil {
		return
	}
	appendPumpQuoteGroup(gi, quoteMint, pumpQuote.Pool)
	for _, coin := range pumpQuote.Chain {
		appendPumpQuoteGroup(gi, coin.Mint, coin.Pool)
	}
}

func appendPumpQuoteGroup(gi *solana.GenericInstruction, mint solana.PublicKey, pool *PumpQuotePoolAccounts) {
	gi.AccountValues = append(gi.AccountValues, solana.NewAccountMeta(DeriveBondingCurve(mint), false, false))
	if pool == nil {
		return
	}
	gi.AccountValues = append(gi.AccountValues,
		solana.NewAccountMeta(pool.Pool, false, false),
		solana.NewAccountMeta(pool.BaseVault, false, false),
		solana.NewAccountMeta(pool.QuoteVault, false, false),
	)
}

func (c *Client) GetBuyV2Instruction(creator solana.PublicKey, partialFill bool, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount solana.PublicKey, amount, maxSolCost uint64, feeRecipient, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
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
		pump.OptionBool{V0: partialFill},
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

func (c *Client) GetBuyExactQuoteInV2Instruction(creator solana.PublicKey, partialFill bool, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount solana.PublicKey, amount, tokenCost uint64, feeRecipient, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
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
		pump.OptionBool{V0: partialFill},
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
