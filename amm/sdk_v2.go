package amm

import (
	"context"
	"errors"

	"github.com/krazyTry/pump-go/amm/helpers"
	pump_amm "github.com/krazyTry/pump-go/gen/amm"
	solana "github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

func (s *Client) buybackATA(globalConfig *GlobalConfig, quoteMint, quoteTokenProgram solana.PublicKey) solana.PublicKey {
	return helpers.FindAssociatedTokenAddress(GetBuybackFeeRecipient(globalConfig), quoteMint, quoteTokenProgram)
}

// BuyV2Instruction builds buy_v2. The buyback account is the recipient's quote ATA.
func (s *Client) BuyV2Instruction(baseOut, maxQuoteIn uint64, globalConfig *GlobalConfig, pool *AccountWithPool, user, userBaseTokenAccount, userQuoteTokenAccount, baseTokenProgram, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewBuyV2Instruction(
		baseOut, maxQuoteIn,
		pool.PublicKey, user, s.GlobalConfig, pool.Account.BaseMint, pool.Account.QuoteMint,
		userBaseTokenAccount, userQuoteTokenAccount,
		pool.Account.PoolBaseTokenAccount, pool.Account.PoolQuoteTokenAccount,
		baseTokenProgram, quoteTokenProgram, solana.SystemProgramID,
		DeriveUserVolumeAccumulator(user), s.FeeConfig,
		s.buybackATA(globalConfig, pool.Account.QuoteMint, quoteTokenProgram),
		s.EventAuthority, ProgramID,
	)
}

// BuyExactQuoteInV2Instruction builds buy_exact_quote_in_v2.
func (s *Client) BuyExactQuoteInV2Instruction(spendableQuoteIn, minBaseOut uint64, globalConfig *GlobalConfig, pool *AccountWithPool, user, userBaseTokenAccount, userQuoteTokenAccount, baseTokenProgram, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewBuyExactQuoteInV2Instruction(
		spendableQuoteIn, minBaseOut,
		pool.PublicKey, user, s.GlobalConfig, pool.Account.BaseMint, pool.Account.QuoteMint,
		userBaseTokenAccount, userQuoteTokenAccount,
		pool.Account.PoolBaseTokenAccount, pool.Account.PoolQuoteTokenAccount,
		baseTokenProgram, quoteTokenProgram, solana.SystemProgramID,
		DeriveUserVolumeAccumulator(user), s.FeeConfig,
		s.buybackATA(globalConfig, pool.Account.QuoteMint, quoteTokenProgram),
		s.EventAuthority, ProgramID,
	)
}

// SellV2Instruction builds sell_v2.
func (s *Client) SellV2Instruction(baseIn, minQuoteOut uint64, globalConfig *GlobalConfig, pool *AccountWithPool, user, userBaseTokenAccount, userQuoteTokenAccount, baseTokenProgram, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	return pump_amm.NewSellV2Instruction(
		baseIn, minQuoteOut,
		pool.PublicKey, user, s.GlobalConfig, pool.Account.BaseMint, pool.Account.QuoteMint,
		userBaseTokenAccount, userQuoteTokenAccount,
		pool.Account.PoolBaseTokenAccount, pool.Account.PoolQuoteTokenAccount,
		baseTokenProgram, quoteTokenProgram, solana.SystemProgramID,
		DeriveUserVolumeAccumulator(user), s.FeeConfig,
		s.buybackATA(globalConfig, pool.Account.QuoteMint, quoteTokenProgram),
		s.EventAuthority, ProgramID,
	)
}

func (s *Client) prepareUserAccounts(ctx context.Context, user, baseMint, quoteMint, baseTokenProgram, quoteTokenProgram solana.PublicKey, wrapAmount uint64, wrap bool) (solana.PublicKey, solana.PublicKey, []solana.Instruction, []solana.Instruction, error) {
	pre := make([]solana.Instruction, 0)
	post := make([]solana.Instruction, 0)
	userBase, ixA, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, baseMint, user, user, baseTokenProgram)
	if err != nil {
		return solana.PublicKey{}, solana.PublicKey{}, nil, nil, err
	}
	if ixA != nil {
		pre = append(pre, ixA)
	}
	userQuote, ixB, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, quoteMint, user, user, quoteTokenProgram)
	if err != nil {
		return solana.PublicKey{}, solana.PublicKey{}, nil, nil, err
	}
	if ixB != nil {
		pre = append(pre, ixB)
	}
	if wrap && quoteMint.Equals(solana.WrappedSol) {
		wrapIx, err := helpers.WrapSOLInstruction(user, userQuote, wrapAmount)
		if err != nil {
			return solana.PublicKey{}, solana.PublicKey{}, nil, nil, err
		}
		pre = append(pre, wrapIx...)
		unwrap, uerr := helpers.UnwrapSOLInstruction(user, user, true)
		if uerr == nil && unwrap != nil {
			post = append(post, unwrap)
		}
	}
	return userBase, userQuote, pre, post, nil
}

