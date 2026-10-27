package bonding_curve

import (
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	fees "github.com/krazyTry/pump-go/gen/fees"
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

// SetMaxCurveDepthInstruction sets Global.maxCurveDepth. authority signs and must be Global.authority.
func (s *Client) SetMaxCurveDepthInstruction(authority solana.PublicKey, maxCurveDepth uint8) (solana.Instruction, error) {
	return pump.NewSetMaxCurveDepthInstruction(maxCurveDepth, s.Global, authority, s.EventAuthority, ProgramID)
}

// InitializeQuoteControlInstruction creates the QuoteControl PDA. user pays the rent.
func (s *Client) InitializeQuoteControlInstruction(user solana.PublicKey) (solana.Instruction, error) {
	return pump.NewInitializeQuoteControlInstruction(DeriveQuoteControl(), user, solana.SystemProgramID)
}

// SetQuoteControlAdminInstruction assigns QuoteControl.admin. The zero key revokes it.
func (s *Client) SetQuoteControlAdminInstruction(authority, newAdmin solana.PublicKey) (solana.Instruction, error) {
	return pump.NewSetQuoteControlAdminInstruction(newAdmin, s.Global, authority, DeriveQuoteControl(), s.EventAuthority, ProgramID)
}

// AddQuoteControlMintInstruction lists quoteMint with the virtual reserves a new curve starts from.
func (s *Client) AddQuoteControlMintInstruction(authority, quoteMint solana.PublicKey, initialVirtualQuoteReserves uint64) (solana.Instruction, error) {
	return pump.NewAddQuoteControlMintInstruction(
		quoteMint,
		initialVirtualQuoteReserves,
		authority,
		s.Global,
		DeriveQuoteControl(),
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

// RemoveQuoteControlMintInstruction de-lists quoteMint. Existing curves keep trading.
func (s *Client) RemoveQuoteControlMintInstruction(authority, quoteMint solana.PublicKey) (solana.Instruction, error) {
	return pump.NewRemoveQuoteControlMintInstruction(
		quoteMint,
		authority,
		s.Global,
		DeriveQuoteControl(),
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

// UpdateCreatorFeeConfigInstruction sets the per-coin creator-fee gate and its ceiling.
func (s *Client) UpdateCreatorFeeConfigInstruction(authority solana.PublicKey, creatorFeeConfigurable bool, maxConfigurableCreatorFeeBps uint64) (solana.Instruction, error) {
	return pump.NewUpdateCreatorFeeConfigInstruction(
		creatorFeeConfigurable,
		maxConfigurableCreatorFeeBps,
		s.Global,
		authority,
		s.EventAuthority,
		ProgramID,
	)
}

// UpdateHolderRewardConfigInstruction turns holder-reward coins on or off and names the claim signer.
func (s *Client) UpdateHolderRewardConfigInstruction(authority solana.PublicKey, isHolderRewardEnabled bool, holderRewardClaimAuthority solana.PublicKey) (solana.Instruction, error) {
	return pump.NewUpdateHolderRewardConfigInstruction(
		isHolderRewardEnabled,
		holderRewardClaimAuthority,
		s.Global,
		authority,
		s.EventAuthority,
		ProgramID,
	)
}

// HolderPayout is one distribute_fee_to_holders recipient.
type HolderPayout struct {
	Owner  solana.PublicKey
	Amount uint64
}

// DistributeFeeToHoldersInstruction pays holder-reward fees out of holderRewardsPda(mint).
// holderRewardsTokenAccount may be the zero key on a SOL quote that has no parked WSOL account.
func (s *Client) DistributeFeeToHoldersInstruction(holderRewardClaimAuthority, mint, quoteMint, quoteTokenProgram, holderRewardsTokenAccount solana.PublicKey, recipients []HolderPayout) (solana.Instruction, error) {
	quote := NormalizeQuoteMint(quoteMint)
	native := quote.Equals(solana.WrappedSol)
	amounts := make([]uint64, len(recipients))
	for i, r := range recipients {
		amounts[i] = r.Amount
	}
	tokenAccount := holderRewardsTokenAccount
	omitted := tokenAccount.Equals(solana.PublicKey{})
	if omitted {
		tokenAccount = ProgramID
	}
	ix, err := pump.NewDistributeFeeToHoldersInstruction(
		amounts,
		s.Global,
		holderRewardClaimAuthority,
		mint,
		DeriveHolderRewards(mint),
		tokenAccount,
		quote,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
	if err != nil {
		return nil, err
	}
	gi := ix.(*solana.GenericInstruction)
	if omitted {
		gi.AccountValues[4].PublicKey = ProgramID
		gi.AccountValues[4].IsWritable = false
	}
	for _, r := range recipients {
		gi.AccountValues = append(gi.AccountValues,
			solana.NewAccountMeta(r.Owner, native, false),
			solana.NewAccountMeta(quoteATA(r.Owner, quote, quoteTokenProgram), !native, false),
		)
	}
	return ix, nil
}

// SetExoticFlatFeesInstruction sets the flat schedule exotic quotes pay. admin signs and must be FeeConfig.admin.
func (s *Client) SetExoticFlatFeesInstruction(admin solana.PublicKey, exotic Fees) (solana.Instruction, error) {
	return fees.NewSetExoticFlatFeesInstruction(
		fees.Fees{LpFeeBps: exotic.LpFeeBps, ProtocolFeeBps: exotic.ProtocolFeeBps, CreatorFeeBps: exotic.CreatorFeeBps},
		s.FeeConfig,
		admin,
		ProgramID,
		solana.SystemProgramID,
		s.FeesEventAuthority,
		FeeProgramID,
	)
}

// AdminCtoInstruction re-points a coin's creator fees. Pass either newCreator or isHolderReward.
// A non-zero currentCreator is marked writable so the outgoing creator can be paid.
func (s *Client) AdminCtoInstruction(adminSetCreatorAuthority, mint, currentCreator, quoteMint, quoteTokenProgram solana.PublicKey, isHolderReward *bool, creatorFeeBps *uint64, newCreator *solana.PublicKey) (solana.Instruction, error) {
	quote := NormalizeQuoteMint(quoteMint)
	ata := func(owner solana.PublicKey) solana.PublicKey {
		return quoteATA(owner, quote, quoteTokenProgram)
	}
	creatorVault := DeriveCreatorVault(currentCreator)
	holderVault := DeriveCreatorVault(DeriveHolderRewards(mint))
	coinCreatorVault := DeriveCoinCreatorVaultAuthority(currentCreator)
	ix, err := pump.NewAdminCtoInstruction(
		isHolderReward,
		creatorFeeBps,
		newCreator,
		adminSetCreatorAuthority,
		s.Global,
		mint,
		quote,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
		solana.SystemProgramID,
		DeriveBondingCurve(mint),
		currentCreator,
		ata(currentCreator),
		creatorVault,
		ata(creatorVault),
		holderVault,
		ata(holderVault),
		AmmProgramID,
		s.AmmGlobalConfig,
		DerivePoolAuthority(mint),
		DeriveCanonicalPoolPDAWithQuote(DerivePoolAuthority(mint), mint, quote),
		s.AmmEventAuthority,
		coinCreatorVault,
		ata(coinCreatorVault),
		DeriveFeesFeeSharingConfig(mint),
		FeeProgramID,
		s.FeesEventAuthority,
		s.EventAuthority,
		ProgramID,
	)
	if err != nil {
		return nil, err
	}
	if !currentCreator.Equals(solana.PublicKey{}) {
		gi := ix.(*solana.GenericInstruction)
		for _, meta := range gi.AccountValues {
			if meta.PublicKey.Equals(currentCreator) {
				meta.IsWritable = true
			}
		}
	}
	return ix, nil
}

func quoteATA(owner, mint, tokenProgram solana.PublicKey) solana.PublicKey {
	return helpers.FindAssociatedTokenAddress(owner, mint, tokenProgram)
}
