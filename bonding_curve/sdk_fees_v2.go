package bonding_curve

import (
	"errors"
	"fmt"

	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
	fees "github.com/krazyTry/pump-go/gen/fees"
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

func validateShareholders(next []Shareholder) error {
	if len(next) == 0 {
		return errors.New("no shareholders provided")
	}
	if len(next) > MaxShareholders {
		return fmt.Errorf("too many shareholders: max=%d got=%d", MaxShareholders, len(next))
	}
	total := 0
	seen := map[string]struct{}{}
	for _, sh := range next {
		if sh.ShareBps <= 0 {
			return fmt.Errorf("zero or negative share not allowed for address %s", sh.Address.String())
		}
		total += int(sh.ShareBps)
		seen[sh.Address.String()] = struct{}{}
	}
	if total != 10000 {
		return fmt.Errorf("invalid share total: expected 10000 bps got %d", total)
	}
	if len(seen) != len(next) {
		return errors.New("duplicate shareholder addresses not allowed")
	}
	return nil
}

func feeShareholders(next []Shareholder) []fees.Shareholder {
	out := make([]fees.Shareholder, len(next))
	for i, sh := range next {
		out[i] = fees.Shareholder{Address: sh.Address, ShareBps: sh.ShareBps}
	}
	return out
}

// UpdateFeeSharesV2Instruction replaces a sharing config's shareholders for any quote mint.
// currentShareholders are the addresses currently on the config. bondingCurveComplete appends the canonical pool.
func (s *Client) UpdateFeeSharesV2Instruction(authority, mint, quoteMint, quoteTokenProgram solana.PublicKey, current []solana.PublicKey, next []Shareholder, bondingCurveComplete bool) (solana.Instruction, error) {
	if err := validateShareholders(next); err != nil {
		return nil, err
	}
	quote := NormalizeQuoteMint(quoteMint)
	sharing := DeriveFeesFeeSharingConfig(mint)
	coinCreatorVault := DeriveCoinCreatorVaultAuthority(sharing)
	ix, err := fees.NewUpdateFeeSharesV2Instruction(
		feeShareholders(next),
		s.FeesEventAuthority,
		FeeProgramID,
		authority,
		s.Global,
		mint,
		sharing,
		DeriveBondingCurve(mint),
		DeriveCreatorVault(sharing),
		helpers.FindAssociatedTokenAddress(DeriveCreatorVault(sharing), quote, quoteTokenProgram),
		solana.SystemProgramID,
		ProgramID,
		s.EventAuthority,
		AmmProgramID,
		s.AmmEventAuthority,
		quote,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
		coinCreatorVault,
		helpers.FindAssociatedTokenAddress(coinCreatorVault, quote, quoteTokenProgram),
	)
	if err != nil {
		return nil, err
	}
	gi := ix.(*solana.GenericInstruction)
	for _, pubkey := range current {
		gi.AccountValues = append(gi.AccountValues, solana.NewAccountMeta(pubkey, true, false))
	}
	if !quote.Equals(solana.WrappedSol) {
		for _, pubkey := range current {
			gi.AccountValues = append(gi.AccountValues, solana.NewAccountMeta(helpers.FindAssociatedTokenAddress(pubkey, quote, quoteTokenProgram), true, false))
		}
	}
	if bondingCurveComplete {
		pool := DeriveCanonicalPoolPDAWithQuote(DerivePoolAuthority(mint), mint, quote)
		gi.AccountValues = append(gi.AccountValues, solana.NewAccountMeta(pool, false, false))
	}
	return ix, nil
}

// DistributeCreatorFeesV2Instruction pays a sharing config's shareholders in quoteMint.
func (s *Client) DistributeCreatorFeesV2Instruction(payer, mint, sharingConfigAddress, quoteMint, quoteTokenProgram solana.PublicKey, sharing *SharingConfig, shouldInitializeAta bool) (solana.Instruction, error) {
	quote := NormalizeQuoteMint(quoteMint)
	ix, err := pump.NewDistributeCreatorFeesV2Instruction(
		shouldInitializeAta,
		payer,
		mint,
		DeriveBondingCurve(mint),
		sharingConfigAddress,
		DeriveCreatorVault(sharingConfigAddress),
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
		helpers.FindAssociatedTokenAddress(DeriveCreatorVault(sharingConfigAddress), quote, quoteTokenProgram),
		quote,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
	)
	if err != nil {
		return nil, err
	}
	gi := ix.(*solana.GenericInstruction)
	for _, sh := range sharing.Shareholders {
		gi.AccountValues = append(gi.AccountValues, solana.NewAccountMeta(sh.Address, true, false))
	}
	if !quote.Equals(solana.WrappedSol) {
		for _, sh := range sharing.Shareholders {
			gi.AccountValues = append(gi.AccountValues, solana.NewAccountMeta(helpers.FindAssociatedTokenAddress(sh.Address, quote, quoteTokenProgram), true, false))
		}
	}
	return ix, nil
}

// TransferCreatorFeesToPumpV2Instruction sweeps AMM creator fees for a fee-shared coin into the pump creator vault.
func (s *Client) TransferCreatorFeesToPumpV2Instruction(payer, mint, quoteMint, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	quote := NormalizeQuoteMint(quoteMint)
	sharing := DeriveFeesFeeSharingConfig(mint)
	coinCreatorVault := DeriveCoinCreatorVaultAuthority(sharing)
	pumpVault := DeriveCreatorVault(sharing)
	return ammgen.NewTransferCreatorFeesToPumpV2Instruction(
		payer,
		quote,
		quoteTokenProgram,
		solana.SystemProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		sharing,
		coinCreatorVault,
		helpers.FindAssociatedTokenAddress(coinCreatorVault, quote, quoteTokenProgram),
		pumpVault,
		helpers.FindAssociatedTokenAddress(pumpVault, quote, quoteTokenProgram),
		s.AmmEventAuthority,
		AmmProgramID,
	)
}

// CollectCreatorFeeV2Instruction collects creator fees of a non-SOL quote into the creator's quote ATA.
func (s *Client) CollectCreatorFeeV2Instruction(creator, quoteMint, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	quote := NormalizeQuoteMint(quoteMint)
	vault := DeriveCreatorVault(creator)
	return pump.NewCollectCreatorFeeV2Instruction(
		creator,
		helpers.FindAssociatedTokenAddress(creator, quote, quoteTokenProgram),
		vault,
		helpers.FindAssociatedTokenAddress(vault, quote, quoteTokenProgram),
		quote,
		quoteTokenProgram,
		solana.SPLAssociatedTokenAccountProgramID,
		solana.SystemProgramID,
		s.EventAuthority,
		ProgramID,
	)
}
