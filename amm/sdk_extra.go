package amm

import (
	"context"
	"errors"

	solana "github.com/gagliardetto/solana-go"
	"github.com/krazyTry/pump-go/amm/helpers"
	"github.com/shopspring/decimal"
)

func (s *Client) DepositToken0(token0 uint64, slippage float64, pool *Pool) (*DepositResult, error) {
	token0Decimal := decimal.NewFromUint64(token0)
	totalLpTokensDecimal := decimal.NewFromUint64(pool.LpSupply)

	poolBaseTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolBaseTokenAccount)
	if err != nil {
		return nil, err
	}
	baseReserve := decimal.NewFromUint64(poolBaseTokenAccount.Amount)

	poolQuoteTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolQuoteTokenAccount)
	if err != nil {
		return nil, err
	}
	quoteReserve := decimal.NewFromUint64(poolQuoteTokenAccount.Amount)

	if baseReserve.Sign() == 0 || quoteReserve.Sign() == 0 {
		return nil, errors.New("invalid input: baseReserve or quoteReserve cannot be zero")
	}

	if token0Decimal.Sign() == 0 || totalLpTokensDecimal.Sign() == 0 {
		return nil, errors.New("lp amount or total lp tokens cannot be zero")
	}

	token1 := token0Decimal.Mul(quoteReserve).Div(baseReserve)
	f := helpers.SlippageFactor(slippage, true)

	max0 := token0Decimal.Mul(f).Div(helpers.SlippageScale)

	max1 := token1.Mul(f).Div(helpers.SlippageScale)

	lp := token0Decimal.Mul(totalLpTokensDecimal).Div(baseReserve)

	return &DepositResult{
		Token1:    token1.BigInt().Uint64(),
		LpToken:   lp.BigInt().Uint64(),
		MaxToken0: max0.BigInt().Uint64(),
		MaxToken1: max1.BigInt().Uint64(),
	}, nil
}

func (s *Client) DepositLpToken(lpToken uint64, slippage float64, pool *Pool) (*DepositLpTokenResult, error) {
	lpTokenDecimal := decimal.NewFromUint64(lpToken)
	totalLpTokensDecimal := decimal.NewFromUint64(pool.LpSupply)

	poolBaseTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolBaseTokenAccount)
	if err != nil {
		return nil, err
	}
	baseReserve := decimal.NewFromUint64(poolBaseTokenAccount.Amount)

	poolQuoteTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolQuoteTokenAccount)
	if err != nil {
		return nil, err
	}
	quoteReserve := decimal.NewFromUint64(poolQuoteTokenAccount.Amount)

	if baseReserve.Sign() == 0 || quoteReserve.Sign() == 0 {
		return nil, errors.New("invalid input: baseReserve or quoteReserve cannot be zero")
	}

	if lpTokenDecimal.Sign() == 0 || totalLpTokensDecimal.Sign() == 0 {
		return nil, errors.New("lp amount or total lp tokens cannot be zero")
	}

	baseIn, _ := helpers.CeilDiv(baseReserve.Mul(lpTokenDecimal), totalLpTokensDecimal)
	quoteIn, _ := helpers.CeilDiv(quoteReserve.Mul(lpTokenDecimal), totalLpTokensDecimal)
	f := helpers.SlippageFactor(slippage, true)

	maxBase := baseIn.Mul(f).Div(helpers.SlippageScale)

	maxQuote := quoteIn.Mul(f).Div(helpers.SlippageScale)

	return &DepositLpTokenResult{
		MaxBase:  maxBase.BigInt().Uint64(),
		MaxQuote: maxQuote.BigInt().Uint64(),
	}, nil
}

