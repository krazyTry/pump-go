package bonding_curve

import (
	"context"

	solana "github.com/gagliardetto/solana-go"
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	pump "github.com/krazyTry/pump-go/gen/pump"
)

// Deprecated: Use CreateAndBuyV2Instructions.
func (c *Client) CreateAndBuyInstructions(ctx context.Context, global *pump.Global, baseMint solana.PublicKey, name, symbol, uri string, creator, user solana.PublicKey, amount, solAmount uint64) ([]solana.Instruction, error) {
	createIx, err := c.CreateInstruction(creator, user, baseMint, name, symbol, uri)
	if err != nil {
		return nil, err
	}
	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	outputTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMint, user, user, solana.TokenProgramID)
	if err != nil {
		return nil, err
	}

	inputTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, solana.WrappedSol, user, user, solana.TokenProgramID)
	if err != nil {
		return nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, inputTokenAccount, solAmount)
	if err != nil {
		return nil, err
	}
	pre = append(pre, wrapIx...)

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	buyIx, err := c.GetBuyInstruction(creator, user, baseMint, solana.TokenProgramID, outputTokenAccount, amount, solAmount, helpers.GetFeeRecipient(global, false))
	if err != nil {
		return nil, err
	}

	return func() []solana.Instruction {
		ix := []solana.Instruction{createIx}
		ix = append(ix, pre...)
		ix = append(ix, buyIx)
		ix = append(ix, post...)
		return ix
	}(), nil
}

func (c *Client) CreateAndBuyV2Instructions(ctx context.Context, global *pump.Global, creator, user, baseMint solana.PublicKey, name, symbol, uri string, amount, solAmount uint64) ([]solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	baseTokenProgram := solana.TokenProgramID
	quoteMint := solana.WrappedSol
	quoteTokenProgram := solana.TokenProgramID

	createIx, err := c.CreateInstruction(creator, user, baseMint, name, symbol, uri)
	if err != nil {
		return nil, nil, nil, err
	}
	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	baseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	quoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, quoteTokenAccount, solAmount)
	if err != nil {
		return nil, nil, nil, err
	}
	pre = append(pre, wrapIx...)

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	buyIx, err := c.GetBuyV2Instruction(
		creator,
		user,
		baseMint,
		quoteMint,
		baseTokenProgram,
		quoteTokenProgram,
		baseTokenAccount,
		quoteTokenAccount,
		amount,
		solAmount,
		helpers.GetStaticRandomFeeRecipient(), // helpers.GetFeeRecipient(global, false),
		helpers.GetStaticRandomFeeRecipientForBuyback(),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	return []solana.Instruction{createIx, buyIx}, pre, post, nil
}

// Deprecated: Use CreateV2AndBuyV2Instructions.
func (c *Client) CreateV2AndBuyInstructions(ctx context.Context, global *pump.Global, baseMint solana.PublicKey, name, symbol, uri string, creator, user solana.PublicKey, amount, solAmount uint64, mayhemMode bool, cashback bool) ([]solana.Instruction, error) {
	createIx, err := c.CreateV2Instruction(creator, user, baseMint, solana.WrappedSol, solana.TokenProgramID, name, symbol, uri, mayhemMode, cashback)
	if err != nil {
		return nil, err
	}

	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	outputTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMint, user, user, solana.Token2022ProgramID)
	if err != nil {
		return nil, err
	}

	inputTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, solana.WrappedSol, user, user, solana.TokenProgramID)
	if err != nil {
		return nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, inputTokenAccount, solAmount)
	if err != nil {
		return nil, err
	}
	pre = append(pre, wrapIx...)

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	buyIx, err := c.GetBuyInstruction(creator, user, baseMint, solana.Token2022ProgramID, outputTokenAccount, amount, solAmount, helpers.GetFeeRecipient(global, mayhemMode))
	if err != nil {
		return nil, err
	}

	return func() []solana.Instruction {
		ix := []solana.Instruction{createIx}
		ix = append(ix, pre...)
		ix = append(ix, buyIx)
		ix = append(ix, post...)
		return ix
	}(), nil
}

