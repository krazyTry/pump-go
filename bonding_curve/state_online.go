package bonding_curve

import (
	"context"
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
	ammclient "github.com/krazyTry/pump-go/amm"
	ammhelpers "github.com/krazyTry/pump-go/amm/helpers"
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/token"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

// DecodeGlobal decodes a pump Global account.
func DecodeGlobal(data []byte) (*Global, error) { return pump.ParseAccount_Global(data) }

// DecodeFeeConfig decodes a pump FeeConfig account.
func DecodeFeeConfig(data []byte) (*FeeConfig, error) { return pump.ParseAccount_FeeConfig(data) }

// DecodeBondingCurve decodes a bonding-curve account, including a short legacy account the parser accepts.
func DecodeBondingCurve(data []byte) (*BondingCurve, error) {
	return pump.ParseAccount_BondingCurve(data)
}

// DecodeQuoteControl decodes a QuoteControl account.
func DecodeQuoteControl(data []byte) (*pump.QuoteControl, error) {
	return pump.ParseAccount_QuoteControl(data)
}

// DecodeSharingConfig decodes a fee-sharing config account.
func DecodeSharingConfig(data []byte) (*SharingConfig, error) {
	return pump.ParseAccount_SharingConfig(data)
}

// FetchQuoteControl reads the QuoteControl PDA. A missing account returns nil, nil.
func (o *Client) FetchQuoteControl(ctx context.Context) (*pump.QuoteControl, error) {
	acc, err := o.RPC.GetAccountInfo(ctx, DeriveQuoteControl())
	if errors.Is(err, rpc.ErrNotFound) || acc == nil || acc.Value == nil {
		if err != nil && !errors.Is(err, rpc.ErrNotFound) {
			return nil, err
		}
		return nil, nil
	}
	return pump.ParseAccount_QuoteControl(acc.Value.Data.GetBinary())
}

// FetchQuoteTokenProgram is the program that owns quoteMint. SOL and the zero key are the Token program.
func (o *Client) FetchQuoteTokenProgram(ctx context.Context, quoteMint solana.PublicKey) (solana.PublicKey, error) {
	if IsLegacyQuoteMint(quoteMint) {
		return solana.TokenProgramID, nil
	}
	info, err := ammhelpers.GetTokenInfo(ctx, o.RPC, quoteMint)
	if err != nil {
		return solana.PublicKey{}, err
	}
	if !info.Owner.Equals(solana.TokenProgramID) && !info.Owner.Equals(solana.Token2022ProgramID) {
		return solana.PublicKey{}, fmt.Errorf("quote mint %s is owned by %s, not a token program", quoteMint, info.Owner)
	}
	return info.Owner, nil
}

// FetchQuoteTransferFee is the Token-2022 transfer fee quoteMint charges this epoch.
// SOL and mints without a transfer fee return nil.
func (o *Client) FetchQuoteTransferFee(ctx context.Context, quoteMint solana.PublicKey) (*QuoteTransferFee, error) {
	if IsLegacyQuoteMint(quoteMint) {
		return nil, nil
	}
	info, err := ammhelpers.GetTokenInfo(ctx, o.RPC, quoteMint)
	if err != nil {
		return nil, err
	}
	if info.Ext == nil || info.Ext.TransferFeeConfig == nil || info.Fee.FeeBps == 0 {
		return nil, nil
	}
	return &QuoteTransferFee{BasisPoints: uint64(info.Fee.FeeBps), MaximumFee: info.Fee.MaxFee}, nil
}

// SupportedQuoteMint is one mint create_v2 currently accepts.
type SupportedQuoteMint struct {
	Mint                        solana.PublicKey
	Source                      string
	InitialVirtualQuoteReserves uint64
}

// FetchSupportedQuoteMints lists SOL, Global's whitelist, then QuoteControl entries Global does not already list.
func (o *Client) FetchSupportedQuoteMints(ctx context.Context) ([]SupportedQuoteMint, error) {
	global, err := o.FetchGlobal(ctx)
	if err != nil {
		return nil, err
	}
	control, err := o.FetchQuoteControl(ctx)
	if err != nil {
		return nil, err
	}
	out := []SupportedQuoteMint{{
		Mint:                        solana.WrappedSol,
		Source:                      "sol",
		InitialVirtualQuoteReserves: global.InitialVirtualSolReserves,
	}}
	for _, mint := range global.WhitelistedQuoteMints {
		if IsLegacyQuoteMint(mint) {
			continue
		}
		out = append(out, SupportedQuoteMint{Mint: mint, Source: "global", InitialVirtualQuoteReserves: global.InitialVirtualQuoteReserves})
	}
	if control != nil {
		for _, entry := range control.Mints {
			seen := false
			for _, have := range out {
				if have.Mint.Equals(entry.Mint) {
					seen = true
					break
				}
			}
			if !seen {
				out = append(out, SupportedQuoteMint{Mint: entry.Mint, Source: "quoteControl", InitialVirtualQuoteReserves: entry.InitialVirtualQuoteReserves})
			}
		}
	}
	return out, nil
}

// ResolvedQuote is what a create_v2 builder needs for one quote mint.
type ResolvedQuote struct {
	SupportedQuoteMint
	TokenProgram solana.PublicKey
	PumpQuote    *PumpQuoteCurve
	Accounts     *PumpQuoteAccounts
}

// ResolveQuoteMint resolves one quote the way create_v2 admits it.
// A mint in neither list that is itself a pump coin resolves as pumpCoin.
func (o *Client) ResolveQuoteMint(ctx context.Context, quoteMint solana.PublicKey, global *Global) (*ResolvedQuote, error) {
	if global == nil {
		var err error
		global, err = o.FetchGlobal(ctx)
		if err != nil {
			return nil, err
		}
	}
	if IsLegacyQuoteMint(quoteMint) {
		return &ResolvedQuote{
			SupportedQuoteMint: SupportedQuoteMint{Mint: solana.WrappedSol, Source: "sol", InitialVirtualQuoteReserves: global.InitialVirtualSolReserves},
			TokenProgram:       solana.TokenProgramID,
		}, nil
	}
	listed, err := o.FetchSupportedQuoteMints(ctx)
	if err != nil {
		return nil, err
	}
	for _, entry := range listed {
		if entry.Mint.Equals(quoteMint) {
			program, err := o.FetchQuoteTokenProgram(ctx, quoteMint)
			if err != nil {
				return nil, err
			}
			return &ResolvedQuote{SupportedQuoteMint: entry, TokenProgram: program}, nil
		}
	}
	curve, accounts, err := o.FetchPumpQuote(ctx, quoteMint, global)
	if err != nil {
		return nil, err
	}
	if curve == nil {
		return nil, &UnsupportedQuoteMintError{QuoteMint: quoteMint}
	}
	program, err := o.FetchQuoteTokenProgram(ctx, quoteMint)
	if err != nil {
		return nil, err
	}
	seed, err := PumpQuoteReserves(global, curve, true, curve.QuoteControl)
	if err != nil {
		return nil, err
	}
	return &ResolvedQuote{
		SupportedQuoteMint: SupportedQuoteMint{Mint: quoteMint, Source: "pumpCoin", InitialVirtualQuoteReserves: seed.VirtualQuoteReserves},
		TokenProgram:       program,
		PumpQuote:          curve,
		Accounts:           accounts,
	}, nil
}

// FetchPumpQuote reads a pump coin offered as a quote, including its chain and migrated pools.
// A missing bonding curve returns nil, nil, nil. global may be nil to skip the admission checks.
func (o *Client) FetchPumpQuote(ctx context.Context, quoteMint solana.PublicKey, global *Global) (*PumpQuoteCurve, *PumpQuoteAccounts, error) {
	infos, err := o.RPC.GetMultipleAccounts(ctx, DeriveBondingCurve(quoteMint), DeriveQuoteControl())
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return nil, nil, err
	}
	if infos == nil || len(infos.Value) == 0 || infos.Value[0] == nil || !infos.Value[0].Owner.Equals(ProgramID) {
		return nil, nil, nil
	}
	curve, err := pump.ParseAccount_BondingCurve(infos.Value[0].Data.GetBinary())
	if err != nil {
		return nil, nil, err
	}
	var control *pump.QuoteControl
	if len(infos.Value) > 1 && infos.Value[1] != nil {
		control, err = pump.ParseAccount_QuoteControl(infos.Value[1].Data.GetBinary())
		if err != nil {
			return nil, nil, err
		}
	}
	quote := &PumpQuoteCurve{
		PumpQuoteChainCoin: PumpQuoteChainCoin{Mint: quoteMint, BondingCurve: curve},
		QuoteControlKnown:  true,
		QuoteControl:       control,
	}
	if global != nil {
		if _, err := PumpQuoteGate(global, quoteMint, curve); err != nil {
			return nil, nil, err
		}
	}
	coin := &quote.PumpQuoteChainCoin
	for i := 0; i < int(curve.Depth); i++ {
		mint := coin.BondingCurve.QuoteMint
		acc, err := o.RPC.GetAccountInfo(ctx, DeriveBondingCurve(mint))
		if err != nil || acc == nil || acc.Value == nil || !acc.Value.Owner.Equals(ProgramID) {
			return nil, nil, fmt.Errorf("quote chain of %s: no pump bonding curve for %s", quoteMint, mint)
		}
		nextCurve, err := pump.ParseAccount_BondingCurve(acc.Value.Data.GetBinary())
		if err != nil {
			return nil, nil, err
		}
		next := PumpQuoteChainCoin{Mint: mint, BondingCurve: nextCurve}
		quote.Chain = append(quote.Chain, next)
		coin = &quote.Chain[len(quote.Chain)-1]
	}
	if global != nil {
		if _, err := PumpQuoteAdmission(global, quote, true, control); err != nil {
			return nil, nil, err
		}
	}
	accounts := &PumpQuoteAccounts{}
	chain := append([]PumpQuoteChainCoin{quote.PumpQuoteChainCoin}, quote.Chain...)
	for i := range chain {
		if !IsBondingCurveMigrated(chain[i].BondingCurve) {
			continue
		}
		pool, poolAccounts, err := o.fetchQuotePool(ctx, chain[i].Mint, chain[i].BondingCurve)
		if err != nil {
			return nil, nil, err
		}
		if i == 0 {
			quote.Pool = pool
			accounts.Pool = poolAccounts
		} else {
			quote.Chain[i-1].Pool = pool
			accounts.Chain = append(accounts.Chain, PumpQuoteChainAccounts{Mint: chain[i].Mint, Pool: poolAccounts})
		}
	}
	return quote, accounts, nil
}