func (s *Client) Withdraw(lpAmount uint64, slippage float64, pool *Pool) (*WithdrawResult, error) {
	lpAmountDecimal := decimal.NewFromUint64(lpAmount)
	totalLpTokensDecimal := decimal.NewFromUint64(pool.LpSupply)

	poolBaseTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolBaseTokenAccount)
	if err != nil {
		return nil, err
	}
	baseReserve := decimal.NewFromUint64(poolBaseTokenAccount.Amount)

	poolQuoteTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolQuoteTokenAccount)
	if err != nil {
		return nil, err
	}
	quoteReserve := decimal.NewFromUint64(poolQuoteTokenAccount.Amount)

	if baseReserve.Sign() == 0 || quoteReserve.Sign() == 0 {
		return nil, errors.New("invalid input: baseReserve or quoteReserve cannot be zero")
	}

	if lpAmountDecimal.Sign() == 0 || totalLpTokensDecimal.Sign() == 0 {
		return nil, errors.New("lp amount or total lp tokens cannot be zero")
	}

	base := baseReserve.Mul(lpAmountDecimal).Div(totalLpTokensDecimal)

	quote := quoteReserve.Mul(lpAmountDecimal).Div(totalLpTokensDecimal)

	f := helpers.SlippageFactor(slippage, false)

	minBase := base.Mul(f).Div(helpers.SlippageScale)
	minQuote := quote.Mul(f).Div(helpers.SlippageScale)

	return &WithdrawResult{
		Base:     base.BigInt().Uint64(),
		Quote:    quote.BigInt().Uint64(),
		MinBase:  minBase.BigInt().Uint64(),
		MinQuote: minQuote.BigInt().Uint64(),
	}, nil
}

func (s *Client) BuyBaseInput(base uint64, slippage float64, pool *Pool, globalConfig *GlobalConfig, feeConfig *FeeConfig) (*BuyBaseInputResult, error) {
	baseDecimal := decimal.NewFromUint64(base)

	baseMintAccount, err := helpers.GetTokenInfo(context.TODO(), s.RPC, pool.BaseMint)
	if err != nil {
		return nil, err
	}
	baseMintSupplyDecimal := decimal.NewFromUint64(baseMintAccount.Mint.Supply)

	poolBaseTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolBaseTokenAccount)
	if err != nil {
		return nil, err
	}
	baseReserve := decimal.NewFromUint64(poolBaseTokenAccount.Amount)

	poolQuoteTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolQuoteTokenAccount)
	if err != nil {
		return nil, err
	}
	quoteReserve := decimal.NewFromUint64(poolQuoteTokenAccount.Amount)

	if baseReserve.Sign() == 0 || quoteReserve.Sign() == 0 {
		return nil, errors.New("invalid input: baseReserve or quoteReserve cannot be zero")
	}

	if baseDecimal.Cmp(baseReserve) > 0 {
		return nil, errors.New("cannot buy more base tokens than the pool reserves")
	}

	n := quoteReserve.Mul(baseDecimal)
	d := baseReserve.Sub(baseDecimal)

	if d.Sign() == 0 {
		return nil, errors.New("pool would be depleted; denominator is zero")
	}

	quoteIn, _ := helpers.CeilDiv(n, d)

	fees, err := ComputeFeesBps(globalConfig, feeConfig, pool.Creator, baseMintSupplyDecimal, pool.BaseMint, baseReserve, quoteReserve, quoteIn)
	if err != nil {
		return nil, err
	}
	lpFee, _ := helpers.Fee(quoteIn, decimal.NewFromUint64(fees.LpFeeBps))
	protocolFee, _ := helpers.Fee(quoteIn, decimal.NewFromUint64(fees.ProtocolFeeBps))
	creatorBps := decimal.NewFromUint64(fees.CreatorFeeBps)

	if pool.CoinCreator.Equals(solana.PublicKey{}) {
		creatorBps = decimal.Zero
	}

	creatorFee, _ := helpers.Fee(quoteIn, creatorBps)

	total := quoteIn.Add(lpFee).Add(protocolFee).Add(creatorFee)

	f := helpers.SlippageFactor(slippage, true)
	maxQuote := total.Mul(f).Div(helpers.SlippageScale)

	return &BuyBaseInputResult{
		InternalQuoteAmount: quoteIn.BigInt().Uint64(),
		UIQuote:             total.BigInt().Uint64(),
		MaxQuote:            maxQuote.BigInt().Uint64(),
	}, nil
}

