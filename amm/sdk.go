package amm

import (
	solana "github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/krazyTry/pump-go/amm/helpers"
	pump_amm "github.com/krazyTry/pump-go/gen/amm"
)

// type AdminSDK struct{}

// func NewAdminSDK() *AdminSDK { return &AdminSDK{} }

// func toFixed8(in []solana.PublicKey) [8]solana.PublicKey {
// 	var out [8]solana.PublicKey
// 	copy(out[:], in)
// 	return out
// }

// func (s *AdminSDK) CreateConfig(lpFeeBps, protocolFeeBps uint64, protocolFeeRecipients []solana.PublicKey, coinCreatorFeeBps uint64, admin, adminSetCoinCreatorAuthority, globalConfigPDA, eventAuthorityPDA solana.PublicKey) (solana.Instruction, error) {
// 	return pump_amm.NewCreateConfigInstruction(
// 		lpFeeBps,
// 		protocolFeeBps,
// 		toFixed8(protocolFeeRecipients),
// 		coinCreatorFeeBps,
// 		adminSetCoinCreatorAuthority,
// 		admin,
// 		globalConfigPDA,
// 		system.ProgramID,
// 		eventAuthorityPDA,
// 		AmmProgramID,
// 	)
// }

// func (s *AdminSDK) Disable(disableCreatePool, disableDeposit, disableWithdraw, disableBuy, disableSell bool, admin, globalConfigPDA, eventAuthorityPDA solana.PublicKey) (solana.Instruction, error) {
// 	return pump_amm.NewDisableInstruction(
// 		disableCreatePool,
// 		disableDeposit,
// 		disableWithdraw,
// 		disableBuy,
// 		disableSell,
// 		admin,
// 		globalConfigPDA,
// 		eventAuthorityPDA,
// 		AmmProgramID,
// 	)
// }

// func (s *AdminSDK) UpdateAdmin(admin, newAdmin, globalConfigPDA, eventAuthorityPDA solana.PublicKey) (solana.Instruction, error) {
// 	return pump_amm.NewUpdateAdminInstruction(
// 		admin,
// 		newAdmin,
// 		globalConfigPDA,
// 		eventAuthorityPDA,
// 		AmmProgramID,
// 	)
// }

// func (s *AdminSDK) UpdateFeeConfig(lpFeeBps, protocolFeeBps uint64, protocolFeeRecipients []solana.PublicKey, coinCreatorFeeBps uint64, admin, adminSetCoinCreatorAuthority, globalConfigPDA, eventAuthorityPDA solana.PublicKey) (solana.Instruction, error) {
// 	return pump_amm.NewUpdateFeeConfigInstruction(
// 		lpFeeBps,
// 		protocolFeeBps,
// 		toFixed8(protocolFeeRecipients),
// 		coinCreatorFeeBps,
// 		adminSetCoinCreatorAuthority,
// 		admin,
// 		globalConfigPDA,
// 		eventAuthorityPDA,
// 		AmmProgramID,
// 	)
// }

// func (s *AdminSDK) AdminSetCoinCreator(newCoinCreator, adminSetCoinCreatorAuthority, pool, globalConfigPDA, eventAuthorityPDA solana.PublicKey) (solana.Instruction, error) {
// 	return pump_amm.NewAdminSetCoinCreatorInstruction(
// 		newCoinCreator,
// 		adminSetCoinCreatorAuthority,
// 		globalConfigPDA,
// 		pool,
// 		eventAuthorityPDA,
// 		AmmProgramID,
// 	)
// }

// func (s *AdminSDK) AdminUpdateTokenIncentives(startTime, endTime, secondsInDay int64, dayNumber, tokenSupplyPerDay uint64, admin, mint, globalIncentiveTokenAccount, tokenProgram, globalConfigPDA,
// 	globalVolumeAccumulatorPDA, eventAuthorityPDA solana.PublicKey) (solana.Instruction, error) {
// 	return pump_amm.NewAdminUpdateTokenIncentivesInstruction(
// 		startTime,
// 		endTime,
// 		secondsInDay,
// 		dayNumber,
// 		tokenSupplyPerDay,
// 		admin,
// 		globalConfigPDA,
// 		globalVolumeAccumulatorPDA,
// 		mint,
// 		globalIncentiveTokenAccount,
// 		associatedtokenaccount.ProgramID,
// 		system.ProgramID,
// 		tokenProgram,
// 		eventAuthorityPDA,
// 		AmmProgramID,
// 	)
// }