// BuyV2Instructions builds buy_v2 plus the user's ATAs. SOL quotes are wrapped for maxQuoteIn.
func (s *Client) BuyV2Instructions(ctx context.Context, baseOut, maxQuoteIn uint64, globalConfig *GlobalConfig, pool *AccountWithPool, user, baseTokenProgram, quoteTokenProgram solana.PublicKey) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	userBase, userQuote, pre, post, err := s.prepareUserAccounts(ctx, user, pool.Account.BaseMint, pool.Account.QuoteMint, baseTokenProgram, quoteTokenProgram, maxQuoteIn, true)
	if err != nil {
		return nil, nil, nil, err
	}
	ix, err := s.BuyV2Instruction(baseOut, maxQuoteIn, globalConfig, pool, user, userBase, userQuote, baseTokenProgram, quoteTokenProgram)
	return ix, pre, post, err
}

// BuyExactQuoteInV2Instructions builds buy_exact_quote_in_v2 plus the user's ATAs.
func (s *Client) BuyExactQuoteInV2Instructions(ctx context.Context, minBaseOut, spendableQuoteIn uint64, globalConfig *GlobalConfig, pool *AccountWithPool, user, baseTokenProgram, quoteTokenProgram solana.PublicKey) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	userBase, userQuote, pre, post, err := s.prepareUserAccounts(ctx, user, pool.Account.BaseMint, pool.Account.QuoteMint, baseTokenProgram, quoteTokenProgram, spendableQuoteIn, true)
	if err != nil {
		return nil, nil, nil, err
	}
	ix, err := s.BuyExactQuoteInV2Instruction(spendableQuoteIn, minBaseOut, globalConfig, pool, user, userBase, userQuote, baseTokenProgram, quoteTokenProgram)
	return ix, pre, post, err
}

// SellV2Instructions builds sell_v2 plus the user's ATAs. A WSOL payout is unwrapped.
func (s *Client) SellV2Instructions(ctx context.Context, baseIn, minQuoteOut uint64, globalConfig *GlobalConfig, pool *AccountWithPool, user, baseTokenProgram, quoteTokenProgram solana.PublicKey) (solana.Instruction, []solana.Instruction, []solana.Instruction, error) {
	userBase, userQuote, pre, post, err := s.prepareUserAccounts(ctx, user, pool.Account.BaseMint, pool.Account.QuoteMint, baseTokenProgram, quoteTokenProgram, 0, false)
	if err != nil {
		return nil, nil, nil, err
	}
	if pool.Account.QuoteMint.Equals(solana.WrappedSol) {
		unwrap, uerr := helpers.UnwrapSOLInstruction(user, user, true)
		if uerr == nil && unwrap != nil {
			post = append(post, unwrap)
		}
	}
	ix, err := s.SellV2Instruction(baseIn, minQuoteOut, globalConfig, pool, user, userBase, userQuote, baseTokenProgram, quoteTokenProgram)
	return ix, pre, post, err
}

// InitBoostInstruction initializes the boost vault.
func (s *Client) InitBoostInstruction(poolKey solana.PublicKey, pool *Pool, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	authorityPDA := DeriveBoostVaultAuthority(poolKey)
	ix, err := pump_amm.NewInitBoostInstruction(
		poolKey, s.GlobalConfig, pool.Creator, pool.BaseMint, pool.QuoteMint,
		pool.PoolBaseTokenAccount, pool.PoolQuoteTokenAccount,
		authorityPDA, helpers.FindAssociatedTokenAddress(authorityPDA, pool.QuoteMint, quoteTokenProgram),
		quoteTokenProgram, solana.SystemProgramID, solana.SPLAssociatedTokenAccountProgramID, s.EventAuthority, ProgramID,
	)
	if err != nil {
		return nil, err
	}
	return ix, nil
}