func (s *Client) BuyQuoteInput(quote uint64, slippage float64, pool *Pool, globalConfig *GlobalConfig, feeConfig *FeeConfig) (*BuyQuoteInputResult, error) {
	quoteDecimal := decimal.NewFromUint64(quote)

	baseMintAccount, err := helpers.GetTokenInfo(context.TODO(), s.RPC, pool.BaseMint)
	if err != nil {
		return nil, err
	}
	baseMintSupplyDecimal := decimal.NewFromUint64(baseMintAccount.Mint.Supply)

	poolBaseTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolBaseTokenAccount)
	if err != nil {
		return nil, err
	}
	baseReserve := decimal.NewFromUint64(poolBaseTokenAccount.Amount)

	poolQuoteTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolQuoteTokenAccount)
	if err != nil {
		return nil, err
	}
	quoteReserve := decimal.NewFromUint64(poolQuoteTokenAccount.Amount)

	if baseReserve.Sign() == 0 || quoteReserve.Sign() == 0 {
		return nil, errors.New("invalid input: baseReserve or quoteReserve cannot be zero")
	}

	fees, err := ComputeFeesBps(globalConfig, feeConfig, pool.Creator, baseMintSupplyDecimal, pool.BaseMint, baseReserve, quoteReserve, quoteDecimal)
	if err != nil {
		return nil, err
	}
	totalFeeBpsDecimal := decimal.NewFromUint64(fees.LpFeeBps).Add(decimal.NewFromUint64(fees.ProtocolFeeBps))

	if !pool.CoinCreator.Equals(solana.PublicKey{}) {
		totalFeeBpsDecimal = totalFeeBpsDecimal.Add(decimal.NewFromUint64(fees.CreatorFeeBps))
	}

	den := totalFeeBpsDecimal.Add(helpers.BpsDenominator)

	effective := quoteDecimal.Mul(helpers.BpsDenominator)

	effective = effective.Div(den)

	n := baseReserve.Mul(effective)
	d := quoteReserve.Add(effective)
	if d.Sign() == 0 {
		return nil, errors.New("pool would be depleted; denominator is zero")
	}
	baseOut := n.Div(d)
	f := helpers.SlippageFactor(slippage, true)

	maxQuote := quoteDecimal.Mul(f).Div(helpers.SlippageScale)

	return &BuyQuoteInputResult{
		Base:                     baseOut.BigInt().Uint64(),
		InternalQuoteWithoutFees: effective.BigInt().Uint64(),
		MaxQuote:                 maxQuote.BigInt().Uint64(),
	}, nil
}

func (s *Client) SellBaseInput(base uint64, slippage float64, pool *Pool, globalConfig *GlobalConfig, feeConfig *FeeConfig) (*SellBaseInputResult, error) {
	baseDecimal := decimal.NewFromUint64(base)
	baseMintAccount, err := helpers.GetTokenInfo(context.TODO(), s.RPC, pool.BaseMint)
	if err != nil {
		return nil, err
	}
	baseMintSupplyDecimal := decimal.NewFromUint64(baseMintAccount.Mint.Supply)

	poolBaseTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolBaseTokenAccount)
	if err != nil {
		return nil, err
	}
	baseReserve := decimal.NewFromUint64(poolBaseTokenAccount.Amount)

	poolQuoteTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolQuoteTokenAccount)
	if err != nil {
		return nil, err
	}
	quoteReserve := decimal.NewFromUint64(poolQuoteTokenAccount.Amount)

	if baseReserve.Sign() == 0 || quoteReserve.Sign() == 0 {
		return nil, errors.New("invalid input: baseReserve or quoteReserve cannot be zero")
	}

	quoteOut := quoteReserve.Mul(baseDecimal).Div(baseReserve.Add(baseDecimal))

	fees, err := ComputeFeesBps(globalConfig, feeConfig, pool.Creator, baseMintSupplyDecimal, pool.BaseMint, baseReserve, quoteReserve, quoteOut)
	if err != nil {
		return nil, err
	}
	lpFee, _ := helpers.Fee(quoteOut, decimal.NewFromUint64(fees.LpFeeBps))
	protocolFee, _ := helpers.Fee(quoteOut, decimal.NewFromUint64(fees.ProtocolFeeBps))
	creatorBps := decimal.NewFromUint64(fees.CreatorFeeBps)
	if pool.CoinCreator.Equals(solana.PublicKey{}) {
		creatorBps = decimal.Zero
	}
	creatorFee, _ := helpers.Fee(quoteOut, creatorBps)
	final := quoteOut.Sub(lpFee).Sub(protocolFee).Sub(creatorFee)
	if final.Sign() < 0 {
		return nil, errors.New("fees exceed total output; final quote is negative")
	}
	f := helpers.SlippageFactor(slippage, false)
	minQuote := final.Mul(f).Div(helpers.SlippageScale)
	return &SellBaseInputResult{
		UIQuote:                final.BigInt().Uint64(),
		MinQuote:               minQuote.BigInt().Uint64(),
		InternalQuoteAmountOut: quoteOut.BigInt().Uint64(),
	}, nil
}