type CreatePoolParams struct {
	Index        uint16
	BaseIn       uint64
	QuoteIn      uint64
	IsMayhemMode bool
	IsCashback   bool
}

func (s *Client) CreatePoolInstruction(
	params *CreatePoolParams,

	creator solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	userBaseTokenAccount solana.PublicKey,
	userQuoteTokenAccount solana.PublicKey,
	baseTokenProgram solana.PublicKey,
	quoteTokenProgram solana.PublicKey,
) ([]solana.Instruction, error) {

	pool := DeriveCanonicalPumpPoolPDA(params.Index, creator, baseMint, quoteMint)
	lpMint := DeriveLpMint(pool)

	userPoolTokenAccount := helpers.FindAssociatedTokenAddress(creator, lpMint, solana.Token2022ProgramID)
	poolBaseTokenAccount := helpers.FindAssociatedTokenAddress(pool, baseMint, baseTokenProgram)
	poolQuoteTokenAccount := helpers.FindAssociatedTokenAddress(pool, quoteMint, quoteTokenProgram)

	createIx, err := pump_amm.NewCreatePoolInstruction(
		params.Index,
		params.BaseIn,
		params.QuoteIn,
		system.ProgramID,
		params.IsMayhemMode,
		pump_amm.OptionBool{V0: params.IsCashback},
		pool,
		s.GlobalConfig,
		creator,
		baseMint,
		quoteMint,
		lpMint,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		userPoolTokenAccount,
		poolBaseTokenAccount,
		poolQuoteTokenAccount,
		system.ProgramID,
		solana.Token2022ProgramID,
		baseTokenProgram,
		quoteTokenProgram,
		associatedtokenaccount.ProgramID,
		s.EventAuthority,
		ProgramID,
	)
	if err != nil {
		return nil, err
	}

	extendAccountIx, err := s.ExtendAccountInstruction(pool, creator)
	if err != nil {
		return nil, err
	}

	return []solana.Instruction{createIx, extendAccountIx}, nil
}

type LiquidityAccounts struct {
	pool                  solana.PublicKey
	user                  solana.PublicKey
	baseMint              solana.PublicKey
	quoteMint             solana.PublicKey
	lpMint                solana.PublicKey
	userBaseTokenAccount  solana.PublicKey
	userQuoteTokenAccount solana.PublicKey
	userPoolTokenAccount  solana.PublicKey
	poolBaseTokenAccount  solana.PublicKey
	poolQuoteTokenAccount solana.PublicKey
	tokenProgram          solana.PublicKey
}

type DepositParams struct {
	LpOut      uint64
	MaxBaseIn  uint64
	MaxQuoteIn uint64
}

