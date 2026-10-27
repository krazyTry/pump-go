package amm

import (
	"context"

	"github.com/krazyTry/pump-go/amm/helpers"
	solana "github.com/solana-foundation/solana-go/v2"
)

func quoteTransferFeeFromToken(info *helpers.TokenInfo) *QuoteTransferFee {
	if info == nil || info.Fee.FeeBps == 0 {
		return nil
	}
	return &QuoteTransferFee{BasisPoints: uint64(info.Fee.FeeBps), MaximumFee: info.Fee.MaxFee}
}

func (s *Client) loadSwapState(pool *Pool) (supply, baseReserve, quoteReserve uint64, quoteFee *QuoteTransferFee, err error) {
	baseMint, err := helpers.GetTokenInfo(context.TODO(), s.RPC, pool.BaseMint)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	quoteMint, err := helpers.GetTokenInfo(context.TODO(), s.RPC, pool.QuoteMint)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	baseAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolBaseTokenAccount)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	quoteAccount, err := helpers.GetAccountInfo(context.TODO(), s.RPC, pool.PoolQuoteTokenAccount)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	return baseMint.Mint.Supply, baseAccount.Amount, quoteAccount.Amount, quoteTransferFeeFromToken(quoteMint), nil
}

func (s *Client) DepositToken0(token0 uint64, slippage float64, pool *Pool) (*DepositResult, error) {
	_, baseReserve, quoteReserve, quoteFee, err := s.loadSwapState(pool)
	if err != nil {
		return nil, err
	}
	realQuote, err := RealQuoteReserves(quoteReserve, pool.ProtocolFees, pool.CreatorFees)
	if err != nil {
		return nil, err
	}
	result, err := DepositBase(DepositBaseParams{
		Base:             token0,
		Slippage:         slippage,
		BaseReserve:      baseReserve,
		QuoteReserve:     realQuote,
		TotalLpTokens:    pool.LpSupply,
		QuoteTransferFee: quoteFee,
	})
	if err != nil {
		return nil, err
	}
	return &DepositResult{
		Token1:    result.Quote,
		LpToken:   result.LpToken,
		MaxToken0: result.MaxBase,
		MaxToken1: result.MaxQuote,
	}, nil
}