// SetBoostAuthorityInstruction sets the boost authority for a pool.
func (s *Client) SetBoostAuthorityInstruction(poolKey solana.PublicKey, admin solana.PublicKey) (solana.Instruction, error) {
	authorityPDA := DeriveBoostVaultAuthority(poolKey)
	ix, err := pump_amm.NewSetBoostAuthorityInstruction(
		admin, s.GlobalConfig, authorityPDA, solana.SystemProgramID, s.EventAuthority, ProgramID,
	)
	if err != nil {
		return nil, err
	}
	return ix, nil
}

// BoostBuyAndBurnInstruction buys base with the boost vault and burns it. quoteMint is marked writable.
func (s *Client) BoostBuyAndBurnInstruction(poolKey solana.PublicKey, pool *Pool, authority solana.PublicKey, quoteAmountIn, minBaseAmountBurned uint64, baseTokenProgram, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {

	authorityPDA := DeriveBoostVaultAuthority(poolKey)
	ix, err := pump_amm.NewBoostBuyAndBurnInstruction(
		quoteAmountIn, minBaseAmountBurned,
		poolKey, authority, s.GlobalConfig, pool.BaseMint, pool.QuoteMint,
		pool.PoolBaseTokenAccount, pool.PoolQuoteTokenAccount,
		authorityPDA, helpers.FindAssociatedTokenAddress(authorityPDA, pool.QuoteMint, quoteTokenProgram),
		baseTokenProgram, quoteTokenProgram, s.EventAuthority, ProgramID,
	)
	if err != nil {
		return nil, err
	}
	gi := ix.(*solana.GenericInstruction)
	for _, meta := range gi.AccountValues {
		if meta.PublicKey.Equals(pool.QuoteMint) {
			meta.IsWritable = true
		}
	}
	return ix, nil
}

// SweepCreatorFeeInstruction pays pool.CreatorFees into the coin-creator vault.
// The quote leaves pool.PoolQuoteTokenAccount, the vault the pool actually holds.
func (s *Client) SweepCreatorFeeInstruction(payer, poolKey solana.PublicKey, pool *Pool, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	recipient := DeriveCoinCreatorVault(pool.CoinCreator)
	return pump_amm.NewSweepCreatorFeeInstruction(
		payer, s.GlobalConfig, poolKey, pool.QuoteMint, quoteTokenProgram, pool.PoolQuoteTokenAccount,
		recipient, CoinCreatorVaultAta(recipient, pool.QuoteMint, quoteTokenProgram),
		solana.SystemProgramID, solana.SPLAssociatedTokenAccountProgramID, s.EventAuthority, ProgramID,
	)
}

// SweepProtocolFeeInstruction pays pool.ProtocolFees to a listed protocol fee recipient.
func (s *Client) SweepProtocolFeeInstruction(payer, poolKey solana.PublicKey, pool *Pool, quoteTokenProgram solana.PublicKey, globalConfig *GlobalConfig) (solana.Instruction, error) {
	recipient := GetFeeRecipient(globalConfig, pool.IsMayhemMode)
	return pump_amm.NewSweepProtocolFeeInstruction(
		payer, s.GlobalConfig, poolKey, pool.QuoteMint, quoteTokenProgram, pool.PoolQuoteTokenAccount,
		recipient, CoinCreatorVaultAta(recipient, pool.QuoteMint, quoteTokenProgram),
		solana.SystemProgramID, solana.SPLAssociatedTokenAccountProgramID, s.EventAuthority, ProgramID,
	)
}

// TransferCreatorFeesToPumpV2Instruction moves AMM creator fees into the pump creator vault.
// coinCreator is the sharing-config PDA for a fee-shared coin.
func (s *Client) TransferCreatorFeesToPumpV2Instruction(payer, coinCreator, quoteMint, quoteTokenProgram solana.PublicKey) (solana.Instruction, error) {
	vault := DeriveCoinCreatorVault(coinCreator)
	pumpVault, _, _ := solana.FindProgramAddress([][]byte{[]byte("creator-vault"), coinCreator.Bytes()}, PumpProgramID)
	return pump_amm.NewTransferCreatorFeesToPumpV2Instruction(
		payer, quoteMint, quoteTokenProgram, solana.SystemProgramID, solana.SPLAssociatedTokenAccountProgramID,
		coinCreator, vault, helpers.FindAssociatedTokenAddress(vault, quoteMint, quoteTokenProgram),
		pumpVault, helpers.FindAssociatedTokenAddress(pumpVault, quoteMint, quoteTokenProgram),
		s.EventAuthority, ProgramID,
	)
}

// MultiHopSwapParams is the input of MultiHopSwapInstructions.
type MultiHopSwapParams struct {
	User                solana.PublicKey
	InMint              solana.PublicKey
	Hops                []MultiHopHop
	AmountIn            uint64
	MinAmountOut        uint64
	GlobalConfig        *GlobalConfig
	BuybackFeeRecipient solana.PublicKey
	UserInTokenAccount  solana.PublicKey
	UserOutTokenAccount solana.PublicKey
}

// MultiHopSwapInstructions builds multi_hop_swap. Hop accounts follow the fixed accounts.
// A custom user token account is left as given. The user's own ATA is created when it is missing,
// a wSOL input is wrapped, and a wSOL output is closed. A buy that starts on a SOL curve is not wrapped.
func (s *Client) MultiHopSwapInstructions(ctx context.Context, p MultiHopSwapParams) ([]solana.Instruction, error) {
	if p.AmountIn == 0 || p.MinAmountOut == 0 {
		return nil, errors.New("amountIn and minAmountOut must be positive.")
	}
	route, err := ResolveMultiHopRoute(p.InMint, p.Hops)
	if err != nil {
		return nil, err
	}
	inMint := route.Mints[0]
	outMint := route.Mints[len(route.Mints)-1]
	first, last := p.Hops[0], p.Hops[len(p.Hops)-1]
	inProgram := venueBaseTokenProgram(first)
	outProgram := VenueQuoteTokenProgram(last)
	if route.IsBuy {
		inProgram = VenueQuoteTokenProgram(first)
		outProgram = venueBaseTokenProgram(last)
	}
	protocol := last
	if route.IsBuy {
		protocol = first
	}
	buyback := p.BuybackFeeRecipient
	if buyback.Equals(solana.PublicKey{}) {
		if protocol.Kind == "curve" {
			return nil, errors.New("The route's protocol leg is a bonding curve: pass buybackFeeRecipient (a wallet listed in pump's Global.buybackFeeRecipients).")
		}
		buyback = GetBuybackFeeRecipient(p.GlobalConfig)
	}
	_, protocolQuote := venueMints(protocol)
	userIn := p.UserInTokenAccount
	if userIn.Equals(solana.PublicKey{}) {
		userIn = CoinCreatorVaultAta(p.User, inMint, inProgram)
	}
	userOut := p.UserOutTokenAccount
	if userOut.Equals(solana.PublicKey{}) {
		userOut = CoinCreatorVaultAta(p.User, outMint, outProgram)
	}
	ix, err := pump_amm.NewMultiHopSwapInstruction(
		p.AmountIn, p.MinAmountOut,
		p.User, userIn, userOut,
		s.GlobalConfig, s.FeeConfig, DeriveUserVolumeAccumulator(p.User),
		CoinCreatorVaultAta(buyback, protocolQuote, VenueQuoteTokenProgram(protocol)),
		solana.TokenProgramID, solana.Token2022ProgramID, solana.SystemProgramID,
		s.EventAuthority, ProgramID, PumpProgramID,
		DerivePumpGlobal(), DerivePumpFeeConfig(), DerivePumpEventAuthority(),
	)
	if err != nil {
		return nil, err
	}
	gi := ix.(*solana.GenericInstruction)
	for _, hop := range p.Hops {
		baseMint, quoteMint := venueMints(hop)
		venue := hop.PoolKey
		baseVault, quoteVault := solana.PublicKey{}, solana.PublicKey{}
		if hop.Kind == "pool" && hop.Pool != nil {
			baseVault = hop.Pool.PoolBaseTokenAccount
			quoteVault = hop.Pool.PoolQuoteTokenAccount
		} else {
			venue = DerivePumpBondingCurve(baseMint)
			baseVault = CoinCreatorVaultAta(venue, baseMint, venueBaseTokenProgram(hop))
			quoteVault = CoinCreatorVaultAta(venue, quoteMint, VenueQuoteTokenProgram(hop))
		}
		gi.AccountValues = append(gi.AccountValues,
			solana.NewAccountMeta(baseMint, false, false),
			solana.NewAccountMeta(quoteMint, false, false),
			solana.NewAccountMeta(venue, true, false),
			solana.NewAccountMeta(baseVault, true, false),
			solana.NewAccountMeta(quoteVault, true, false),
		)
	}
	var pre, post []solana.Instruction
	if userIn.Equals(CoinCreatorVaultAta(p.User, inMint, inProgram)) {
		_, create, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, inMint, p.User, p.User, inProgram)
		if err != nil {
			return nil, err
		}
		if create != nil {
			pre = append(pre, create)
		}
	}
	wrapAmount := p.AmountIn
	if route.IsBuy && IsSolCurve(first) {
		wrapAmount = 0
	}
	if inMint.Equals(solana.WrappedSol) && wrapAmount > 0 && userIn.Equals(CoinCreatorVaultAta(p.User, inMint, inProgram)) {
		wrap, err := helpers.WrapSOLInstruction(p.User, userIn, wrapAmount)
		if err != nil {
			return nil, err
		}
		pre = append(pre, wrap...)
		if closeIn, err := helpers.UnwrapSOLInstruction(p.User, p.User, true); err == nil && closeIn != nil {
			post = append(post, closeIn)
		}
	}
	if userOut.Equals(CoinCreatorVaultAta(p.User, outMint, outProgram)) {
		_, create, err := helpers.GetOrCreateATAInstruction(ctx, s.RPC, outMint, p.User, p.User, outProgram)
		if err != nil {
			return nil, err
		}
		if create != nil {
			pre = append(pre, create)
		}
	}
	if outMint.Equals(solana.WrappedSol) {
		if closeOut, err := helpers.UnwrapSOLInstruction(p.User, p.User, true); err == nil && closeOut != nil {
			post = append(post, closeOut)
		}
	}
	out := append(pre, ix)
	return append(out, post...), nil
}

