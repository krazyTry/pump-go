package amm

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	solana "github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/rpc"
)

func (s *Client) accountData(ctx context.Context, key solana.PublicKey) ([]byte, *rpc.Account, error) {
	acc, err := s.RPC.GetAccountInfoWithOpts(ctx, key, &rpc.GetAccountInfoOpts{Commitment: s.Commitment})
	if err != nil || acc == nil || acc.Value == nil {
		return nil, nil, fmt.Errorf("account %s not found", key)
	}
	return acc.Value.Data.GetBinary(), acc.Value, nil
}

func (s *Client) multipleAccounts(ctx context.Context, keys []solana.PublicKey) ([]*rpc.Account, error) {
	out := make([]*rpc.Account, 0, len(keys))
	for start := 0; start < len(keys); start += 100 {
		end := start + 100
		if end > len(keys) {
			end = len(keys)
		}
		res, err := s.RPC.GetMultipleAccountsWithOpts(ctx, keys[start:end], &rpc.GetMultipleAccountsOpts{Commitment: s.Commitment})
		if err != nil {
			return nil, err
		}
		if res == nil || len(res.Value) != end-start {
			return nil, fmt.Errorf("getMultipleAccounts returned %d accounts", 0)
		}
		out = append(out, res.Value...)
	}
	return out, nil
}

func tokenAmount(acc *rpc.Account, tokenProgram solana.PublicKey) uint64 {
	if acc == nil || !acc.Owner.Equals(tokenProgram) {
		return 0
	}
	data := acc.Data.GetBinary()
	if len(data) < 72 {
		return 0
	}
	return binary.LittleEndian.Uint64(data[64:72])
}

func mintSupply(data []byte) (uint64, error) {
	if len(data) < 44 {
		return 0, fmt.Errorf("mint account is %d bytes", len(data))
	}
	return binary.LittleEndian.Uint64(data[36:44]), nil
}

func (s *Client) resolveQuote(ctx context.Context, quoteMint, quoteTokenProgram solana.PublicKey) (solana.PublicKey, solana.PublicKey, error) {
	quoteMint = CanonicalPoolQuoteMint(quoteMint)
	if !quoteTokenProgram.Equals(solana.PublicKey{}) {
		return quoteMint, quoteTokenProgram, nil
	}
	_, acc, err := s.accountData(ctx, quoteMint)
	if err != nil {
		return solana.PublicKey{}, solana.PublicKey{}, err
	}
	return quoteMint, acc.Owner, nil
}

// GetCoinCreatorVaultBalance is the creator's uncollected AMM fees quoted in quoteMint.
// A missing vault reads as zero. A zero quoteMint is WSOL.
func (s *Client) GetCoinCreatorVaultBalance(ctx context.Context, coinCreator, quoteMint, quoteTokenProgram solana.PublicKey) (uint64, error) {
	quoteMint, quoteTokenProgram, err := s.resolveQuote(ctx, quoteMint, quoteTokenProgram)
	if err != nil {
		return 0, err
	}
	ata := CoinCreatorVaultAta(DeriveCoinCreatorVault(coinCreator), quoteMint, quoteTokenProgram)
	acc, err := s.RPC.GetAccountInfoWithOpts(ctx, ata, &rpc.GetAccountInfoOpts{Commitment: s.Commitment})
	if err != nil || acc == nil || acc.Value == nil {
		return 0, nil
	}
	return tokenAmount(acc.Value, quoteTokenProgram), nil
}

// QuoteMint is one quote a creator vault may hold.
type QuoteMint struct {
	Mint         solana.PublicKey
	TokenProgram solana.PublicKey
}

// GetCoinCreatorVaultBalances returns uncollected fees for each quote, keyed by the mint address.
// Mints are used as given. A missing vault reads as zero.
func (s *Client) GetCoinCreatorVaultBalances(ctx context.Context, coinCreator solana.PublicKey, quotes []QuoteMint) (map[solana.PublicKey]uint64, error) {
	authority := DeriveCoinCreatorVault(coinCreator)
	atas := make([]solana.PublicKey, len(quotes))
	for i, quote := range quotes {
		atas[i] = CoinCreatorVaultAta(authority, quote.Mint, quote.TokenProgram)
	}
	accounts, err := s.multipleAccounts(ctx, atas)
	if err != nil {
		return nil, err
	}
	out := make(map[solana.PublicKey]uint64, len(quotes))
	for i, quote := range quotes {
		out[quote.Mint] = tokenAmount(accounts[i], quote.TokenProgram)
	}
	return out, nil
}

