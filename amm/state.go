package amm

import (
	"context"
	"fmt"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/krazyTry/pump-go/amm/helpers"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
)

func (s *Client) FetchGlobalConfig(ctx context.Context) (*GlobalConfig, error) {
	acc, err := s.RPC.GetAccountInfoWithOpts(ctx, s.GlobalConfig, &rpc.GetAccountInfoOpts{Commitment: s.Commitment})
	if err != nil || acc == nil || acc.Value == nil {
		return nil, fmt.Errorf("global config account not found")
	}
	return ammgen.ParseAccount_GlobalConfig(acc.Value.Data.GetBinary())
}

func (s *Client) FetchFeeConfig(ctx context.Context) (*FeeConfig, error) {
	acc, err := s.RPC.GetAccountInfoWithOpts(ctx, s.FeeConfig, &rpc.GetAccountInfoOpts{Commitment: s.Commitment})
	if err != nil || acc == nil || acc.Value == nil {
		return nil, fmt.Errorf("fee config account not found")
	}
	return ammgen.ParseAccount_FeeConfig(acc.Value.Data.GetBinary())
}

func (s *Client) FetchPool(ctx context.Context, pool solana.PublicKey) (*AccountWithPool, error) {
	acc, err := s.RPC.GetAccountInfoWithOpts(ctx, pool, &rpc.GetAccountInfoOpts{Commitment: s.Commitment})
	if err != nil || acc == nil || acc.Value == nil {
		return nil, fmt.Errorf("pool account not found")
	}

	pl, err := ammgen.ParseAccount_Pool(acc.Value.Data.GetBinary())
	if err != nil {
		return nil, err
	}

	return &AccountWithPool{PublicKey: pool, Account: pl}, nil
}

func (s *Client) FetchPoolByBaseMint(ctx context.Context, baseMint solana.PublicKey) ([]*AccountWithPool, error) {
	filters := helpers.CreateProgramAccountFilter(helpers.AccountKeyPool, &helpers.Filter{
		Owner:  baseMint,
		Offset: helpers.ComputeStructOffset(new(ammgen.Pool), "BaseMint"),
	})

	accs, err := s.RPC.GetProgramAccountsWithOpts(ctx, ProgramID, &rpc.GetProgramAccountsOpts{Commitment: s.Commitment, Filters: filters})
	if err != nil {
		return nil, err
	}
	out := []*AccountWithPool{}

	for _, acc := range accs {
		pl, err := ammgen.ParseAccount_Pool(acc.Account.Data.GetBinary())
		if err != nil {
			continue
		}
		out = append(out, &AccountWithPool{PublicKey: acc.Pubkey, Account: pl})
	}

	return out, nil
}

func (s *Client) FetchGlobalVolumeAccumulator(ctx context.Context, globalVolumeAccumulatorPDA solana.PublicKey) (*ammgen.GlobalVolumeAccumulator, error) {
	acc, err := s.RPC.GetAccountInfoWithOpts(ctx, globalVolumeAccumulatorPDA, &rpc.GetAccountInfoOpts{Commitment: s.Commitment})
	if err != nil || acc == nil || acc.Value == nil {
		return nil, fmt.Errorf("global volume accumulator account not found")
	}
	return ammgen.ParseAccount_GlobalVolumeAccumulator(acc.Value.Data.GetBinary())
}

func (s *Client) FetchUserVolumeAccumulator(ctx context.Context, user solana.PublicKey) (*ammgen.UserVolumeAccumulator, error) {
	acc, err := s.RPC.GetAccountInfoWithOpts(ctx, DeriveUserVolumeAccumulator(user), &rpc.GetAccountInfoOpts{Commitment: s.Commitment})
	if err != nil || acc == nil || acc.Value == nil {
		return nil, fmt.Errorf("user volume accumulator account not found")
	}
	return ammgen.ParseAccount_UserVolumeAccumulator(acc.Value.Data.GetBinary())
}

func (s *Client) FetchPoolLiquidityAmountByUser(ctx context.Context, pool *ammgen.Pool, user solana.PublicKey) (uint64, error) {
	userPoolTokenAccount := helpers.FindAssociatedTokenAddress(user, pool.LpMint, solana.Token2022ProgramID)

	account, err := helpers.GetAccountInfo(ctx, s.RPC, userPoolTokenAccount)
	if err != nil {
		return 0, err
	}
	return account.Amount, nil
}
