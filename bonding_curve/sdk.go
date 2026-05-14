package bonding_curve

import (
	"errors"
	"fmt"

	solana "github.com/gagliardetto/solana-go"
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	amm "github.com/krazyTry/pump-go/gen/amm"
	fees "github.com/krazyTry/pump-go/gen/fees"
	pump "github.com/krazyTry/pump-go/gen/pump"
)

// SPL-Token
func (s *Client) CreateInstruction(creator, user, baseMint solana.PublicKey, name, symbol, uri string) (solana.Instruction, error) {

	mintAuthority := DeriveMintAuthority()

	metadata := DeriveMintMetadata(baseMint)
	bondingCurve := DeriveBondingCurve(baseMint)

	bondingCurveAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, solana.TokenProgramID)

	return pump.NewCreateInstruction(
		name,
		symbol,
		uri,
		creator,
		baseMint,
		mintAuthority,
		bondingCurve,
		bondingCurveAta,
		s.Global,
		MetaplexProgramID,
		metadata,
		user,
		solana.SystemProgramID,
		solana.TokenProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		solana.SysVarRentPubkey,
		s.EventAuthority,
		ProgramID,
	)
}

// Deprecated: Use GetBuyV2Instruction.
func (s *Client) GetBuyInstruction(creator, user, baseMint, userTokenAccount solana.PublicKey, amount, maxSolCost uint64, feeRecipient solana.PublicKey) (solana.Instruction, error) {
	creatorVault := DeriveCreatorVault(creator)

	bondingCurve := DeriveBondingCurve(baseMint)
	userVolumeAccumulator := DeriveUserVolumeAccumulator(user)

	bondingCurveAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, solana.TokenProgramID)

	return pump.NewBuyInstruction(
		amount,
		maxSolCost,
		pump.OptionBool{V0: true},
		s.Global,
		feeRecipient,
		baseMint,
		bondingCurve,
		bondingCurveAta,
		userTokenAccount,
		user,
		solana.SystemProgramID,
		solana.TokenProgramID,
		creatorVault,
		s.EventAuthority,
		ProgramID,
		s.GlobalVolumeAccumulator,
		userVolumeAccumulator,
		s.FeeConfig,
		fees.ProgramID,
	)
}

// Deprecated: Use GetSellV2Instruction.
func (s *Client) GetSellInstruction(creator, user, baseMint, userTokenAccount solana.PublicKey, amount, minSolOutput uint64, feeRecipient solana.PublicKey) (solana.Instruction, error) {
	bondingCurve := DeriveBondingCurve(baseMint)
	creatorVault := DeriveCreatorVault(creator)

	bondingCurveAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, solana.TokenProgramID)

	return pump.NewSellInstruction(
		amount,
		minSolOutput,
		s.Global,
		feeRecipient,
		baseMint,
		bondingCurve,
		bondingCurveAta,
		userTokenAccount,
		user,
		solana.SystemProgramID,
		creatorVault,
		solana.TokenProgramID,
		s.EventAuthority,
		ProgramID,
		s.FeeConfig,
		fees.ProgramID,
	)
}

// Deprecated: Use MigrateV2Instruction.
func (c *Client) MigrateInstruction(withdrawAuthority, user, baseMint, tokenProgram solana.PublicKey) (solana.Instruction, error) {
	if tokenProgram.Equals(solana.PublicKey{}) {
		tokenProgram = solana.TokenProgramID
	}

	poolAuthority := DerivePoolAuthority(baseMint)

	poolAuthorityAta := helpers.FindAssociatedTokenAddress(poolAuthority, baseMint, tokenProgram)
	poolAuthorityWSolAta := helpers.FindAssociatedTokenAddress(poolAuthority, solana.WrappedSol, solana.TokenProgramID)

	bondingCurve := DeriveBondingCurve(baseMint)
	bondingCurveAta := helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, tokenProgram)

	pool := DeriveCanonicalPoolPDAWithQuote(poolAuthority, baseMint, solana.WrappedSol)

	lpMint := DeriveAmmLpMint(pool)

	userPoolTokenAta := helpers.FindAssociatedTokenAddress(user, lpMint, solana.Token2022ProgramID)
	poolBaseTokenAta := helpers.FindAssociatedTokenAddress(pool, baseMint, tokenProgram)
	poolQuoteTokenAta := helpers.FindAssociatedTokenAddress(pool, solana.WrappedSol, solana.TokenProgramID)

	return pump.NewMigrateInstruction(
		c.Global,
		withdrawAuthority,
		baseMint,
		bondingCurve,
		bondingCurveAta,
		user,
		solana.SystemProgramID,
		tokenProgram,
		AmmProgramID,
		pool,
		poolAuthority,
		poolAuthorityAta,
		poolAuthorityWSolAta,
		c.AmmGlobalConfig,
		solana.WrappedSol,
		lpMint,
		userPoolTokenAta,
		poolBaseTokenAta,
		poolQuoteTokenAta,
		solana.Token2022ProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		c.AmmEventAuthority,
		c.EventAuthority,
		ProgramID,
		solana.SysVarRentPubkey,
	)
}