func (s *Client) volumeAccumulators(ctx context.Context, user solana.PublicKey) (*GlobalVolumeAccumulator, *UserVolumeAccumulator, error) {
	accounts, err := s.multipleAccounts(ctx, []solana.PublicKey{s.GlobalVolumeAccumulator, DeriveUserVolumeAccumulator(user)})
	if err != nil {
		return nil, nil, err
	}
	if accounts[0] == nil || accounts[1] == nil {
		return nil, nil, nil
	}
	global, err := DecodeGlobalVolumeAccumulator(accounts[0].Data.GetBinary())
	if err != nil {
		return nil, nil, err
	}
	userAcc, err := DecodeUserVolumeAccumulator(accounts[1].Data.GetBinary())
	if err != nil {
		return nil, nil, err
	}
	return global, userAcc, nil
}

// GetTotalUnclaimedTokens is the user's unclaimed incentive tokens. Missing accounts read as zero.
func (s *Client) GetTotalUnclaimedTokens(ctx context.Context, user solana.PublicKey) (uint64, error) {
	global, userAcc, err := s.volumeAccumulators(ctx, user)
	if err != nil || global == nil || userAcc == nil {
		return 0, err
	}
	return TotalUnclaimedTokens(global, userAcc, time.Now().Unix()), nil
}

// GetCurrentDayTokens is the user's incentive share for the current day. Missing accounts read as zero.
func (s *Client) GetCurrentDayTokens(ctx context.Context, user solana.PublicKey) (uint64, error) {
	global, userAcc, err := s.volumeAccumulators(ctx, user)
	if err != nil || global == nil || userAcc == nil {
		return 0, err
	}
	return CurrentDayTokens(global, userAcc, time.Now().Unix()), nil
}

// SetCoinCreatorInstructions builds set_coin_creator, preceded by extend_account when the pool is shorter than PoolAccountNewSize.
func (s *Client) SetCoinCreatorInstructions(ctx context.Context, poolKey, payer solana.PublicKey) ([]solana.Instruction, error) {
	data, _, err := s.accountData(ctx, poolKey)
	if err != nil {
		return nil, err
	}
	pool, err := DecodePool(data)
	if err != nil {
		return nil, err
	}
	var out []solana.Instruction
	if len(data) < PoolAccountNewSize {
		ix, err := s.ExtendAccountInstruction(poolKey, payer)
		if err != nil {
			return nil, err
		}
		out = append(out, ix)
	}
	ix, err := s.SetCoinCreatorInstruction(poolKey, DeriveMetadata(pool.BaseMint), DerivePumpBondingCurve(pool.BaseMint))
	if err != nil {
		return nil, err
	}
	return append(out, ix), nil
}

// SweepCreatorFee reads the pool and its quote token program, then builds sweep_creator_fee.
func (s *Client) SweepCreatorFee(ctx context.Context, poolKey, payer solana.PublicKey) (solana.Instruction, error) {
	pool, err := s.FetchPool(ctx, poolKey)
	if err != nil {
		return nil, err
	}
	_, quoteProgram, err := s.resolveQuote(ctx, pool.Account.QuoteMint, solana.PublicKey{})
	if err != nil {
		return nil, err
	}
	return s.SweepCreatorFeeInstruction(payer, poolKey, pool.Account, quoteProgram)
}

// TransferCreatorFeesToPumpV2 resolves the quote token program and builds transfer_creator_fees_to_pump_v2.
// A zero quoteMint is WSOL.
func (s *Client) TransferCreatorFeesToPumpV2(ctx context.Context, payer, coinCreator, quoteMint solana.PublicKey) (solana.Instruction, error) {
	quoteMint, quoteProgram, err := s.resolveQuote(ctx, quoteMint, solana.PublicKey{})
	if err != nil {
		return nil, err
	}
	return s.TransferCreatorFeesToPumpV2Instruction(payer, coinCreator, quoteMint, quoteProgram)
}

// BoostBuyAndBurn reads the pool and its mint owners, then builds boost_buy_and_burn.
func (s *Client) BoostBuyAndBurn(ctx context.Context, poolKey, authority solana.PublicKey, quoteAmountIn, minBaseAmountBurned uint64) (solana.Instruction, error) {
	data, _, err := s.accountData(ctx, poolKey)
	if err != nil {
		return nil, err
	}
	pool, err := DecodePool(data)
	if err != nil {
		return nil, err
	}
	accounts, err := s.multipleAccounts(ctx, []solana.PublicKey{pool.BaseMint, pool.QuoteMint})
	if err != nil {
		return nil, err
	}
	if accounts[0] == nil {
		return nil, fmt.Errorf("baseMint=%s not found", pool.BaseMint)
	}
	if accounts[1] == nil {
		return nil, fmt.Errorf("quoteMint=%s not found", pool.QuoteMint)
	}
	return s.BoostBuyAndBurnInstruction(poolKey, pool, authority, quoteAmountIn, minBaseAmountBurned, accounts[0].Owner, accounts[1].Owner)
}

