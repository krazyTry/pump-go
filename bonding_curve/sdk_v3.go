package bonding_curve

import (
	"context"

	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	pump "github.com/krazyTry/pump-go/gen/pump"
	solana "github.com/solana-foundation/solana-go/v2"
)

func resolveV3Buyback(quoteMint, quoteTokenProgram, buybackWallet solana.PublicKey) solana.PublicKey {
	quote := helpers.NormalizeQuoteMint(quoteMint)
	wallet := buybackWallet
	if wallet.Equals(solana.PublicKey{}) {
		if helpers.IsLegacyQuoteMint(quoteMint) {
			return helpers.GetStaticRandomBuybackFeeRecipient()
		}
		wallet = helpers.GetBuybackFeeRecipientForTokenQuote()
	}
	if helpers.IsLegacyQuoteMint(quoteMint) {
		return wallet
	}
	return helpers.FindAssociatedTokenAddress(wallet, quote, quoteTokenProgram)
}

// GetBuyV3Instruction builds buy_v3. buybackFeeRecipient is the wallet; a zero
// key selects the default. On a token quote the instruction is given that wallet's quote ATA.
func (c *Client) GetBuyV3Instruction(partialFill bool, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount solana.PublicKey, amount, maxSolCost uint64, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
	quote := helpers.NormalizeQuoteMint(quoteMint)
	bondingCurve := DeriveBondingCurve(baseMint)
	return pump.NewBuyV3Instruction(
		amount,
		maxSolCost,
		pump.OptionBool{V0: partialFill},
		c.Global,
		baseMint,
		quote,
		baseTokenProgram,
		quoteTokenProgram,
		bondingCurve,
		helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, baseTokenProgram),
		helpers.FindAssociatedTokenAddress(bondingCurve, quote, quoteTokenProgram),
		user,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		DeriveUserVolumeAccumulator(user),
		c.FeeConfig,
		resolveV3Buyback(quoteMint, quoteTokenProgram, buybackFeeRecipient),
		solana.SystemProgramID,
		c.EventAuthority,
		ProgramID,
	)
}

// GetBuyExactQuoteInV3Instruction builds buy_exact_quote_in_v3.
func (c *Client) GetBuyExactQuoteInV3Instruction(partialFill bool, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount solana.PublicKey, spendableQuoteIn, minTokensOut uint64, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
	quote := helpers.NormalizeQuoteMint(quoteMint)
	bondingCurve := DeriveBondingCurve(baseMint)
	return pump.NewBuyExactQuoteInV3Instruction(
		spendableQuoteIn,
		minTokensOut,
		pump.OptionBool{V0: partialFill},
		c.Global,
		baseMint,
		quote,
		baseTokenProgram,
		quoteTokenProgram,
		bondingCurve,
		helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, baseTokenProgram),
		helpers.FindAssociatedTokenAddress(bondingCurve, quote, quoteTokenProgram),
		user,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		DeriveUserVolumeAccumulator(user),
		c.FeeConfig,
		resolveV3Buyback(quoteMint, quoteTokenProgram, buybackFeeRecipient),
		solana.SystemProgramID,
		c.EventAuthority,
		ProgramID,
	)
}

// GetSellV3Instruction builds sell_v3.
func (c *Client) GetSellV3Instruction(user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount solana.PublicKey, amount, minSolOutput uint64, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
	quote := helpers.NormalizeQuoteMint(quoteMint)
	bondingCurve := DeriveBondingCurve(baseMint)
	return pump.NewSellV3Instruction(
		amount,
		minSolOutput,
		c.Global,
		baseMint,
		quote,
		baseTokenProgram,
		quoteTokenProgram,
		bondingCurve,
		helpers.FindAssociatedTokenAddress(bondingCurve, baseMint, baseTokenProgram),
		helpers.FindAssociatedTokenAddress(bondingCurve, quote, quoteTokenProgram),
		user,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		DeriveUserVolumeAccumulator(user),
		c.FeeConfig,
		resolveV3Buyback(quoteMint, quoteTokenProgram, buybackFeeRecipient),
		solana.SystemProgramID,
		c.EventAuthority,
		ProgramID,
	)
}