func (s *Client) SellQuoteInput(quote uint64, slippage float64, pool *Pool, globalConfig *GlobalConfig, feeConfig *FeeConfig) (*SellQuoteInputResult, error) {
	quoteDecimal := decimal.NewFromUint64(quote)

	baseMintAccount, err := helpers.GetTokenInfo(context.TODO(), s.RPC, pool.BaseMint)
	if err != nil {
		return nil, err
	}
	baseMintSupplyDecimal := decimal.NewFromUint64(baseMintAccount.Mint.Supply)

	poolBaseTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolBaseTokenAccount)
	if err != nil {
		return nil, err
	}
	baseReserve := decimal.NewFromUint64(poolBaseTokenAccount.Amount)

	poolQuoteTokenAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolQuoteTokenAccount)
	if err != nil {
		return nil, err
	}
	quoteReserve := decimal.NewFromUint64(poolQuoteTokenAccount.Amount)

	if baseReserve.Sign() == 0 || quoteReserve.Sign() == 0 {
		return nil, errors.New("invalid input: baseReserve or quoteReserve cannot be zero")
	}
	if quoteDecimal.Cmp(quoteReserve) > 0 {
		return nil, errors.New("cannot receive more quote tokens than the pool quote reserves")
	}
	fees, err := ComputeFeesBps(globalConfig, feeConfig, pool.Creator, baseMintSupplyDecimal, pool.BaseMint, baseReserve, quoteReserve, quoteDecimal)
	if err != nil {
		return nil, err
	}

	totalFee := decimal.NewFromUint64(fees.LpFeeBps).Add(decimal.NewFromUint64(fees.ProtocolFeeBps))
	if !pool.CoinCreator.Equals(solana.PublicKey{}) {
		totalFee = totalFee.Add(decimal.NewFromUint64(fees.CreatorFeeBps))
	}

	rawDen := totalFee.Sub(helpers.BpsDenominator)
	rawNum := quoteDecimal.Mul(helpers.BpsDenominator)
	rawQuote, _ := helpers.CeilDiv(rawNum, rawDen)
	if rawQuote.Cmp(quoteReserve) >= 0 {
		return nil, errors.New("desired quote amount exceeds available reserve")
	}

	baseIn, _ := helpers.CeilDiv(baseReserve.Mul(rawQuote), quoteReserve.Sub(rawQuote))
	f := helpers.SlippageFactor(slippage, false)
	minQuote := quoteDecimal.Mul(f).Div(helpers.SlippageScale)
	return &SellQuoteInputResult{
		InternalRawQuote: rawQuote.BigInt().Uint64(),
		Base:             baseIn.BigInt().Uint64(),
		MinQuote:         minQuote.BigInt().Uint64(),
	}, nil
}