func (o *Client) fetchQuotePool(ctx context.Context, mint solana.PublicKey, curve *BondingCurve) (*PumpQuotePool, *PumpQuotePoolAccounts, error) {
	poolKey := DeriveCanonicalPoolPDAWithQuote(DerivePoolAuthority(mint), mint, NormalizeQuoteMint(curve.QuoteMint))
	acc, err := o.RPC.GetAccountInfo(ctx, poolKey)
	if err != nil || acc == nil || acc.Value == nil || !acc.Value.Owner.Equals(AmmProgramID) {
		return nil, nil, &QuotePoolNotFoundError{QuoteMint: mint, Pool: poolKey}
	}
	pool, err := ammgen.ParseAccount_Pool(acc.Value.Data.GetBinary())
	if err != nil {
		return nil, nil, fmt.Errorf("pool %s is not a pump-amm pool: %w", poolKey, err)
	}
	vaults, err := o.RPC.GetMultipleAccounts(ctx, pool.PoolBaseTokenAccount, pool.PoolQuoteTokenAccount)
	if err != nil {
		return nil, nil, err
	}
	base, err := initializedTokenAmount(pool.PoolBaseTokenAccount, vaults.Value[0])
	if err != nil {
		return nil, nil, err
	}
	quoteAmt, err := initializedTokenAmount(pool.PoolQuoteTokenAccount, vaults.Value[1])
	if err != nil {
		return nil, nil, err
	}
	return &PumpQuotePool{
		BaseReserves:         u64(base),
		QuoteReserves:        u64(quoteAmt),
		VirtualQuoteReserves: pool.VirtualQuoteReserves.BigInt(),
	}, &PumpQuotePoolAccounts{
		Pool:       poolKey,
		BaseVault:  pool.PoolBaseTokenAccount,
		QuoteVault: pool.PoolQuoteTokenAccount,
	}, nil
}