func (s *Client) DepositLpToken(lpToken uint64, slippage float64, pool *Pool) (*DepositLpTokenResult, error) {
	_, baseReserve, quoteReserve, quoteFee, err := s.loadSwapState(pool)
	if err != nil {
		return nil, err
	}
	realQuote, err := RealQuoteReserves(quoteReserve, pool.ProtocolFees, pool.CreatorFees)
	if err != nil {
		return nil, err
	}
	result, err := DepositLpToken(DepositLpTokenParams{
		LpToken:          lpToken,
		Slippage:         slippage,
		BaseReserve:      baseReserve,
		QuoteReserve:     realQuote,
		TotalLpTokens:    pool.LpSupply,
		QuoteTransferFee: quoteFee,
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Client) Withdraw(lpAmount uint64, slippage float64, pool *Pool) (*WithdrawResult, error) {
	_, baseReserve, quoteReserve, quoteFee, err := s.loadSwapState(pool)
	if err != nil {
		return nil, err
	}
	realQuote, err := RealQuoteReserves(quoteReserve, pool.ProtocolFees, pool.CreatorFees)
	if err != nil {
		return nil, err
	}
	result, err := Withdraw(WithdrawParamsCalc{
		LpAmount:         lpAmount,
		Slippage:         slippage,
		BaseReserve:      baseReserve,
		QuoteReserve:     realQuote,
		TotalLpTokens:    pool.LpSupply,
		QuoteTransferFee: quoteFee,
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// TradeOptions selects buy_v2 pricing when the pool accepts it.
type TradeOptions struct {
	V2 bool
}

func tradeV2(pool *Pool, opts []TradeOptions) bool {
	return len(opts) > 0 && opts[0].V2 && SupportsTradeV2(pool)
}

func (s *Client) BuyBaseInput(base uint64, slippage float64, pool *Pool, globalConfig *GlobalConfig, feeConfig *FeeConfig, opts ...TradeOptions) (*BuyBaseInputResult, error) {
	supply, baseReserve, quoteReserve, quoteFee, err := s.loadSwapState(pool)
	if err != nil {
		return nil, err
	}
	result, err := BuyBaseInput(BuyBaseInputParams{
		Base:                 base,
		Slippage:             slippage,
		BaseReserve:          baseReserve,
		QuoteReserve:         quoteReserve,
		VirtualQuoteReserves: pool.VirtualQuoteReserves.BigInt(),
		GlobalConfig:         globalConfig,
		BaseMintSupply:       supply,
		BaseMint:             pool.BaseMint,
		CoinCreator:          pool.CoinCreator,
		Creator:              pool.Creator,
		FeeConfig:            feeConfig,
		QuoteMint:            pool.QuoteMint,
		IsMayhemMode:         pool.IsMayhemMode,
		CreatorFeeBps:        pool.CreatorFeeBps,
		QuoteTransferFee:     quoteFee,
		TradeV2:              tradeV2(pool, opts),
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Client) BuyQuoteInput(quote uint64, slippage float64, pool *Pool, globalConfig *GlobalConfig, feeConfig *FeeConfig, opts ...TradeOptions) (*BuyQuoteInputResult, error) {
	supply, baseReserve, quoteReserve, quoteFee, err := s.loadSwapState(pool)
	if err != nil {
		return nil, err
	}
	result, err := BuyQuoteInput(BuyQuoteInputParams{
		Quote:                quote,
		Slippage:             slippage,
		BaseReserve:          baseReserve,
		QuoteReserve:         quoteReserve,
		VirtualQuoteReserves: pool.VirtualQuoteReserves.BigInt(),
		GlobalConfig:         globalConfig,
		BaseMintSupply:       supply,
		BaseMint:             pool.BaseMint,
		CoinCreator:          pool.CoinCreator,
		Creator:              pool.Creator,
		FeeConfig:            feeConfig,
		QuoteMint:            pool.QuoteMint,
		IsMayhemMode:         pool.IsMayhemMode,
		CreatorFeeBps:        pool.CreatorFeeBps,
		QuoteTransferFee:     quoteFee,
		TradeV2:              tradeV2(pool, opts),
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Client) SellBaseInput(base uint64, slippage float64, pool *Pool, globalConfig *GlobalConfig, feeConfig *FeeConfig) (*SellBaseInputResult, error) {
	supply, baseReserve, quoteReserve, quoteFee, err := s.loadSwapState(pool)
	if err != nil {
		return nil, err
	}
	buckets, err := RealQuoteReserves(quoteReserve, pool.ProtocolFees, pool.CreatorFees)
	if err != nil {
		return nil, err
	}
	feeBuckets := quoteReserve - buckets
	result, err := SellBaseInput(SellBaseInputParams{
		Base:                 base,
		Slippage:             slippage,
		BaseReserve:          baseReserve,
		QuoteReserve:         quoteReserve,
		VirtualQuoteReserves: pool.VirtualQuoteReserves.BigInt(),
		FeeBucketsTotal:      feeBuckets,
		GlobalConfig:         globalConfig,
		BaseMintSupply:       supply,
		BaseMint:             pool.BaseMint,
		CoinCreator:          pool.CoinCreator,
		Creator:              pool.Creator,
		FeeConfig:            feeConfig,
		QuoteMint:            pool.QuoteMint,
		IsMayhemMode:         pool.IsMayhemMode,
		CreatorFeeBps:        pool.CreatorFeeBps,
		QuoteTransferFee:     quoteFee,
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Client) SellQuoteInput(quote uint64, slippage float64, pool *Pool, globalConfig *GlobalConfig, feeConfig *FeeConfig) (*SellQuoteInputResult, error) {
	supply, baseReserve, quoteReserve, quoteFee, err := s.loadSwapState(pool)
	if err != nil {
		return nil, err
	}
	buckets, err := RealQuoteReserves(quoteReserve, pool.ProtocolFees, pool.CreatorFees)
	if err != nil {
		return nil, err
	}
	result, err := SellQuoteInput(SellQuoteInputParams{
		Quote:                quote,
		Slippage:             slippage,
		BaseReserve:          baseReserve,
		QuoteReserve:         quoteReserve,
		VirtualQuoteReserves: pool.VirtualQuoteReserves.BigInt(),
		FeeBucketsTotal:      quoteReserve - buckets,
		GlobalConfig:         globalConfig,
		BaseMintSupply:       supply,
		BaseMint:             pool.BaseMint,
		CoinCreator:          pool.CoinCreator,
		Creator:              pool.Creator,
		FeeConfig:            feeConfig,
		QuoteMint:            pool.QuoteMint,
		IsMayhemMode:         pool.IsMayhemMode,
		CreatorFeeBps:        pool.CreatorFeeBps,
		QuoteTransferFee:     quoteFee,
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
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
