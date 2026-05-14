package bonding_curve

import (
	"context"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/krazyTry/pump-go/bonding_curve"
)

// 7kCd7jFy6MMKSnL5So96RSSarsBDiBrxtTy7semGoBgZ 5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1
// 4YwmadgZWofhxn1f2HNjyeDx5eKUwU5WVhxC2ZTPMfKM 5vXSDWQMecswf8pvupiFkCAeQX5ZNrxvLvNifWrnQtybDQ3dEmWiNLPnfaM52jLtAFxrR9EmG2dng8BNrDHSsn7q
// FfzvytqAzMFFg3av6Ht5HbaXwdqyogkZua1AvgZFLiRd 5gwR5Wzmhy3hqGdJv36f4D6ZRAz8i18XbyGgj4STkRMSt3cneV91a5emmNqVsfwESq2GXCXR4Adcr5LKQ5uu7Zk3
// Ep6iSRnmcP1uwYV4eQKEkfPTQQrpp6nHby2yU9G1dwbo 2u4atG3WQr5Y7jZ3x7LPqDet26NeRjPTrjKxdzdoMqjQn1rsMGsGhkUyhDDGWJZbDgdyzruvv3SouEJeXtPCigT3

func TestCreatePool(t *testing.T) {
	pumpService := bonding_curve.NewClient(rpcClient, rpc.CommitmentFinalized)

	name := "PumpGoTest"
	symbol := "PUMPGOTEST"
	uri := "https://pump.fun/logo.png"

	ownerWallet := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1")}
	owner := ownerWallet.PublicKey()
	fmt.Println("owner address:", owner)

	mintWallet := solana.NewWallet()
	// ogSCRqsqUX2EAqiym8d1Em8pYXXg5M3QK9BL5j6GFhD &{5craheZijH3mFYSfsrDUPzxfv5zo4oKfJGv856LHugXg2JkN7FJYYgPM2yefWYfFFcr44htSPm7yw9y6fw3Ka7ab}
	baseMint := mintWallet.PublicKey()

	fmt.Println("try to create token mint address:", baseMint, mintWallet)

	creator := owner

	createIx, err := pumpService.CreateInstruction(
		creator,
		owner,
		baseMint,
		name,
		symbol,
		uri,
	)
	if err != nil {
		t.Fatal(err)
	}

	instructions := []solana.Instruction{createIx}
	sig, err := SendInstruction(context.Background(), rpcClient, wsClient, instructions, ownerWallet.PublicKey(), func(key solana.PublicKey) *solana.PrivateKey {
		switch {
		case key.Equals(mintWallet.PublicKey()):
			return &mintWallet.PrivateKey
		case key.Equals(ownerWallet.PublicKey()):
			return &ownerWallet.PrivateKey
		default:
			return nil
		}
	})
	if err != nil {
		t.Fatal("create SendTransaction() fail", err)
	}
	fmt.Println("create success Success sig:", sig.String())
}

func TestCreateV2Pool(t *testing.T) {
	pumpService := bonding_curve.NewClient(rpcClient, rpc.CommitmentFinalized)

	name := "PumpGoTest"
	symbol := "PUMPGOTEST"
	uri := "https://pump.fun/logo.png"

	ownerWallet := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1")}
	owner := ownerWallet.PublicKey()
	fmt.Println("owner address:", owner)

	mintWallet := solana.NewWallet()
	// Q6yj3JguAAz4ta55bg2uLBwyPr8K4Zzme79VtY1vmRh &{4afHFZdhQ4kwstJ8NfgZXrXbNgQPsb429FPyAvaiUK1uFZ1iCv29kgk297gFmuHHQpaDjmetMG9n929vYAux5MJw}

	baseMint := mintWallet.PublicKey()

	fmt.Println("try to createV2 token mint address:", baseMint, mintWallet)

	creator := owner

	createIx, err := pumpService.CreateV2Instruction(
		creator,
		owner,
		baseMint,
		solana.WrappedSol,
		solana.TokenProgramID,
		name,
		symbol,
		uri,
		false,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	instructions := []solana.Instruction{createIx}
	sig, err := SendInstruction(context.Background(), rpcClient, wsClient, instructions, ownerWallet.PublicKey(), func(key solana.PublicKey) *solana.PrivateKey {
		switch {
		case key.Equals(mintWallet.PublicKey()):
			return &mintWallet.PrivateKey
		case key.Equals(ownerWallet.PublicKey()):
			return &ownerWallet.PrivateKey
		default:
			return nil
		}
	})
	if err != nil {
		t.Fatal("createV2 SendTransaction() fail", err)
	}
	fmt.Println("createV2 success Success sig:", sig.String())
}