func initializedTokenAmount(address solana.PublicKey, info *rpc.Account) (uint64, error) {
	if info == nil || (!info.Owner.Equals(solana.TokenProgramID) && !info.Owner.Equals(solana.Token2022ProgramID)) {
		return 0, fmt.Errorf("pool vault %s is missing, uninitialized or frozen", address)
	}
	acc := &token.Account{}
	if err := acc.UnmarshalWithDecoder(bin.NewBinDecoder(info.Data.GetBinary())); err != nil {
		return 0, fmt.Errorf("pool vault %s is missing, uninitialized or frozen", address)
	}
	if acc.State != token.Initialized {
		return 0, fmt.Errorf("pool vault %s is missing, uninitialized or frozen", address)
	}
	return acc.Amount, nil
}

// BuyState is the on-chain state a buy quote and buy instruction need.
type BuyState struct {
	BondingCurve          *BondingCurve
	UserBaseExists        bool
	CurveBaseTokenBalance uint64
	QuoteMint             solana.PublicKey
	QuoteTokenProgram     solana.PublicKey
	QuoteTransferFee      *QuoteTransferFee
}

// FetchBuyState reads the curve, the user's base ATA and the curve's base balance.
// quoteMintHint may be the zero key when the caller does not know the quote yet.
func (o *Client) FetchBuyState(ctx context.Context, mint, user, tokenProgram, quoteMintHint solana.PublicKey) (*BuyState, error) {
	curveKey := DeriveBondingCurve(mint)
	userATA := helpers.FindAssociatedTokenAddress(user, mint, tokenProgram)
	keys := []solana.PublicKey{curveKey, userATA}
	hint := !quoteMintHint.Equals(solana.PublicKey{}) && !IsLegacyQuoteMint(quoteMintHint)
	if hint {
		keys = append(keys, quoteMintHint)
	}
	baseATAs := []solana.PublicKey{
		helpers.FindAssociatedTokenAddress(curveKey, mint, solana.TokenProgramID),
		helpers.FindAssociatedTokenAddress(curveKey, mint, solana.Token2022ProgramID),
	}
	keys = append(keys, baseATAs...)
	infos, err := o.RPC.GetMultipleAccounts(ctx, keys...)
	if err != nil {
		return nil, err
	}
	if infos.Value[0] == nil {
		return nil, fmt.Errorf("bonding curve account not found for mint: %s", mint)
	}
	curve, err := pump.ParseAccount_BondingCurve(infos.Value[0].Data.GetBinary())
	if err != nil {
		return nil, err
	}
	state := &BuyState{BondingCurve: curve, UserBaseExists: infos.Value[1] != nil}
	baseInfos := infos.Value[len(infos.Value)-2:]
	for _, info := range baseInfos {
		if info == nil {
			continue
		}
		amount, err := initializedTokenAmount(solana.PublicKey{}, info)
		if err != nil {
			continue
		}
		state.CurveBaseTokenBalance += amount
	}
	var hintInfo *rpc.Account
	if hint {
		hintInfo = infos.Value[2]
	}
	if err := o.fillQuote(ctx, state, curve, quoteMintHint, hintInfo); err != nil {
		return nil, err
	}
	return state, nil
}

