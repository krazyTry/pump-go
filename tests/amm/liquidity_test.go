package amm

import (
	"context"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/krazyTry/pump-go/amm"
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
)

func TestDeposit(t *testing.T) {
	ammService := amm.NewClient(rpcClient, rpc.CommitmentFinalized)

	// 7kCd7jFy6MMKSnL5So96RSSarsBDiBrxtTy7semGoBgZ 5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1
	// FfzvytqAzMFFg3av6Ht5HbaXwdqyogkZua1AvgZFLiRd 5gwR5Wzmhy3hqGdJv36f4D6ZRAz8i18XbyGgj4STkRMSt3cneV91a5emmNqVsfwESq2GXCXR4Adcr5LKQ5uu7Zk3
	partner := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1")}
	fmt.Println("partner address:", partner.PublicKey())

	// SPL
	// 6kh2zgx1GdEKcvByc9Wi68DBXVuYP4mRgZxtsLVgNyuv &{LTQq74waDewBH7HoWYcXUBxPn3eQsKnM5RVnSdbzHGoYYskD6vSPaekJ9N4CfHEDtmdcDB6uiSzj4pyirKBWorU}
	// pool: 2zgcfbn5Gh8jD46YeV7FG81rKVaPRHhTvVrUWSrhhCts
	// lpMint: FsS6cuQJb6VS4GvJz5r2LywCrSRzJG2yDbT1FQhbtrPP

	// 2022
	// 8NvYAiyLYhefUtgAwkhuWnJnxHUJio28tfoLXy77bzPy &{2wG7dfCLbLAkGJPY6TsU6mgnoT1nUnzJmJnYrxBq7EBUV9aXNSg4kSpma9j2ZAEXCAkKQsxf5Y7ZpMgWhWdzRNMV}
	// pool: 3heEq1hv9JiLtNnzMU9J3qev1Qf9hF9zAzr1ZXN9A8Mb
	// lpMint: 9fhxrYR6bdqvU6FUXR3Q6Aqp8t53bS3ddtRMZiyWySCy

	// pool, err := ammService.FetchPool(context.Background(), solana.MustPublicKeyFromBase58("2zgcfbn5Gh8jD46YeV7FG81rKVaPRHhTvVrUWSrhhCts"))
	// if err != nil {
	// 	t.Fatalf("FetchPool failed: %v", err)
	// }

	// baseMint := solana.MustPublicKeyFromBase58("6kh2zgx1GdEKcvByc9Wi68DBXVuYP4mRgZxtsLVgNyuv")
	// baseTokenProgram := solana.TokenProgramID

	baseMint := solana.MustPublicKeyFromBase58("8NvYAiyLYhefUtgAwkhuWnJnxHUJio28tfoLXy77bzPy")
	baseTokenProgram := solana.Token2022ProgramID

	pools, err := ammService.FetchPoolByBaseMint(context.Background(), baseMint)
	if err != nil {
		t.Fatal(err)
	}

	pool := pools[0]
	quoteMint := pool.Account.QuoteMint
	quoteTokenProgram := solana.TokenProgramID
	if pool.Account.QuoteMint != solana.WrappedSol {
		quoteTokenProgram, _ = helpers.GetTokenProgram(context.TODO(), rpcClient, pool.Account.QuoteMint)
	}

	{
		depositToken0Result, err := ammService.DepositToken0(1e5, 10, pool.Account)
		if err != nil {
			t.Fatal(err)
		}

		params := &amm.DepositParams{
			LpOut:      depositToken0Result.LpToken,
			MaxBaseIn:  depositToken0Result.MaxToken0,
			MaxQuoteIn: depositToken0Result.MaxToken1,
		}

		depositIx, pre, post, err := ammService.DepositInstructions(context.Background(), params, pool, partner.PublicKey(), baseMint, quoteMint, baseTokenProgram, quoteTokenProgram)
		if err != nil {
			t.Fatal(err)
		}

		var instructions []solana.Instruction
		instructions = append(instructions, pre...)
		instructions = append(instructions, depositIx)
		instructions = append(instructions, post...)

		sig, err := SendInstruction(context.Background(), rpcClient, wsClient, instructions, partner.PublicKey(), func(key solana.PublicKey) *solana.PrivateKey {
			switch {
			case key.Equals(partner.PublicKey()):
				return &partner.PrivateKey
			default:
				return nil
			}
		})
		if err != nil {
			t.Fatal("deposit SendTransaction() fail", err)
		}
		fmt.Println("deposit token success Success sig:", sig.String())
	}

	{
		lpAmount, err := ammService.FetchPoolLiquidityAmountByUser(context.Background(), pool.Account, partner.PublicKey())
		if err != nil {
			t.Fatal(err)
		}

		withdrawResult, err := ammService.Withdraw(lpAmount, 10, pool.Account)
		if err != nil {
			t.Fatal(err)
		}

		params := &amm.WithdrawParams{
			LpIn:        lpAmount,
			MinBaseOut:  withdrawResult.MinBase,
			MinQuoteOut: withdrawResult.MinQuote,
		}

		withdrawIx, pre, post, err := ammService.WithdrawInstructions(context.Background(), params, pool, partner.PublicKey(), baseMint, quoteMint, baseTokenProgram, quoteTokenProgram)
		if err != nil {
			t.Fatal(err)
		}

		var instructions []solana.Instruction
		instructions = append(instructions, pre...)
		instructions = append(instructions, withdrawIx)
		instructions = append(instructions, post...)

		sig, err := SendInstruction(context.Background(), rpcClient, wsClient, instructions, partner.PublicKey(), func(key solana.PublicKey) *solana.PrivateKey {
			switch {
			case key.Equals(partner.PublicKey()):
				return &partner.PrivateKey
			default:
				return nil
			}
		})
		if err != nil {
			t.Fatal("withdraw SendTransaction() fail", err)
		}
		fmt.Println("withdraw token success Success sig:", sig.String())
	}
}