func (c *Client) CreateV2AndBuyV2Instructions(ctx context.Context, global *pump.Global, creator, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram solana.PublicKey, name, symbol, uri string, amount, quoteAmount uint64, mayhemMode bool, cashback bool) ([]solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	if quoteMint.Equals(solana.PublicKey{}) {
		quoteMint = solana.WrappedSol
	}
	if quoteTokenProgram.Equals(solana.PublicKey{}) {
		quoteTokenProgram = solana.TokenProgramID
	}

	createIx, err := c.CreateV2Instruction(creator, user, baseMint, quoteMint, quoteTokenProgram, name, symbol, uri, mayhemMode, cashback)
	if err != nil {
		return nil, nil, nil, err
	}

	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	baseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	quoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, quoteTokenAccount, quoteAmount)
	if err != nil {
		return nil, nil, nil, err
	}
	pre = append(pre, wrapIx...)

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	buyIx, err := c.GetBuyV2Instruction(
		creator,
		user,
		baseMint,
		quoteMint,
		baseTokenProgram,
		quoteTokenProgram,
		baseTokenAccount,
		quoteTokenAccount,
		amount,
		quoteAmount,
		helpers.GetFeeRecipient(global, mayhemMode),
		helpers.GetStaticRandomFeeRecipientForBuyback(),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	return []solana.Instruction{createIx, buyIx}, pre, post, nil
}

// Deprecated: Use BuyV2Instructions.
func (c *Client) BuyInstructions(ctx context.Context, global *pump.Global, bondingCurve *pump.BondingCurve, user, baseMint, baseTokenProgram solana.PublicKey, amount, solAmount uint64, slippage float64) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	maxSolCost := solAmount + (solAmount*uint64(slippage*10))/1000

	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	outputTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMint, user, user, solana.TokenProgramID)
	if err != nil {
		return nil, nil, nil, err
	}

	inputTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, solana.WrappedSol, user, user, solana.TokenProgramID)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, inputTokenAccount, maxSolCost)
	if err != nil {
		return nil, nil, nil, err
	}
	pre = append(pre, wrapIx...)

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	ix, err := c.GetBuyInstruction(bondingCurve.Creator, user, baseMint, baseTokenProgram, outputTokenAccount, amount, maxSolCost, helpers.GetFeeRecipient(global, bondingCurve.IsMayhemMode))
	if err != nil {
		return nil, nil, nil, err
	}

	return ix, pre, post, nil
}

func (c *Client) BuyV2Instructions(ctx context.Context, global *pump.Global, bondingCurve *pump.BondingCurve, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram solana.PublicKey, amount, quoteAmount uint64, slippage float64) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {

	if quoteMint.Equals(solana.PublicKey{}) || quoteMint.Equals(solana.WrappedSol) {
		quoteMint = solana.WrappedSol
		quoteTokenProgram = solana.TokenProgramID
	}

	maxQuote := quoteAmount + (quoteAmount*uint64(slippage*10))/1000

	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	baseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	quoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, quoteTokenAccount, maxQuote)
	if err != nil {
		return nil, nil, nil, err
	}
	pre = append(pre, wrapIx...)

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	ix, err := c.GetBuyV2Instruction(
		bondingCurve.Creator,
		user,
		baseMint,
		quoteMint,
		baseTokenProgram,
		quoteTokenProgram,
		baseTokenAccount,
		quoteTokenAccount,
		amount,
		maxQuote,
		helpers.GetFeeRecipient(global, bondingCurve.IsMayhemMode),
		helpers.GetStaticRandomFeeRecipientForBuyback(),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	return ix, pre, post, nil
}

// Deprecated: Use SellV2Instructions.
func (c *Client) SellInstructions(ctx context.Context, global *pump.Global, bondingCurve *pump.BondingCurve, user, baseMint, baseTokenProgram solana.PublicKey, amount, solAmount uint64, slippage float64) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	minSolOutput := solAmount - (solAmount*uint64(slippage*10))/1000

	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	inputTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	_, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, solana.WrappedSol, user, user, solana.TokenProgramID)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	ix, err := c.GetSellInstruction(bondingCurve.Creator, user, baseMint, baseTokenProgram, inputTokenAccount, amount, minSolOutput, helpers.GetFeeRecipient(global, bondingCurve.IsMayhemMode))
	if err != nil {
		return nil, nil, nil, err
	}

	return ix, pre, post, nil
}

func (c *Client) SellV2Instructions(ctx context.Context, global *pump.Global, bondingCurve *pump.BondingCurve, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram solana.PublicKey, amount, quoteAmount uint64, slippage float64) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {

	if quoteMint.Equals(solana.PublicKey{}) || quoteMint.Equals(solana.WrappedSol) {
		quoteMint = solana.WrappedSol
		quoteTokenProgram = solana.TokenProgramID
	}

	minQuote := quoteAmount - (quoteAmount*uint64(slippage*10))/1000

	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	baseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	quoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, c.RPC, solana.WrappedSol, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	ix, err := c.GetSellV2Instruction(
		bondingCurve.Creator,
		user,
		baseMint,
		quoteMint,
		baseTokenProgram,
		quoteTokenProgram,
		baseTokenAccount,
		quoteTokenAccount,
		amount,
		minQuote,
		helpers.GetFeeRecipient(global, bondingCurve.IsMayhemMode),
		helpers.GetStaticRandomFeeRecipientForBuyback(),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	return ix, pre, post, nil
}