// SellState is the on-chain state a sell quote needs.
type SellState struct {
	BondingCurve      *BondingCurve
	QuoteMint         solana.PublicKey
	QuoteTokenProgram solana.PublicKey
	QuoteTransferFee  *QuoteTransferFee
}

// FetchSellState reads the curve and the user's base ATA. The user ATA must exist.
func (o *Client) FetchSellState(ctx context.Context, mint, user, tokenProgram, quoteMintHint solana.PublicKey) (*SellState, error) {
	keys := []solana.PublicKey{DeriveBondingCurve(mint), helpers.FindAssociatedTokenAddress(user, mint, tokenProgram)}
	hint := !quoteMintHint.Equals(solana.PublicKey{}) && !IsLegacyQuoteMint(quoteMintHint)
	if hint {
		keys = append(keys, quoteMintHint)
	}
	infos, err := o.RPC.GetMultipleAccounts(ctx, keys...)
	if err != nil {
		return nil, err
	}
	if infos.Value[0] == nil {
		return nil, fmt.Errorf("bonding curve account not found for mint: %s", mint)
	}
	if infos.Value[1] == nil {
		return nil, fmt.Errorf("associated token account not found for mint: %s and user: %s", mint, user)
	}
	curve, err := pump.ParseAccount_BondingCurve(infos.Value[0].Data.GetBinary())
	if err != nil {
		return nil, err
	}
	buy := &BuyState{BondingCurve: curve}
	var hintInfo *rpc.Account
	if hint {
		hintInfo = infos.Value[2]
	}
	if err := o.fillQuote(ctx, buy, curve, quoteMintHint, hintInfo); err != nil {
		return nil, err
	}
	return &SellState{BondingCurve: curve, QuoteMint: buy.QuoteMint, QuoteTokenProgram: buy.QuoteTokenProgram, QuoteTransferFee: buy.QuoteTransferFee}, nil
}