// ClaimTokenIncentives builds claim_token_incentives. An unset mint or a missing accumulator returns no instructions.
func (s *Client) ClaimTokenIncentives(ctx context.Context, user, payer solana.PublicKey) ([]solana.Instruction, error) {
	global, err := s.FetchGlobalVolumeAccumulator(ctx, s.GlobalVolumeAccumulator)
	if err != nil || global.Mint.Equals(solana.PublicKey{}) {
		return nil, nil
	}
	accounts, err := s.multipleAccounts(ctx, []solana.PublicKey{global.Mint, DeriveUserVolumeAccumulator(user)})
	if err != nil {
		return nil, err
	}
	if accounts[0] == nil || accounts[1] == nil {
		return nil, nil
	}
	tokenProgram := accounts[0].Owner
	ix, err := s.ClaimTokenIncentivesInstruction(
		user,
		payer,
		CoinCreatorVaultAta(user, global.Mint, tokenProgram),
		CoinCreatorVaultAta(s.GlobalVolumeAccumulator, global.Mint, tokenProgram),
		global.Mint,
		tokenProgram,
	)
	if err != nil {
		return nil, err
	}
	return []solana.Instruction{ix}, nil
}

// MultiHopPoolHopsResult is the chain state multiHopSwapQuote prices a pool route from.
type MultiHopPoolHopsResult struct {
	GlobalConfig *GlobalConfig
	FeeConfig    *FeeConfig
	Hops         []MultiHopHop
}

// MultiHopPoolHops loads canonical pool hops in poolKeys order.
func (s *Client) MultiHopPoolHops(ctx context.Context, poolKeys []solana.PublicKey) (*MultiHopPoolHopsResult, error) {
	keys := append([]solana.PublicKey{s.GlobalConfig, s.FeeConfig}, poolKeys...)
	accounts, err := s.multipleAccounts(ctx, keys)
	if err != nil {
		return nil, err
	}
	if accounts[0] == nil {
		return nil, fmt.Errorf("global config account not found")
	}
	global, err := DecodeGlobalConfig(accounts[0].Data.GetBinary())
	if err != nil {
		return nil, err
	}
	var feeConfig *FeeConfig
	if accounts[1] != nil {
		feeConfig, err = DecodeFeeConfig(accounts[1].Data.GetBinary())
		if err != nil {
			return nil, err
		}
	}
	pools := make([]*Pool, len(poolKeys))
	for i := range poolKeys {
		if accounts[i+2] == nil {
			return nil, fmt.Errorf("pool=%s not found", poolKeys[i])
		}
		pools[i], err = DecodePool(accounts[i+2].Data.GetBinary())
		if err != nil {
			return nil, err
		}
	}
	seen := map[string]solana.PublicKey{}
	var extra []solana.PublicKey
	for _, pool := range pools {
		for _, key := range []solana.PublicKey{pool.BaseMint, pool.QuoteMint, pool.PoolBaseTokenAccount, pool.PoolQuoteTokenAccount} {
			id := key.String()
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = key
			extra = append(extra, key)
		}
	}
	extraAccounts, err := s.multipleAccounts(ctx, extra)
	if err != nil {
		return nil, err
	}
	byKey := map[string]*rpc.Account{}
	for i, key := range extra {
		if extraAccounts[i] == nil {
			return nil, fmt.Errorf("account=%s not found", key)
		}
		byKey[key.String()] = extraAccounts[i]
	}
	info := func(key solana.PublicKey) *rpc.Account { return byKey[key.String()] }
	epoch, err := s.RPC.GetEpochInfo(ctx, s.Commitment)
	if err != nil {
		return nil, err
	}
	hops := make([]MultiHopHop, len(pools))
	for i, pool := range pools {
		baseInfo := info(pool.BaseMint)
		quoteInfo := info(pool.QuoteMint)
		supply, err := mintSupply(baseInfo.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		var quoteFee *QuoteTransferFee
		cfg, err := ParseTransferFeeConfig(quoteInfo.Owner, quoteInfo.Data.GetBinary())
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			fee, err := TransferFeeAt(cfg, epoch.Epoch)
			if err != nil {
				return nil, err
			}
			quoteFee = &fee
		}
		hops[i] = MultiHopHop{
			Kind:              "pool",
			PoolKey:           poolKeys[i],
			Pool:              pool,
			PoolBaseAmount:    tokenAmount(info(pool.PoolBaseTokenAccount), baseInfo.Owner),
			PoolQuoteAmount:   tokenAmount(info(pool.PoolQuoteTokenAccount), quoteInfo.Owner),
			BaseMintSupply:    supply,
			BaseTokenProgram:  baseInfo.Owner,
			QuoteTokenProgram: quoteInfo.Owner,
			QuoteTransferFee:  quoteFee,
		}
	}
	return &MultiHopPoolHopsResult{GlobalConfig: global, FeeConfig: feeConfig, Hops: hops}, nil
}