func (s *Client) DepositInstruction(
	params *DepositParams,
	pool *AccountWithPool,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	userBaseTokenAccount solana.PublicKey,
	userQuoteTokenAccount solana.PublicKey,
	userPoolTokenAccount solana.PublicKey,
) (solana.Instruction, error) {
	return pump_amm.NewDepositInstruction(
		params.LpOut,
		params.MaxBaseIn,
		params.MaxQuoteIn,
		pool.PublicKey,
		s.GlobalConfig,
		user,
		baseMint,
		quoteMint,
		pool.Account.LpMint,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		userPoolTokenAccount,
		pool.Account.PoolBaseTokenAccount,
		pool.Account.PoolQuoteTokenAccount,
		solana.TokenProgramID,
		solana.Token2022ProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

type WithdrawParams struct {
	LpIn        uint64
	MinBaseOut  uint64
	MinQuoteOut uint64
}

func (s *Client) WithdrawInstruction(
	params *WithdrawParams,
	pool *AccountWithPool,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	userBaseTokenAccount solana.PublicKey,
	userQuoteTokenAccount solana.PublicKey,
	userPoolTokenAccount solana.PublicKey,
) (solana.Instruction, error) {
	return pump_amm.NewWithdrawInstruction(
		params.LpIn,
		params.MinBaseOut,
		params.MinQuoteOut,
		pool.PublicKey,
		s.GlobalConfig,
		user,
		baseMint,
		quoteMint,
		pool.Account.LpMint,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		userPoolTokenAccount,
		pool.Account.PoolBaseTokenAccount,
		pool.Account.PoolQuoteTokenAccount,
		solana.TokenProgramID,
		solana.Token2022ProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

type BuyParams struct {
	BaseOut     uint64
	MaxQuoteIn  uint64
	TrackVolume bool
}

func (s *Client) BuyInstruction(
	params *BuyParams,
	globalConfig *GlobalConfig,
	pool *AccountWithPool,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	baseTokenProgram solana.PublicKey,
	quoteTokenProgram solana.PublicKey,
	userBaseTokenAccount solana.PublicKey,
	userQuoteTokenAccount solana.PublicKey,
) (solana.Instruction, error) {

	userVolumeAccumulator := DeriveUserVolumeAccumulator(user)
	userVolumeAccumulatorAta := helpers.FindAssociatedTokenAddress(userVolumeAccumulator, quoteMint, quoteTokenProgram)

	coinCreatorVaultAuthority := DeriveCoinCreatorVault(pool.Account.CoinCreator)
	coinCreatorVaultAta := helpers.FindAssociatedTokenAddress(coinCreatorVaultAuthority, quoteMint, quoteTokenProgram)
	protocolFeeRecipient := GetFeeRecipient(globalConfig, pool.Account.IsMayhemMode)
	protocolFeeRecipientTokenAccount := helpers.FindAssociatedTokenAddress(protocolFeeRecipient, quoteMint, quoteTokenProgram)
	buybackFeeRecipient := GetBuybackFeeRecipient(globalConfig)
	buybackFeeRecipientTokenAccount := helpers.FindAssociatedTokenAddress(buybackFeeRecipient, quoteMint, quoteTokenProgram)

	ix, err := pump_amm.NewBuyInstruction(
		params.BaseOut,
		params.MaxQuoteIn,
		pump_amm.OptionBool{V0: params.TrackVolume},
		pool.PublicKey,
		user,
		s.GlobalConfig,
		baseMint,
		quoteMint,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		pool.Account.PoolBaseTokenAccount,
		pool.Account.PoolQuoteTokenAccount,
		protocolFeeRecipient,
		protocolFeeRecipientTokenAccount,
		baseTokenProgram,
		quoteTokenProgram,
		system.ProgramID,
		associatedtokenaccount.ProgramID,
		s.EventAuthority,
		ProgramID,
		coinCreatorVaultAta,
		coinCreatorVaultAuthority,
		s.GlobalVolumeAccumulator,
		userVolumeAccumulator,
		s.FeeConfig,
		FeeProgramID,
	)
	if err != nil {
		return nil, err
	}
	var remainingAccounts solana.AccountMetaSlice

	if pool.Account.IsCashbackCoin {
		remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(userVolumeAccumulatorAta, true, false))
	}

	if !pool.Account.CoinCreator.Equals(solana.PublicKey{}) {
		remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(DerivePoolV2PDA(baseMint), false, false))
	}

	remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(buybackFeeRecipient, false, false))
	remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(buybackFeeRecipientTokenAccount, true, false))

	if gi, ok := ix.(*solana.GenericInstruction); ok {
		gi.AccountValues = append(gi.AccountValues, remainingAccounts...)
	}
	return ix, nil
}

type SellParams struct {
	BaseIn      uint64
	MinQuoteOut uint64
}

