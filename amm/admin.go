package amm

import (
	"context"
	"errors"

	pump_amm "github.com/krazyTry/pump-go/gen/amm"
	solana "github.com/solana-foundation/solana-go/v2"
	associatedtokenaccount "github.com/solana-foundation/solana-go/v2/programs/associated-token-account"
	"github.com/solana-foundation/solana-go/v2/programs/system"
)

func fixedRecipients(in []solana.PublicKey) ([8]solana.PublicKey, error) {
	var out [8]solana.PublicKey
	if len(in) != len(out) {
		return out, errors.New("protocol fee recipients must contain 8 keys")
	}
	copy(out[:], in)
	return out, nil
}

// CreateConfigInstruction builds create_config.
func (s *Client) CreateConfigInstruction(lpFeeBps, protocolFeeBps uint64, protocolFeeRecipients []solana.PublicKey, coinCreatorFeeBps uint64, admin, adminSetCoinCreatorAuthority solana.PublicKey) (solana.Instruction, error) {
	recipients, err := fixedRecipients(protocolFeeRecipients)
	if err != nil {
		return nil, err
	}
	return pump_amm.NewCreateConfigInstruction(
		lpFeeBps,
		protocolFeeBps,
		recipients,
		coinCreatorFeeBps,
		adminSetCoinCreatorAuthority,
		admin,
		s.GlobalConfig,
		system.ProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

// DisableInstruction builds disable.
func (s *Client) DisableInstruction(disableCreatePool, disableDeposit, disableWithdraw, disableBuy, disableSell bool, admin solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewDisableInstruction(
		disableCreatePool,
		disableDeposit,
		disableWithdraw,
		disableBuy,
		disableSell,
		admin,
		s.GlobalConfig,
		s.EventAuthority,
		ProgramID,
	)
}

// UpdateAdminInstruction builds update_admin.
func (s *Client) UpdateAdminInstruction(admin, newAdmin solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewUpdateAdminInstruction(
		admin,
		s.GlobalConfig,
		newAdmin,
		s.EventAuthority,
		ProgramID,
	)
}

// UpdateFeeConfigInstruction builds update_fee_config.
func (s *Client) UpdateFeeConfigInstruction(lpFeeBps, protocolFeeBps uint64, protocolFeeRecipients []solana.PublicKey, coinCreatorFeeBps uint64, admin, adminSetCoinCreatorAuthority solana.PublicKey) (solana.Instruction, error) {
	recipients, err := fixedRecipients(protocolFeeRecipients)
	if err != nil {
		return nil, err
	}
	return pump_amm.NewUpdateFeeConfigInstruction(
		lpFeeBps,
		protocolFeeBps,
		recipients,
		coinCreatorFeeBps,
		adminSetCoinCreatorAuthority,
		admin,
		s.GlobalConfig,
		s.EventAuthority,
		ProgramID,
	)
}

// UpdateCreatorFeeConfigInstruction builds update_creator_fee_config. admin signs and pays the rent.
func (s *Client) UpdateCreatorFeeConfigInstruction(admin solana.PublicKey, creatorFeeConfigurable bool, maxConfigurableCreatorFeeBps uint64) (solana.Instruction, error) {
	return pump_amm.NewUpdateCreatorFeeConfigInstruction(
		creatorFeeConfigurable,
		maxConfigurableCreatorFeeBps,
		admin,
		s.GlobalConfig,
		system.ProgramID,
		s.EventAuthority,
		ProgramID,
	)
}

// UpdateCreatorFeeConfig reads GlobalConfig.admin and builds update_creator_fee_config.
func (s *Client) UpdateCreatorFeeConfig(ctx context.Context, creatorFeeConfigurable bool, maxConfigurableCreatorFeeBps uint64) (solana.Instruction, error) {
	cfg, err := s.FetchGlobalConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s.UpdateCreatorFeeConfigInstruction(cfg.Admin, creatorFeeConfigurable, maxConfigurableCreatorFeeBps)
}

// AdminUpdateTokenIncentivesInstruction builds admin_update_token_incentives.
// A zero mint uses the pump mint. A zero token program uses Token-2022. A zero secondsInADay uses 86400.
func (s *Client) AdminUpdateTokenIncentivesInstruction(admin, mint, tokenProgram solana.PublicKey, startTime, endTime, secondsInADay int64, dayNumber, tokenSupplyPerDay uint64) (solana.Instruction, error) {
	if mint.Equals(solana.PublicKey{}) {
		mint = PumpMint
	}
	if tokenProgram.Equals(solana.PublicKey{}) {
		tokenProgram = solana.Token2022ProgramID
	}
	if secondsInADay == 0 {
		secondsInADay = 86400
	}
	return pump_amm.NewAdminUpdateTokenIncentivesInstruction(
		startTime,
		endTime,
		secondsInADay,
		dayNumber,
		tokenSupplyPerDay,
		admin,
		s.GlobalConfig,
		s.GlobalVolumeAccumulator,
		mint,
		CoinCreatorVaultAta(s.GlobalVolumeAccumulator, mint, tokenProgram),
		associatedtokenaccount.ProgramID,
		system.ProgramID,
		tokenProgram,
		s.EventAuthority,
		ProgramID,
	)
}

// AdminUpdateTokenIncentives reads GlobalConfig.admin and builds admin_update_token_incentives.
func (s *Client) AdminUpdateTokenIncentives(ctx context.Context, mint, tokenProgram solana.PublicKey, startTime, endTime, secondsInADay int64, dayNumber, tokenSupplyPerDay uint64) (solana.Instruction, error) {
	cfg, err := s.FetchGlobalConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s.AdminUpdateTokenIncentivesInstruction(cfg.Admin, mint, tokenProgram, startTime, endTime, secondsInADay, dayNumber, tokenSupplyPerDay)
}
