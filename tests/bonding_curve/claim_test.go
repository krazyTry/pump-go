package bonding_curve

import (
	"context"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/krazyTry/pump-go/bonding_curve"
)

func TestClaim(t *testing.T) {
	pumpService := bonding_curve.NewClient(rpcClient, rpc.CommitmentFinalized)

	// 7kCd7jFy6MMKSnL5So96RSSarsBDiBrxtTy7semGoBgZ 5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1
	ownerWallet := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1")}
	owner := ownerWallet.PublicKey()
	fmt.Println("owner address:", owner)

	// SPL
	// ogSCRqsqUX2EAqiym8d1Em8pYXXg5M3QK9BL5j6GFhD &{5craheZijH3mFYSfsrDUPzxfv5zo4oKfJGv856LHugXg2JkN7FJYYgPM2yefWYfFFcr44htSPm7yw9y6fw3Ka7ab}

	// 2022
	// Q6yj3JguAAz4ta55bg2uLBwyPr8K4Zzme79VtY1vmRh &{4afHFZdhQ4kwstJ8NfgZXrXbNgQPsb429FPyAvaiUK1uFZ1iCv29kgk297gFmuHHQpaDjmetMG9n929vYAux5MJw}

	// baseMint := solana.MustPublicKeyFromBase58("ogSCRqsqUX2EAqiym8d1Em8pYXXg5M3QK9BL5j6GFhD")

	creatorVault := bonding_curve.DeriveCreatorVault(owner)

	balance, err := Balance(context.TODO(), rpcClient, creatorVault)
	if err != nil {
		t.Fatal(err)
	}

	if balance <= 890880 {
		return
	}

	fee := balance - 890880 // 890880 is Account deposit
	fmt.Println("fee:", fee)

	claimIx, err := pumpService.CollectCreatorFeeInstruction(owner)
	if err != nil {
		t.Fatal(err)
	}

	var instructions []solana.Instruction
	instructions = append(instructions, claimIx)

	sig, err := SendInstruction(context.Background(), rpcClient, wsClient, instructions, ownerWallet.PublicKey(), func(key solana.PublicKey) *solana.PrivateKey {
		switch {
		case key.Equals(ownerWallet.PublicKey()):
			return &ownerWallet.PrivateKey
		default:
			return nil
		}
	})
	if err != nil {
		t.Fatal("claim SendTransaction() fail", err)
	}
	fmt.Println("claim success Success sig:", sig.String())
}