func (o *Client) fillQuote(ctx context.Context, state *BuyState, curve *BondingCurve, hint solana.PublicKey, hintInfo *rpc.Account) error {
	state.QuoteMint = NormalizeQuoteMint(curve.QuoteMint)
	if IsLegacyQuoteMint(curve.QuoteMint) {
		state.QuoteTokenProgram = solana.TokenProgramID
		return nil
	}
	if hint.Equals(state.QuoteMint) && hintInfo != nil {
		state.QuoteTokenProgram = hintInfo.Owner
	} else {
		program, err := o.FetchQuoteTokenProgram(ctx, state.QuoteMint)
		if err != nil {
			return err
		}
		state.QuoteTokenProgram = program
	}
	fee, err := o.FetchQuoteTransferFee(ctx, state.QuoteMint)
	if err != nil {
		return err
	}
	state.QuoteTransferFee = fee
	return nil
}

// GetCreatorVaultBalance is the lamports in the creator vault above rent exemption.
func (o *Client) GetCreatorVaultBalance(ctx context.Context, creator solana.PublicKey) (uint64, error) {
	acc, err := o.RPC.GetAccountInfo(ctx, DeriveCreatorVault(creator))
	if errors.Is(err, rpc.ErrNotFound) || acc == nil || acc.Value == nil {
		if err != nil && !errors.Is(err, rpc.ErrNotFound) {
			return 0, err
		}
		return 0, nil
	}
	rent, err := o.RPC.GetMinimumBalanceForRentExemption(ctx, uint64(len(acc.Value.Data.GetBinary())), o.Commitment)
	if err != nil {
		return 0, err
	}
	if acc.Value.Lamports < rent {
		return 0, nil
	}
	return acc.Value.Lamports - rent, nil
}

// GetCreatorVaultBalanceBothPrograms adds the AMM coin-creator WSOL vault to the pump creator vault.
func (o *Client) GetCreatorVaultBalanceBothPrograms(ctx context.Context, creator solana.PublicKey) (uint64, error) {
	pumpBalance, err := o.GetCreatorVaultBalance(ctx, creator)
	if err != nil {
		return 0, err
	}
	ammBalance, err := ammclient.NewClient(o.RPC, o.Commitment).GetCoinCreatorVaultBalance(ctx, creator, solana.WrappedSol, solana.TokenProgramID)
	if err != nil {
		return 0, err
	}
	return pumpBalance + ammBalance, nil
}

