package amm

import (
	"context"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/krazyTry/pump-go/amm"
)

// 7kCd7jFy6MMKSnL5So96RSSarsBDiBrxtTy7semGoBgZ 5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1
// 4YwmadgZWofhxn1f2HNjyeDx5eKUwU5WVhxC2ZTPMfKM 5vXSDWQMecswf8pvupiFkCAeQX5ZNrxvLvNifWrnQtybDQ3dEmWiNLPnfaM52jLtAFxrR9EmG2dng8BNrDHSsn7q
// FfzvytqAzMFFg3av6Ht5HbaXwdqyogkZua1AvgZFLiRd 5gwR5Wzmhy3hqGdJv36f4D6ZRAz8i18XbyGgj4STkRMSt3cneV91a5emmNqVsfwESq2GXCXR4Adcr5LKQ5uu7Zk3
// Ep6iSRnmcP1uwYV4eQKEkfPTQQrpp6nHby2yU9G1dwbo 2u4atG3WQr5Y7jZ3x7LPqDet26NeRjPTrjKxdzdoMqjQn1rsMGsGhkUyhDDGWJZbDgdyzruvv3SouEJeXtPCigT3

func TestClaim(t *testing.T) {
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

	// baseMint := solana.MustPublicKeyFromBase58("6kh2zgx1GdEKcvByc9Wi68DBXVuYP4mRgZxtsLVgNyuv")
	// baseTokenProgram := solana.TokenProgramID

	baseMint := solana.MustPublicKeyFromBase58("8NvYAiyLYhefUtgAwkhuWnJnxHUJio28tfoLXy77bzPy")

	pools, err := ammService.FetchPoolByBaseMint(context.Background(), baseMint)
	if err != nil {
		t.Fatal(err)
	}

	pool := pools[0]

	if pool.Account.CoinCreator.Equals(solana.PublicKey{}) {
		return
	}

	ix, err := ammService.TransferCreatorFeesToPumpInstruction(pool.Account.CoinCreator)
	if err != nil {
		t.Fatal(err)
	}

	sig, err := SendInstruction(context.Background(), rpcClient, wsClient, []solana.Instruction{ix}, partner.PublicKey(), func(key solana.PublicKey) *solana.PrivateKey {
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
