package helpers

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

func FindAssociatedTokenAddress(owner, quoteMint, quoteTokenProgram solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{owner.Bytes(), quoteTokenProgram.Bytes(), quoteMint.Bytes()}, solana.SPLAssociatedTokenAccountProgramID)
	return key
}

// GetOrCreateATAInstruction returns the ATA pubkey and an optional create instruction if it doesn't exist.
func GetOrCreateATAInstruction(ctx context.Context, client *rpc.Client, tokenMint, owner, payer solana.PublicKey, tokenProgram solana.PublicKey) (solana.PublicKey, solana.Instruction, error) {
	ata := FindAssociatedTokenAddress(owner, tokenMint, tokenProgram)

	if _, err := client.GetAccountInfo(ctx, ata); err == nil {
		return ata, nil, nil
	}

	// create if missing
	ix := CreateAssociatedTokenAccountInstruction(payer, ata, owner, tokenMint, tokenProgram)

	return ata, ix, nil
}

// CreateAssociatedTokenAccountInstruction builds an ATA create instruction that supports custom token programs (SPL/Token2022).
func CreateAssociatedTokenAccountInstruction(payer, ata, owner, mint, tokenProgram solana.PublicKey) solana.Instruction {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(payer, true, true),
		solana.NewAccountMeta(ata, true, false),
		solana.NewAccountMeta(owner, false, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(system.ProgramID, false, false),
		solana.NewAccountMeta(tokenProgram, false, false),
	}
	return solana.NewInstruction(solana.SPLAssociatedTokenAccountProgramID, accounts, nil)
}

func UnwrapSOLInstruction(owner, receiver solana.PublicKey, allowOwnerOffCurve bool) (solana.Instruction, error) {
	ata := FindAssociatedTokenAddress(owner, solana.WrappedSol, token.ProgramID)

	return token.NewCloseAccountInstructionBuilder().
		SetAccount(ata).
		SetDestinationAccount(receiver).
		SetOwnerAccount(owner).
		Build(), nil
}

func WrapSOLInstruction(from, to solana.PublicKey, amount uint64) ([]solana.Instruction, error) {
	transferIx := system.NewTransferInstructionBuilder().
		SetFundingAccount(from).
		SetRecipientAccount(to).
		SetLamports(amount).
		Build()

	syncIx := token.NewSyncNativeInstructionBuilder().
		SetTokenAccount(to).
		Build()

	return []solana.Instruction{transferIx, syncIx}, nil
}

func GetTokenProgram(ctx context.Context, client *rpc.Client, tokenMint solana.PublicKey) (solana.PublicKey, error) {
	acc, err := client.GetAccountInfo(ctx, tokenMint)
	if err != nil {
		return solana.PublicKey{}, err
	}
	if acc == nil || acc.Value == nil {
		return solana.PublicKey{}, fmt.Errorf("mint not found")
	}
	owner := acc.Value.Owner
	if owner.Equals(token.ProgramID) {
		return solana.TokenProgramID, nil
	}
	return solana.Token2022ProgramID, nil
}
