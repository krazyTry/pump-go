package bonding_curve

import (
	"context"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/krazyTry/pump-go/bonding_curve"
	"github.com/krazyTry/pump-go/bonding_curve/helpers"
)

func TestSwap(t *testing.T) {
	pumpService := bonding_curve.NewClient(rpcClient, rpc.CommitmentFinalized)

	// 4YwmadgZWofhxn1f2HNjyeDx5eKUwU5WVhxC2ZTPMfKM 5vXSDWQMecswf8pvupiFkCAeQX5ZNrxvLvNifWrnQtybDQ3dEmWiNLPnfaM52jLtAFxrR9EmG2dng8BNrDHSsn7q
	partner := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5vXSDWQMecswf8pvupiFkCAeQX5ZNrxvLvNifWrnQtybDQ3dEmWiNLPnfaM52jLtAFxrR9EmG2dng8BNrDHSsn7q")}
	fmt.Println("partner address:", partner.PublicKey())

	// SPL
	// ogSCRqsqUX2EAqiym8d1Em8pYXXg5M3QK9BL5j6GFhD &{5craheZijH3mFYSfsrDUPzxfv5zo4oKfJGv856LHugXg2JkN7FJYYgPM2yefWYfFFcr44htSPm7yw9y6fw3Ka7ab}

	// 2022
	// Q6yj3JguAAz4ta55bg2uLBwyPr8K4Zzme79VtY1vmRh &{4afHFZdhQ4kwstJ8NfgZXrXbNgQPsb429FPyAvaiUK1uFZ1iCv29kgk297gFmuHHQpaDjmetMG9n929vYAux5MJw}

	baseMint := solana.MustPublicKeyFromBase58("ogSCRqsqUX2EAqiym8d1Em8pYXXg5M3QK9BL5j6GFhD")
	baseTokenProgram := solana.TokenProgramID

	// baseMint := solana.MustPublicKeyFromBase58("Q6yj3JguAAz4ta55bg2uLBwyPr8K4Zzme79VtY1vmRh")
	// baseTokenProgram := solana.Token2022ProgramID

	global, _ := pumpService.FetchGlobal(context.TODO())
	feeConfig, _ := pumpService.FetchFeeConfig(context.TODO())

	bondingCurve, _ := pumpService.FetchBondingCurve(context.Background(), baseMint)

	// quoteMint := bondingCurve.QuoteMint
	// quoteTokenProgram := solana.TokenProgramID
	// if bondingCurve.QuoteMint != solana.WrappedSol {
	// 	quoteTokenProgram, _ = helpers.GetTokenProgram(context.Background(), rpcClient, quoteMint)
	// }

	{
		amountIn := uint64(10 * 1e6)

		tokenAmount, err := bonding_curve.GetBuyTokenAmountFromSolAmount(
			global,
			feeConfig,
			bondingCurve.TokenTotalSupply,
			bondingCurve,
			amountIn,
		)
		if err != nil {
			t.Fatal(err)
		}

		solAmount, err := bonding_curve.GetBuySolAmountFromTokenAmount(
			global,
			feeConfig,
			bondingCurve.TokenTotalSupply,
			bondingCurve,
			tokenAmount,
		)
		if err != nil {
			t.Fatal(err)
		}

		buyIx, pre, post, err := pumpService.BuyInstructions(context.Background(), global, bondingCurve, partner.PublicKey(), baseMint, baseTokenProgram, amountIn, solAmount, 10)
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
		solAmount, err := bonding_curve.GetSellSolAmountFromTokenAmount(
			global,
			feeConfig,
			bondingCurve.TokenTotalSupply,
			bondingCurve,
			amountIn,
		)
		if err != nil {
			t.Fatal(err)
		}

		sellIx, pre, post, err := pumpService.SellInstructions(context.Background(), global, bondingCurve, partner.PublicKey(), baseMint, baseTokenProgram, amountIn, solAmount, 10)
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

func TestSwapV2(t *testing.T) {
	pumpService := bonding_curve.NewClient(rpcClient, rpc.CommitmentFinalized)

	// 4YwmadgZWofhxn1f2HNjyeDx5eKUwU5WVhxC2ZTPMfKM 5vXSDWQMecswf8pvupiFkCAeQX5ZNrxvLvNifWrnQtybDQ3dEmWiNLPnfaM52jLtAFxrR9EmG2dng8BNrDHSsn7q
	partner := &solana.Wallet{PrivateKey: solana.MustPrivateKeyFromBase58("5vXSDWQMecswf8pvupiFkCAeQX5ZNrxvLvNifWrnQtybDQ3dEmWiNLPnfaM52jLtAFxrR9EmG2dng8BNrDHSsn7q")}
	fmt.Println("partner address:", partner.PublicKey())

	// SPL
	// ogSCRqsqUX2EAqiym8d1Em8pYXXg5M3QK9BL5j6GFhD &{5craheZijH3mFYSfsrDUPzxfv5zo4oKfJGv856LHugXg2JkN7FJYYgPM2yefWYfFFcr44htSPm7yw9y6fw3Ka7ab}

	// 2022
	// Q6yj3JguAAz4ta55bg2uLBwyPr8K4Zzme79VtY1vmRh &{4afHFZdhQ4kwstJ8NfgZXrXbNgQPsb429FPyAvaiUK1uFZ1iCv29kgk297gFmuHHQpaDjmetMG9n929vYAux5MJw}

	// baseMint := solana.MustPublicKeyFromBase58("ogSCRqsqUX2EAqiym8d1Em8pYXXg5M3QK9BL5j6GFhD")
	// baseTokenProgram := solana.TokenProgramID

	baseMint := solana.MustPublicKeyFromBase58("Q6yj3JguAAz4ta55bg2uLBwyPr8K4Zzme79VtY1vmRh")
	baseTokenProgram := solana.Token2022ProgramID

	global, _ := pumpService.FetchGlobal(context.TODO())
	feeConfig, _ := pumpService.FetchFeeConfig(context.TODO())

	bondingCurve, _ := pumpService.FetchBondingCurve(context.Background(), baseMint)

	quoteMint := bondingCurve.QuoteMint
	quoteTokenProgram := solana.TokenProgramID
	if bondingCurve.QuoteMint != solana.WrappedSol {
		quoteTokenProgram, _ = helpers.GetTokenProgram(context.Background(), rpcClient, quoteMint)
	}

	{
		amountIn := uint64(10 * 1e6)

		tokenAmount, err := bonding_curve.GetBuyTokenAmountFromSolAmount(
			global,
			feeConfig,
			bondingCurve.TokenTotalSupply,
			bondingCurve,
			amountIn,
		)
		if err != nil {
			t.Fatal(err)
		}

		solAmount, err := bonding_curve.GetBuySolAmountFromTokenAmount(
			global,
			feeConfig,
			bondingCurve.TokenTotalSupply,
			bondingCurve,
			tokenAmount,
		)
		if err != nil {
			t.Fatal(err)
		}

		buyIx, pre, post, err := pumpService.BuyV2Instructions(context.Background(), global, bondingCurve, partner.PublicKey(), baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, amountIn, solAmount, 10)
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
		solAmount, err := bonding_curve.GetSellSolAmountFromTokenAmount(
			global,
			feeConfig,
			bondingCurve.TokenTotalSupply,
			bondingCurve,
			amountIn,
		)
		if err != nil {
			t.Fatal(err)
		}

		sellIx, pre, post, err := pumpService.SellV2Instructions(context.Background(), global, bondingCurve, partner.PublicKey(), baseMint, quoteMint, baseTokenProgram, quoteTokenProgram, amountIn, solAmount, 10)
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