func (s *Client) CreatePoolInstructions(
	ctx context.Context,
	params *CreatePoolParams,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	baseTokenProgram solana.PublicKey,
	quoteTokenProgram solana.PublicKey,
) ([]solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	userBaseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	userQuoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, userQuoteTokenAccount, params.QuoteIn)
	if err != nil {
		return nil, nil, nil, err
	}
	pre = append(pre, wrapIx...)

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	ix, err := s.CreatePoolInstruction(params, user, baseMint, quoteMint, userBaseTokenAccount, userQuoteTokenAccount, baseTokenProgram, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}
	return ix, pre, post, nil
}

func (s *Client) BuyInstructions(
	ctx context.Context,
	params *BuyParams,
	globalConfig *GlobalConfig,
	pool *AccountWithPool,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	baseTokenProgram solana.PublicKey,
	quoteTokenProgram solana.PublicKey,
) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {

	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	userBaseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	userQuoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, userQuoteTokenAccount, params.MaxQuoteIn)
	if err != nil {
		return nil, nil, nil, err
	}
	pre = append(pre, wrapIx...)

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	buyIx, err := s.BuyInstruction(params, globalConfig, pool, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount)
	if err != nil {
		return nil, nil, nil, err
	}

	return buyIx, pre, post, nil
}

func (s *Client) SellInstructions(
	ctx context.Context,
	params *SellParams,
	globalConfig *GlobalConfig,
	pool *AccountWithPool,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	baseTokenProgram solana.PublicKey,
	quoteTokenProgram solana.PublicKey,
) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	userBaseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixA != nil {
		pre = append(pre, ixA)
	}

	userQuoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}
	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	buyIx, err := s.SellInstruction(params, globalConfig, pool, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, userBaseTokenAccount, userQuoteTokenAccount)
	if err != nil {
		return nil, nil, nil, err
	}

	return buyIx, pre, post, nil

}

func (s *Client) DepositInstructions(
	ctx context.Context,
	params *DepositParams,
	pool *AccountWithPool,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	baseTokenProgram solana.PublicKey,
	quoteTokenProgram solana.PublicKey,
) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {

	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	userBaseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}
	if ixA != nil {
		pre = append(pre, ixA)
	}

	userQuoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}
	if ixB != nil {
		pre = append(pre, ixB)
	}

	wrapIx, err := helpers.WrapSOLInstruction(user, userQuoteTokenAccount, params.MaxQuoteIn)
	if err != nil {
		return nil, nil, nil, err
	}
	pre = append(pre, wrapIx...)

	userPoolTokenAccount, ixC, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, pool.Account.LpMint, user, user, solana.Token2022ProgramID)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixC != nil {
		pre = append(pre, ixC)
	}

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	ix, err := s.DepositInstruction(
		params,
		pool,
		user,
		baseMint,
		quoteMint,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		userPoolTokenAccount,
	)

	if err != nil {
		return nil, nil, nil, err
	}

	return ix, pre, post, nil
}

func (s *Client) WithdrawInstructions(
	ctx context.Context,
	params *WithdrawParams,
	pool *AccountWithPool,
	user solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	baseTokenProgram solana.PublicKey,
	quoteTokenProgram solana.PublicKey,
) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)

	userBaseTokenAccount, ixA, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}
	if ixA != nil {
		pre = append(pre, ixA)
	}

	userQuoteTokenAccount, ixB, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixB != nil {
		pre = append(pre, ixB)
	}

	userPoolTokenAccount, ixC, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, pool.Account.LpMint, user, user, solana.Token2022ProgramID)
	if err != nil {
		return nil, nil, nil, err
	}

	if ixC != nil {
		pre = append(pre, ixC)
	}

	unwrapIx, uerr := helpers.UnwrapSOLInstruction(user, user, true)
	if uerr == nil && unwrapIx != nil {
		post = append(post, unwrapIx)
	}

	ix, err := s.WithdrawInstruction(
		params,
		pool,
		user,
		baseMint,
		quoteMint,
		userBaseTokenAccount,
		userQuoteTokenAccount,
		userPoolTokenAccount,
	)

	if err != nil {
		return nil, nil, nil, err
	}

	return ix, pre, post, nil
}