func (s *Client) SellInstruction(
	params *SellParams,
	globalConfig *GlobalConfig,
	pool *AccountWithPool,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	baseTokenProgram solana.PublicKey,
	quoteTokenProgram solana.PublicKey,
	userBaseTokenAccount solana.PublicKey,
	userQuoteTokenAccount solana.PublicKey,
) (solana.Instruction, error) {

	userVolumeAccumulator := DeriveUserVolumeAccumulator(user)
	userVolumeAccumulatorAta := helpers.FindAssociatedTokenAddress(userVolumeAccumulator, quoteMint, quoteTokenProgram)
	buybackFeeRecipient := GetBuybackFeeRecipient(globalConfig)
	buybackFeeRecipientTokenAccount := helpers.FindAssociatedTokenAddress(buybackFeeRecipient, quoteMint, quoteTokenProgram)

	coinCreatorVaultAuthority := DeriveCoinCreatorVault(pool.Account.CoinCreator)
	coinCreatorVaultAta := helpers.FindAssociatedTokenAddress(coinCreatorVaultAuthority, quoteMint, quoteTokenProgram)
	protocolFeeRecipient := GetFeeRecipient(globalConfig, pool.Account.IsMayhemMode)
	protocolFeeRecipientTokenAccount := helpers.FindAssociatedTokenAddress(protocolFeeRecipient, quoteMint, quoteTokenProgram)

	ix, err := pump_amm.NewSellInstruction(
		params.BaseIn,
		params.MinQuoteOut,
		pool.PublicKey,
		user,
		s.GlobalConfig,
		baseMint,
		quoteMint,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		pool.Account.PoolBaseTokenAccount,
		pool.Account.PoolQuoteTokenAccount,
		protocolFeeRecipient,
		protocolFeeRecipientTokenAccount,
		baseTokenProgram,
		quoteTokenProgram,
		system.ProgramID,
		associatedtokenaccount.ProgramID,
		s.EventAuthority,
		ProgramID,
		coinCreatorVaultAta,
		coinCreatorVaultAuthority,
		s.FeeConfig,
		FeeProgramID,
	)
	if err != nil {
		return nil, err
	}

	var remainingAccounts solana.AccountMetaSlice

	if pool.Account.IsCashbackCoin {
		remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(userVolumeAccumulatorAta, true, false))
		remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(userVolumeAccumulator, true, false))
	}

	if !pool.Account.CoinCreator.Equals(solana.PublicKey{}) {
		remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(DerivePoolV2PDA(baseMint), false, false))
	}

	remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(buybackFeeRecipient, false, false))
	remainingAccounts = append(remainingAccounts, solana.NewAccountMeta(buybackFeeRecipientTokenAccount, true, false))

	if gi, ok := ix.(*solana.GenericInstruction); ok {
		gi.AccountValues = append(gi.AccountValues, remainingAccounts...)
	}
	return ix, nil
}

func (s *Client) CollectCoinCreatorFeeInstruction(quoteMint, quoteTokenProgram, coinCreator, coinCreatorVaultAuthority, coinCreatorVaultAta, coinCreatorTokenAccount solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewCollectCoinCreatorFeeInstruction(
		quoteMint,
		quoteTokenProgram,
		coinCreator,
		coinCreatorVaultAuthority,
		coinCreatorVaultAta,
		coinCreatorTokenAccount,
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) SetCoinCreatorInstruction(pool, metadata, bondingCurve solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewSetCoinCreatorInstruction(
		pool,
		metadata,
		bondingCurve,
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) SyncUserVolumeAccumulatorInstruction(user solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewSyncUserVolumeAccumulatorInstruction(
		user,
		s.GlobalVolumeAccumulator,
		DeriveUserVolumeAccumulator(user),
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) InitUserVolumeAccumulatorInstruction(payer, user solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewInitUserVolumeAccumulatorInstruction(
		payer,
		user,
		DeriveUserVolumeAccumulator(user),
		system.ProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) CloseUserVolumeAccumulatorInstruction(user solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewCloseUserVolumeAccumulatorInstruction(
		user,
		DeriveUserVolumeAccumulator(user),
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) ClaimTokenIncentivesInstruction(user, payer, userAta, globalIncentiveTokenAccount, mint, tokenProgram solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewClaimTokenIncentivesInstruction(
		user,
		userAta,
		s.GlobalVolumeAccumulator,
		globalIncentiveTokenAccount,
		DeriveUserVolumeAccumulator(user),
		mint,
		tokenProgram,
		system.ProgramID,
		associatedtokenaccount.ProgramID,
		s.EventAuthority,
		ProgramID,
		payer,
	)
}

func (s *Client) ExtendAccountInstruction(account, user solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewExtendAccountInstruction(
		account,
		user,
		system.ProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

func (s *Client) TransferCreatorFeesToPumpInstruction(coinCreator solana.PublicKey) (solana.Instruction, error) {
	coinCreatorVaultAuthority := DeriveCoinCreatorVault(coinCreator)
	coinCreatorVaultAta := helpers.FindAssociatedTokenAddress(coinCreatorVaultAuthority, solana.WrappedSol, solana.TokenProgramID)
	pumpCreatorVault := DerivePumpCreatorVault(coinCreator)

	return pump_amm.NewTransferCreatorFeesToPumpInstruction(
		solana.WrappedSol,
		solana.TokenProgramID,
		system.ProgramID,
		associatedtokenaccount.ProgramID,
		coinCreator,
		coinCreatorVaultAuthority,
		coinCreatorVaultAta,
		pumpCreatorVault,
		s.EventAuthority,
		ProgramID,
	)
}
