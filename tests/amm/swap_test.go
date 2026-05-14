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

func TestSwapV2(t *testing.T) {
	ammService := amm.NewClient(rpcClient, rpc.CommitmentFinalized)

	// FfzvytqAzMFFg3av6Ht5HbaXwdqyogkZua1AvgZFLiRd 5gwR5Wzmhy3hqGdJv36f4D6ZRAz8i18XbyGgj4STkRMSt3cneV91a5emmNqVsfwESq2GXCXR4Adcr5LKQ5uu7Zk3
	partner := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5gwR5Wzmhy3hqGdJv36f4D6ZRAz8i18XbyGgj4STkRMSt3cneV91a5emmNqVsfwESq2GXCXR4Adcr5LKQ5uu7Zk3")}
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

	global, _ := ammService.FetchGlobalConfig(context.Background())
	feeConfig, _ := ammService.FetchFeeConfig(context.Background())

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
		amountIn := uint64(0.01 * 1e9)
		buyQuoteInputResult, err := ammService.BuyQuoteInput(amountIn, 10, pool.Account, global, feeConfig)
		if err != nil {
			t.Fatal(err)
		}

		params := &amm.BuyParams{
			BaseOut:     buyQuoteInputResult.Base,
			MaxQuoteIn:  buyQuoteInputResult.MaxQuote,
			TrackVolume: true,
		}

		buyIx, pre, post, err := ammService.BuyInstructions(context.TODO(), params, global, pool, partner.PublicKey(), baseMint, quoteMint, baseTokenProgram, quoteTokenProgram)
		if err != nil {
			t.Fatal(err)
		}

		var instructions []solana.Instruction
		instructions = append(instructions, pre...)
		instructions = append(instructions, buyIx)
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
			t.Fatal("buy SendTransaction() fail", err)
		}
		fmt.Println("buy token success Success sig:", sig.String())
	}

	{
		amountIn, _ := MintBalance(context.Background(), rpcClient, partner.PublicKey(), baseMint)

		sellBaseInputResult, err := ammService.SellBaseInput(amountIn, 10, pool.Account, global, feeConfig)
		if err != nil {
			t.Fatal(err)
		}

		params := &amm.SellParams{
			BaseIn:      amountIn,
			MinQuoteOut: sellBaseInputResult.MinQuote,
		}

		sellIx, pre, post, err := ammService.SellInstructions(context.Background(), params, global, pool, partner.PublicKey(), baseMint, quoteMint, baseTokenProgram, quoteTokenProgram)
		if err != nil {
			t.Fatal(err)
		}

		var instructions []solana.Instruction
		instructions = append(instructions, pre...)
		instructions = append(instructions, sellIx)
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
			t.Fatal("sell SendTransaction() fail", err)
		}
		fmt.Println("sell token success Success sig:", sig.String())
	}

}