// GetTotalUnclaimedTokensBothPrograms sums the pump and AMM incentive programs.
func (o *Client) GetTotalUnclaimedTokensBothPrograms(ctx context.Context, user solana.PublicKey) (uint64, error) {
	pumpTokens, err := o.GetTotalUnclaimedTokens(ctx, user)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return 0, err
	}
	ammTokens, err := ammclient.NewClient(o.RPC, o.Commitment).GetTotalUnclaimedTokens(ctx, user)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return 0, err
	}
	return pumpTokens + ammTokens, nil
}

// GetCurrentDayTokensBothPrograms sums the current-day incentive of both programs.
func (o *Client) GetCurrentDayTokensBothPrograms(ctx context.Context, user solana.PublicKey) (uint64, error) {
	pumpTokens, err := o.GetCurrentDayTokens(ctx, user)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return 0, err
	}
	ammTokens, err := ammclient.NewClient(o.RPC, o.Commitment).GetCurrentDayTokens(ctx, user)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return 0, err
	}
	return pumpTokens + ammTokens, nil
}

// AdminUpdateTokenIncentivesInstruction updates the pump program's incentive schedule.
// authority is read from Global when the zero key is passed.
func (o *Client) AdminUpdateTokenIncentivesInstruction(ctx context.Context, authority solana.PublicKey, startTime, endTime, secondsInADay int64, dayNumber, tokenSupplyPerDay uint64, mint, tokenProgram solana.PublicKey) (solana.Instruction, error) {
	if authority.Equals(solana.PublicKey{}) {
		global, err := o.FetchGlobal(ctx)
		if err != nil {
			return nil, err
		}
		authority = global.Authority
	}
	if mint.Equals(solana.PublicKey{}) {
		mint = PumpTokenMint
	}
	if tokenProgram.Equals(solana.PublicKey{}) {
		tokenProgram = solana.Token2022ProgramID
	}
	if secondsInADay == 0 {
		secondsInADay = 86400
	}
	return pump.NewAdminUpdateTokenIncentivesInstruction(
		startTime, endTime, secondsInADay, dayNumber, tokenSupplyPerDay,
		authority, o.Global, DeriveGlobalVolumeAccumulator(), mint,
		helpers.FindAssociatedTokenAddress(DeriveGlobalVolumeAccumulator(), mint, tokenProgram),
		solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID, tokenProgram,
		o.EventAuthority, ProgramID,
	)
}

// CollectCreatorFeeV2Instructions collects the AMM vault when it exists, then the pump v2 vault.
func (o *Client) CollectCreatorFeeV2Instructions(ctx context.Context, coinCreator, quoteMint, quoteTokenProgram, feePayer solana.PublicKey) ([]solana.Instruction, error) {
	quote := NormalizeQuoteMint(quoteMint)
	if feePayer.Equals(solana.PublicKey{}) {
		feePayer = coinCreator
	}
	ammIx, err := ammclient.NewClient(o.RPC, o.Commitment).CollectCoinCreatorFeeInstructions(ctx, coinCreator, quote, quoteTokenProgram, feePayer)
	if err != nil {
		return nil, err
	}
	pumpIx, err := o.CollectCreatorFeeV2Instruction(coinCreator, quote, quoteTokenProgram)
	if err != nil {
		return nil, err
	}
	if quote.Equals(solana.WrappedSol) {
		var creates, rest []solana.Instruction
		for _, ix := range ammIx {
			if ix.ProgramID().Equals(solana.SPLAssociatedTokenAccountProgramID) {
				creates = append(creates, ix)
			} else {
				rest = append(rest, ix)
			}
		}
		out := append(creates, pumpIx)
		return append(out, rest...), nil
	}
	return append(ammIx, pumpIx), nil
}
