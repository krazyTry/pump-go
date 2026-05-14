package amm

import (
	"context"
	"fmt"
	"testing"

	token_metadata "github.com/928799934/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/krazyTry/pump-go/amm"
)

// 7kCd7jFy6MMKSnL5So96RSSarsBDiBrxtTy7semGoBgZ 5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1
// 4YwmadgZWofhxn1f2HNjyeDx5eKUwU5WVhxC2ZTPMfKM 5vXSDWQMecswf8pvupiFkCAeQX5ZNrxvLvNifWrnQtybDQ3dEmWiNLPnfaM52jLtAFxrR9EmG2dng8BNrDHSsn7q
// FfzvytqAzMFFg3av6Ht5HbaXwdqyogkZua1AvgZFLiRd 5gwR5Wzmhy3hqGdJv36f4D6ZRAz8i18XbyGgj4STkRMSt3cneV91a5emmNqVsfwESq2GXCXR4Adcr5LKQ5uu7Zk3
// Ep6iSRnmcP1uwYV4eQKEkfPTQQrpp6nHby2yU9G1dwbo 2u4atG3WQr5Y7jZ3x7LPqDet26NeRjPTrjKxdzdoMqjQn1rsMGsGhkUyhDDGWJZbDgdyzruvv3SouEJeXtPCigT3

func TestCreatePool(t *testing.T) {
	ammService := amm.NewClient(rpcClient, rpc.CommitmentFinalized)

	ownerWallet := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5Z9yeEAdetHyPVFuz4orcbDjeddo4VVrJ72CzvFBoseY6XM1zMdTC6wfQvgju4ntyHqgfGwUBptywzY2n1NFvFC1")}
	owner := ownerWallet.PublicKey()
	fmt.Println("owner address:", owner)

	partner := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("2u4atG3WQr5Y7jZ3x7LPqDet26NeRjPTrjKxdzdoMqjQn1rsMGsGhkUyhDDGWJZbDgdyzruvv3SouEJeXtPCigT3")}
	fmt.Println("partner address:", partner.PublicKey())

	mintWallet := solana.NewWallet()
	baseMint := mintWallet.PublicKey()

	fmt.Println("try to mintto token mint address:", baseMint, mintWallet)

	// SPL
	TokenMintto(t, context.TODO(), rpcClient, wsClient, mintWallet, ownerWallet, partner)
	baseTokenProgram := solana.TokenProgramID

	// 2022
	// Token2022MintTo(t, context.TODO(), rpcClient, wsClient, mintWallet, ownerWallet, partner)
	// baseTokenProgram := solana.Token2022ProgramID

	name := "PumpGoTest"
	symbol := "PUMPGOTEST"
	uri := "https://pump.fun/logo.png"

	{
		metadataPDA, _, _ := solana.FindProgramAddress([][]byte{[]byte("metadata"), solana.TokenMetadataProgramID.Bytes(), baseMint.Bytes()}, solana.TokenMetadataProgramID)
		createMetadataIx := token_metadata.NewCreateMetadataAccountV3Instruction(
			token_metadata.CreateMetadataAccountArgsV3{
				Data: token_metadata.DataV2{
					Name:                 name,
					Symbol:               symbol,
					Uri:                  uri,
					SellerFeeBasisPoints: 0,
				},
				IsMutable: true,
			},
			metadataPDA,
			baseMint,
			ownerWallet.PublicKey(),
			ownerWallet.PublicKey(),
			ownerWallet.PublicKey(),
			solana.SystemProgramID,
			solana.SysVarRentPubkey,
		).Build()

		sig, err := SendInstruction(context.Background(), rpcClient, wsClient, []solana.Instruction{createMetadataIx}, ownerWallet.PublicKey(), func(key solana.PublicKey) *solana.PrivateKey {
			switch {
			case key.Equals(ownerWallet.PublicKey()):
				return &ownerWallet.PrivateKey
			default:
				return nil
			}
		})

		if err != nil {
			t.Fatal("create metadata SendTransaction() fail", err)
		}

		fmt.Println("create metadata success Success sig:", sig.String())
	}

	quoteMint := solana.WrappedSol
	quoteTokenProgram := solana.TokenProgramID

	fmt.Println("try to create pool")

	createIxs, pre, post, err := ammService.CreatePoolInstructions(
		context.Background(),
		&amm.CreatePoolParams{
			Index:        1,
			BaseIn:       10_000_000 * 1e6,
			QuoteIn:      0.01 * 1e9,
			IsMayhemMode: false,
			IsCashback:   false,
		},
		partner.PublicKey(),
		baseMint,
		quoteMint,
		baseTokenProgram,
		quoteTokenProgram,
	)
	if err != nil {
		t.Fatal(err)
	}

	var instructions []solana.Instruction
	instructions = append(instructions, pre...)
	instructions = append(instructions, createIxs...)
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
		t.Fatal("create pool SendTransaction() fail", err)
	}
	fmt.Println("create pool success Success sig:", sig.String())
}
