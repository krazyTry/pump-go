package meteora

import (
	"crypto/rand"
	"fmt"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
)

func TestToken(t *testing.T) {
	// ss := solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5zx2kW6MBvySqGuYyLC7K1VBfLqFE1VKbT847hcRS33Meid2DFjc1SSYHJxMt15ZMDDbXhTMN9fJRNnMYPYDCsUz")}
	// fmt.Println(ss.PublicKey().String())

	// return

	s := solana.NewWallet()
	fmt.Println(s.PublicKey().String(), s.PrivateKey.String())
	s1 := solana.NewWallet()
	fmt.Println(s1.PublicKey().String(), s1.PrivateKey.String())
	s2 := solana.NewWallet()
	fmt.Println(s2.PublicKey().String(), s2.PrivateKey.String())
	s3 := solana.NewWallet()
	fmt.Println(s3.PublicKey().String(), s3.PrivateKey.String())

	return

	programID := solana.MustPublicKeyFromBase58("dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN")

	targetSuffix := "ccaw"

	for {
		seed := randomSeed(8)

		for bump := 255; bump >= 0; bump-- {
			seeds := [][]byte{
				[]byte("pool"),
				seed,
				{byte(bump)},
			}

			pda, err := solana.CreateProgramAddress(seeds, programID)
			// pda, _, err := solana.FindProgramAddress(
			// 	[][]byte{[]byte("pool"), seed},
			// 	programID,
			// )
			if err != nil {
				continue
			}

			addr := pda.String()
			if strings.HasSuffix(addr, targetSuffix) {
				fmt.Println("FOUND!")
				fmt.Println("PDA:", addr)
				fmt.Println("Seed:", seed)
				return
			}
		}
	}

	// pub, _, _ := solana.FindProgramAddress([][]byte{[]byte("po12ol"), []byte("pool")}, programID)
	// fmt.Println(pub.String())
}

func randomSeed(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}