func (c *Client) prepareV3Trade(ctx context.Context, curve *pump.BondingCurve, baseMint solana.PublicKey, user, baseMintArg, quoteMint, baseTokenProgram, quoteTokenProgram solana.PublicKey, legacyQuote bool) (solana.PublicKey, solana.PublicKey, solana.PublicKey, []solana.Instruction, []solana.Instruction, error) {
	if curve != nil && curve.IsCashbackCoin {
		return solana.PublicKey{}, solana.PublicKey{}, solana.PublicKey{}, nil, nil, &CashbackCoinNotSupportedError{Mint: baseMint}
	}
	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)
	baseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMintArg, user, user, baseTokenProgram)
	if err != nil {
		return solana.PublicKey{}, solana.PublicKey{}, solana.PublicKey{}, nil, nil, err
	}
	quote := helpers.NormalizeQuoteMint(quoteMint)
	quoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, quote, user, user, quoteTokenProgram)
	if err != nil {
		return solana.PublicKey{}, solana.PublicKey{}, solana.PublicKey{}, nil, nil, err
	}
	if ixA != nil {
		pre = append(pre, ixA)
	}
	if ixB != nil {
		pre = append(pre, ixB)
	}
	if !legacyQuote {
		wallet := helpers.GetBuybackFeeRecipientForTokenQuote()
		buybackATA, ixC, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, quote, wallet, user, quoteTokenProgram)
		if err != nil {
			return solana.PublicKey{}, solana.PublicKey{}, solana.PublicKey{}, nil, nil, err
		}
		if ixC != nil {
			pre = append(pre, ixC)
		}
		_ = buybackATA
	}
	return baseTokenAccount, quoteTokenAccount, quote, pre, post, nil
}

// BuyV3Instructions builds buy_v3 plus the token-account setup around it.
// quoteAmount is the fee-inclusive cost from BuyV3QuoteAmountFromTokenAmount; slippage widens the cap.
func (c *Client) BuyV3Instructions(ctx context.Context, curve *pump.BondingCurve, partialFill bool, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram solana.PublicKey, amount, quoteAmount uint64, slippage float64) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	legacyQuote := helpers.IsLegacyQuoteMint(quoteMint)
	if legacyQuote {
		quoteMint = solana.WrappedSol
		quoteTokenProgram = solana.TokenProgramID
	}
	maxQuote := quoteAmount + (quoteAmount*uint64(slippage*10))/1000
	baseAccount, quoteAccount, _, pre, post, err := c.prepareV3Trade(ctx, curve, baseMint, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, legacyQuote)
	if err != nil {
		return nil, nil, nil, err
	}
	if legacyQuote {
		wrapIx, err := helpers.WrapSOLInstruction(user, quoteAccount, maxQuote)
		if err != nil {
			return nil, nil, nil, err
		}
		pre = append(pre, wrapIx...)
		unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
		if uerr == nil && unwrapIx != nil {
			post = append(post, unwrapIx)
		}
	}
	ix, err := c.GetBuyV3Instruction(partialFill, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, baseAccount, quoteAccount, amount, maxQuote, solana.PublicKey{})
	if err != nil {
		return nil, nil, nil, err
	}
	return ix, pre, post, nil
}

// BuyExactQuoteInV3Instructions spends at most quoteAmount and requires amount tokens less slippage, at least 1.
func (c *Client) BuyExactQuoteInV3Instructions(ctx context.Context, curve *pump.BondingCurve, partialFill bool, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram solana.PublicKey, amount, quoteAmount uint64, slippage float64) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	legacyQuote := helpers.IsLegacyQuoteMint(quoteMint)
	if legacyQuote {
		quoteMint = solana.WrappedSol
		quoteTokenProgram = solana.TokenProgramID
	}
	minTokens := amount - (amount*uint64(slippage*10))/1000
	if minTokens == 0 {
		minTokens = 1
	}
	baseAccount, quoteAccount, _, pre, post, err := c.prepareV3Trade(ctx, curve, baseMint, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, legacyQuote)
	if err != nil {
		return nil, nil, nil, err
	}
	if legacyQuote {
		wrapIx, err := helpers.WrapSOLInstruction(user, quoteAccount, quoteAmount)
		if err != nil {
			return nil, nil, nil, err
		}
		pre = append(pre, wrapIx...)
		unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
		if uerr == nil && unwrapIx != nil {
			post = append(post, unwrapIx)
		}
	}
	ix, err := c.GetBuyExactQuoteInV3Instruction(partialFill, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, baseAccount, quoteAccount, quoteAmount, minTokens, solana.PublicKey{})
	if err != nil {
		return nil, nil, nil, err
	}
	return ix, pre, post, nil
}

// SellV3Instructions sells amount base for at least quoteAmount less slippage.
func (c *Client) SellV3Instructions(ctx context.Context, curve *pump.BondingCurve, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram solana.PublicKey, amount, quoteAmount uint64, slippage float64) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	legacyQuote := helpers.IsLegacyQuoteMint(quoteMint)
	if legacyQuote {
		quoteMint = solana.WrappedSol
		quoteTokenProgram = solana.TokenProgramID
	}
	minQuote := quoteAmount - (quoteAmount*uint64(slippage*10))/1000
	baseAccount, quoteAccount, _, pre, post, err := c.prepareV3Trade(ctx, curve, baseMint, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, legacyQuote)
	if err != nil {
		return nil, nil, nil, err
	}
	if legacyQuote {
		unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
		if uerr == nil && unwrapIx != nil {
			post = append(post, unwrapIx)
		}
	}
	ix, err := c.GetSellV3Instruction(user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, baseAccount, quoteAccount, amount, minQuote, solana.PublicKey{})
	if err != nil {
		return nil, nil, nil, err
	}
	return ix, pre, post, nil
}
