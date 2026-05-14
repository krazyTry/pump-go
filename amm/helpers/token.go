package helpers

import (
	"context"
	"fmt"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetTokenDecimals(ctx context.Context, client *rpc.Client, mint solana.PublicKey, tokenProgram solana.PublicKey) (uint8, error) {
	acc, err := client.GetAccountInfo(ctx, mint)
	if err != nil {
		return 0, err
	}
	if acc == nil || acc.Value == nil {
		return 0, fmt.Errorf("mint not found")
	}
	dec := bin.NewBinDecoder(acc.Value.Data.GetBinary())
	mintAcc := new(token.Mint)
	if err := mintAcc.UnmarshalWithDecoder(dec); err != nil {
		return 0, err
	}
	return mintAcc.Decimals, nil
}

// GetOrCreateATAInstruction returns the ATA pubkey and an optional create instruction if it doesn't exist.
func GetOrCreateATAInstruction(ctx context.Context, client *rpc.Client, tokenMint, owner, payer solana.PublicKey, tokenProgram solana.PublicKey) (solana.PublicKey, solana.Instruction, error) {
	ata := FindAssociatedTokenAddress(owner, tokenMint, tokenProgram)
	if _, err := client.GetAccountInfo(ctx, ata); err == nil {
		return ata, nil, nil
	}
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

func FindAssociatedTokenAddress(wallet, mint, tokenProgram solana.PublicKey) solana.PublicKey {
	ata, _, _ := solana.FindProgramAddress([][]byte{wallet.Bytes(), tokenProgram.Bytes(), mint.Bytes()}, solana.SPLAssociatedTokenAccountProgramID)
	return ata
}

func GetAccountInfo(ctx context.Context, client *rpc.Client, account solana.PublicKey) (*token.Account, error) {
	out, err := client.GetAccountInfo(ctx, account)
	if err != nil {
		return nil, err
	}

	acc := &token.Account{}
	if err := acc.UnmarshalWithDecoder(bin.NewBinDecoder(out.GetBinary())); err != nil {
		return nil, fmt.Errorf("unable to decode account: %w", err)
	}
	return acc, nil
}

func GetTokenInfo(ctx context.Context, client *rpc.Client, mint solana.PublicKey) (*TokenInfo, error) {

	out, err := client.GetAccountInfo(ctx, mint)
	if err != nil {
		return nil, err
	}

	mintAcc := &token.Mint{}

	if err := mintAcc.UnmarshalWithDecoder(bin.NewBinDecoder(out.GetBinary())); err != nil {
		return nil, fmt.Errorf("unable to decode mint: %w", err)
	}

	if !out.Value.Owner.Equals(solana.Token2022ProgramID) {
		return &TokenInfo{
			Owner: out.Value.Owner,
			Mint:  mintAcc,
		}, nil
	}

	ext, err := parseToken2022Extensions(out.GetBinary())
	if err != nil {
		return nil, err
	}

	if ext.TransferFeeConfig == nil {
		return &TokenInfo{
			Owner: out.Value.Owner,
			Mint:  mintAcc,
			Ext:   ext,
		}, nil
	}

	epochInfo, err := client.GetEpochInfo(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Owner: out.Value.Owner,
		Mint:  mintAcc,
		Ext:   ext,
		Fee:   ext.TransferFeeConfig.FeeForEpoch(epochInfo.Epoch),
	}, nil
}