// AutocompleteInitialPoolPrice is floor(initialQuote / initialBase), the initial price a new pool displays.
func AutocompleteInitialPoolPrice(initialBase, initialQuote uint64) (uint64, error) {
	if initialBase == 0 {
		return 0, errors.New("Division by zero: initialBase cannot be zero")
	}
	return initialQuote / initialBase, nil
}

// CollectCoinCreatorFeeInstructions collects AMM creator fees when the vault ATA exists.
// A missing vault returns no instructions. A WSOL payout the creator pays for is unwrapped.
func (s *Client) CollectCoinCreatorFeeInstructions(ctx context.Context, coinCreator, quoteMint, quoteTokenProgram, payer solana.PublicKey) ([]solana.Instruction, error) {
	if payer.Equals(solana.PublicKey{}) {
		payer = coinCreator
	}
	vaultAuth := DeriveCoinCreatorVault(coinCreator)
	vaultATA := helpers.FindAssociatedTokenAddress(vaultAuth, quoteMint, quoteTokenProgram)
	creatorATA := helpers.FindAssociatedTokenAddress(coinCreator, quoteMint, quoteTokenProgram)
	infos, err := s.RPC.GetMultipleAccounts(ctx, vaultATA, creatorATA)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return nil, err
	}
	var vaultInfo, creatorInfo *rpc.Account
	if infos != nil && len(infos.Value) > 0 {
		vaultInfo = infos.Value[0]
		if len(infos.Value) > 1 {
			creatorInfo = infos.Value[1]
		}
	}
	if vaultInfo == nil || !vaultInfo.Owner.Equals(quoteTokenProgram) {
		return nil, nil
	}
	var out []solana.Instruction
	if creatorInfo == nil || !creatorInfo.Owner.Equals(quoteTokenProgram) {
		out = append(out, helpers.CreateAssociatedTokenAccountInstruction(payer, creatorATA, coinCreator, quoteMint, quoteTokenProgram))
	}
	ix, err := s.CollectCoinCreatorFeeInstruction(quoteMint, quoteTokenProgram, coinCreator, vaultAuth, vaultATA, creatorATA)
	if err != nil {
		return nil, err
	}
	out = append(out, ix)
	if quoteMint.Equals(solana.WrappedSol) && coinCreator.Equals(payer) {
		unwrap, uerr := helpers.UnwrapSOLInstruction(coinCreator, coinCreator, true)
		if uerr == nil && unwrap != nil {
			out = append(out, unwrap)
		}
	}
	return out, nil
}