// Deprecated: Use ClaimCashbackV2Instruction.
func (s *Client) ClaimCashbackInstruction(user solana.PublicKey) (solana.Instruction, error) {
	userVolumeAccumulator := DeriveUserVolumeAccumulator(user)

	return pump.NewClaimCashbackInstruction(
		user,
		userVolumeAccumulator,
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

func (c *Client) CollectCreatorFeeInstruction(creator solana.PublicKey) (solana.Instruction, error) {
	creatorVault := DeriveCreatorVault(creator)
	return pump.NewCollectCreatorFeeInstruction(
		creator,
		creatorVault,
		solana.SystemProgramID,
		c.EventAuthority,
		ProgramID,
	)
}

func (s *Client) UpdateFeeShares(authority, mint solana.PublicKey, next []pump.Shareholder) (solana.Instruction, error) {
	if len(next) == 0 {
		return nil, errors.New("no shareholders provided")
	}

	if len(next) > MaxShareholders {
		return nil, fmt.Errorf("too many shareholders: max=%d got=%d", MaxShareholders, len(next))
	}

	total := 0
	seen := map[string]struct{}{}
	for _, sh := range next {
		if sh.ShareBps <= 0 {
			return nil, fmt.Errorf("zero or negative share not allowed for address %s", sh.Address.String())
		}
		total += int(sh.ShareBps)
		seen[sh.Address.String()] = struct{}{}
	}
	if total != 10000 {
		return nil, fmt.Errorf("invalid share total: expected 10000 bps got %d", total)
	}
	if len(seen) != len(next) {
		return nil, errors.New("duplicate shareholder addresses not allowed")
	}

	sharing := DeriveFeesFeeSharingConfig(mint)
	feeShareholders := make([]fees.Shareholder, 0, len(next))
	for _, sh := range next {
		feeShareholders = append(feeShareholders, fees.Shareholder{Address: sh.Address, ShareBps: sh.ShareBps})
	}
	coinCreatorVaultAuthority := DeriveCoinCreatorVaultAuthority(sharing)

	bondingCurve := DeriveBondingCurve(mint)
	creatorVault := DeriveCreatorVault(sharing)

	return fees.NewUpdateFeeSharesInstruction(
		feeShareholders,
		s.FeesEventAuthority,
		fees.ProgramID,
		authority,
		s.Global,
		mint,
		sharing,
		bondingCurve,
		creatorVault,
		solana.SystemProgramID,
		pump.ProgramID,
		s.EventAuthority,
		amm.ProgramID,
		s.AmmEventAuthority,
		solana.WrappedSol,
		solana.TokenProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		coinCreatorVaultAuthority,
		helpers.FindAssociatedTokenAddress(coinCreatorVaultAuthority, solana.WrappedSol, solana.TokenProgramID),
	)
}

func (s *Client) ExtendAccountInstruction(account, user solana.PublicKey) (solana.Instruction, error) {
	return pump.NewExtendAccountInstruction(
		account,
		user,
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) SetCreatorInstruction(mint, setCreatorAuthority, creator solana.PublicKey) (solana.Instruction, error) {

	metadata := DeriveMintMetadata(mint)
	bondingCurve := DeriveBondingCurve(mint)
	return pump.NewSetCreatorInstruction(
		creator,
		setCreatorAuthority,
		s.Global,
		mint,
		metadata,
		bondingCurve,
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) InitUserVolumeAccumulatorInstruction(payer, user solana.PublicKey) (solana.Instruction, error) {
	return pump.NewInitUserVolumeAccumulatorInstruction(
		payer,
		user,
		DeriveUserVolumeAccumulator(user),
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) CloseUserVolumeAccumulatorInstruction(user solana.PublicKey) (solana.Instruction, error) {
	return pump.NewCloseUserVolumeAccumulatorInstruction(
		user,
		DeriveUserVolumeAccumulator(user),
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) SyncUserVolumeAccumulatorInstruction(user solana.PublicKey) (solana.Instruction, error) {
	return pump.NewSyncUserVolumeAccumulatorInstruction(
		user,
		DeriveGlobalVolumeAccumulator(),
		DeriveUserVolumeAccumulator(user),
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) CreateFeeSharingConfigInstruction(creator, mint solana.PublicKey, pool solana.PublicKey) (solana.Instruction, error) {
	bondingCurve := DeriveBondingCurve(mint)
	sharingConfig := DeriveFeesFeeSharingConfig(mint)

	return fees.NewCreateFeeSharingConfigInstruction(
		s.FeesEventAuthority,
		FeeProgramID,
		creator,
		s.Global,
		mint,
		sharingConfig,
		solana.SystemProgramID,
		bondingCurve,
		ProgramID,
		s.EventAuthority,
		pool,
		AmmProgramID,
		s.AmmEventAuthority,
	)
}

func (s *Client) GetMinimumDistributableFeeInstruction(mint solana.PublicKey) (solana.Instruction, error) {
	sharingConfig := DeriveFeesFeeSharingConfig(mint)
	creatorVault := DeriveCreatorVault(sharingConfig)
	bondingCurve := DeriveBondingCurve(mint)

	return pump.NewGetMinimumDistributableFeeInstruction(
		mint,
		bondingCurve,
		sharingConfig,
		creatorVault,
	)
}

func (s *Client) DistributeCreatorFeesInstruction(mint solana.PublicKey) (solana.Instruction, error) {
	sharingConfig := DeriveFeesFeeSharingConfig(mint)
	creatorVault := DeriveCreatorVault(sharingConfig)
	bondingCurve := DeriveBondingCurve(mint)

	return pump.NewDistributeCreatorFeesInstruction(
		mint,
		bondingCurve,
		sharingConfig,
		creatorVault,
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) CreateDonationFeePdaInstruction(coinCreator, baseMint, configID solana.PublicKey) (solana.Instruction, error) {
	donationFeePda := DeriveFeesDonationFeePDA(baseMint, configID)

	poolAuthority := DerivePoolAuthority(baseMint)

	poolAccount := DeriveCanonicalPoolPDAWithQuote(poolAuthority, baseMint, solana.WrappedSol)
	bondingCurve := DeriveBondingCurve(baseMint)
	sharingConfig := DeriveFeesFeeSharingConfig(baseMint)

	return fees.NewCreateDonationFeePdaInstruction(
		s.FeesEventAuthority,
		FeeProgramID,
		coinCreator,
		solana.SystemProgramID,
		s.FeesFeeProgramGlobal,
		donationFeePda,
		configID,
		baseMint,
		bondingCurve,
		poolAccount,
		sharingConfig, // or solana.PublicKey{} ?
	)
}

func (s *Client) CrankDonationFeePdaInstruction(payer, mint, configID solana.PublicKey) (solana.Instruction, error) {
	donationFeePda := DeriveFeesDonationFeePDA(mint, configID)
	donationFeePdaAta := helpers.FindAssociatedTokenAddress(donationFeePda, solana.WrappedSol, solana.TokenProgramID)

	debouncer := DeriveDonationRelayDebouncerPDA(configID, solana.WrappedSol)
	debouncerAta := helpers.FindAssociatedTokenAddress(debouncer, solana.WrappedSol, solana.TokenProgramID)

	return fees.NewCrankDonationFeePdaInstruction(
		s.FeesEventAuthority,
		FeeProgramID,
		payer,
		solana.SystemProgramID,
		solana.TokenProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		solana.SysVarRentPubkey,
		s.FeesFeeProgramGlobal,
		donationFeePda,
		solana.WrappedSol,
		donationFeePdaAta,
		DonationRelayProgramID,
		DeriveDonationRelayEventAuthority(),
		DeriveDonationRelayMintWhitelistPDA(),
		DeriveDonationRelayEpochTrackerPDA(configID, solana.WrappedSol),
		debouncer,
		debouncerAta,
	)
}

func (s *Client) CreateSocialFeePdaInstruction(payer solana.PublicKey, userID string, platform uint8) (solana.Instruction, error) {
	socialFeePda := DeriveFeesSocialFeePDA(userID, platform)

	return fees.NewCreateSocialFeePdaInstruction(
		userID,
		platform,
		payer,
		socialFeePda,
		solana.SystemProgramID,
		s.FeesFeeProgramGlobal,
		s.FeesEventAuthority,
		FeeProgramID,
	)
}

func (s *Client) ClaimSocialFeePdaInstruction(recipient, socialClaimAuthority solana.PublicKey, userID string, platform uint8) (solana.Instruction, error) {
	socialFeePda := DeriveFeesSocialFeePDA(userID, platform)

	return fees.NewClaimSocialFeePdaInstruction(
		userID,
		platform,
		recipient,
		socialFeePda,
		s.FeesFeeProgramGlobal,
		socialClaimAuthority,
		s.FeesEventAuthority,
		FeeProgramID,
	)
}
